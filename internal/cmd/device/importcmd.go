package device

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type ImportOptions struct {
	Yes    bool
	NoWait bool
}

func NewCmdImport(f *factory.Factory) *cobra.Command {
	opts := &ImportOptions{}

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Bulk import devices from a file",
		Long: `Bulk import devices from a CSV or Excel (.xlsx) file.

The file must contain a header row with the following columns:
  name          Device name (required)
  serialNumber  Serial number (required, alphanumeric only)
  mac           MAC address (optional, e.g. AA:BB:CC:DD:EE:FF)
  imei          IMEI number (optional)

CSV files are automatically converted to Excel format before upload.

The import is asynchronous: the file is uploaded and validated, then you
confirm the import, and it runs in the background. By default the command
waits for the import to complete and displays the result.`,
		Example: `  # Import devices from a CSV file
  incloud device import devices.csv

  # Import from Excel
  incloud device import devices.xlsx

  # Skip confirmation prompt
  incloud device import devices.csv -y

  # Don't wait for completion
  incloud device import devices.csv -y --no-wait`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return runImport(f, opts, filePath)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&opts.NoWait, "no-wait", false, "Don't wait for import to complete")

	return cmd
}

func runImport(f *factory.Factory, opts *ImportOptions, filePath string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	ctx, err := cfg.ActiveContext()
	if err != nil {
		return err
	}
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Determine file type and prepare XLSX for upload
	ext := strings.ToLower(filepath.Ext(filePath))
	uploadPath := filePath
	var tmpFile string

	switch ext {
	case ".csv":
		// Convert CSV to XLSX
		xlsxPath, err := csvToXLSX(filePath)
		if err != nil {
			return fmt.Errorf("converting CSV to Excel: %w", err)
		}
		uploadPath = xlsxPath
		tmpFile = xlsxPath
		defer func() { _ = os.Remove(tmpFile) }()
	case ".xlsx", ".xls":
		// Use as-is
	default:
		return fmt.Errorf("unsupported file format %q (expected .csv, .xlsx, or .xls)", ext)
	}

	// Step 1: Upload file
	fmt.Fprintf(f.IO.ErrOut, "Uploading %s...\n", filepath.Base(filePath))
	jobID, err := uploadImportFile(client, ctx.Host, uploadPath)
	if err != nil {
		return err
	}

	// Step 2: Wait for file validation to complete, then show parsed result
	job, err := waitForValidation(client, ctx.Host, jobID)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "Parsed %d device(s) from %s (job: %s)\n", job.Total, job.FileName, jobID)

	// If validation failed, show errors and abort
	if job.Status == "check_fail" {
		fmt.Fprintf(f.IO.ErrOut, "\nValidation failed:\n")
		for errCode, rows := range job.Result {
			fmt.Fprintf(f.IO.ErrOut, "  %s: rows %v\n", errCode, rows)
		}
		return fmt.Errorf("file validation failed, import aborted")
	}

	// Show validation warnings if any
	if len(job.Result) > 0 {
		fmt.Fprintf(f.IO.ErrOut, "\nValidation warnings:\n")
		for errCode, rows := range job.Result {
			fmt.Fprintf(f.IO.ErrOut, "  %s: rows %v\n", errCode, rows)
		}
		fmt.Fprintln(f.IO.ErrOut)
	}

	// Step 3: Confirm
	if !opts.Yes {
		if file, ok := f.IO.In.(*os.File); !ok || !isatty.IsTerminal(file.Fd()) {
			return fmt.Errorf("terminal is non-interactive; use --yes to confirm")
		}
		fmt.Fprintf(f.IO.ErrOut, "Proceed with importing %d device(s)? (y/N) ", job.Total)
		reader := bufio.NewReader(f.IO.In)
		answer, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if a := strings.TrimSpace(answer); a != "y" && a != "Y" {
			fmt.Fprintln(f.IO.ErrOut, "Aborted.")
			return nil
		}
	}

	// Step 4: Confirm import
	fmt.Fprintf(f.IO.ErrOut, "Starting import...\n")
	if err := confirmImport(client, ctx.Host, jobID); err != nil {
		return err
	}

	if opts.NoWait {
		fmt.Fprintf(f.IO.ErrOut, "Import job %s started. Use 'incloud api get /api/v1/devices/imports/%s/detail' to check status.\n", jobID, jobID)
		return nil
	}

	// Step 5: Poll for completion
	job, err = pollImportJob(f, client, ctx.Host, jobID)
	if err != nil {
		return err
	}

	// Step 6: Show final result
	return showImportResult(f, job)
}

// csvToXLSX converts a CSV file to an XLSX file and returns the temporary XLSX path.
func csvToXLSX(csvPath string) (string, error) {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return "", fmt.Errorf("opening CSV file: %w", err)
	}
	defer func() { _ = csvFile.Close() }()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("reading CSV: %w", err)
	}

	if len(records) < 2 {
		return "", fmt.Errorf("CSV file must have a header row and at least one data row")
	}

	xlsx := excelize.NewFile()
	defer func() { _ = xlsx.Close() }()

	sheet := "Sheet1"
	for i, row := range records {
		for j, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				return "", fmt.Errorf("computing cell name: %w", err)
			}
			if err := xlsx.SetCellValue(sheet, cellName, cell); err != nil {
				return "", fmt.Errorf("writing cell %s: %w", cellName, err)
			}
		}
	}

	tmpFile, err := os.CreateTemp("", "incloud-import-*.xlsx")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	if err := xlsx.SaveAs(tmpPath); err != nil {
		_ = os.Remove(tmpPath) //nolint:gosec // tmpPath is from os.CreateTemp, not user input
		return "", fmt.Errorf("saving Excel file: %w", err)
	}

	return tmpPath, nil
}

// uploadImportFile uploads the file via multipart POST and returns the job ID.
func uploadImportFile(client *http.Client, host, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(writer.Close())
	}()

	reqURL := host + "/api/v1/devices/imports"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, pr)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("upload failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return result.Result, nil
}

type importJob struct {
	ID        string           `json:"_id"`
	FileName  string           `json:"fileName"`
	FileSize  int64            `json:"fileSize"`
	Total     int              `json:"total"`
	SuccessNo int              `json:"successNo"`
	FailNo    int              `json:"failNo"`
	Status    string           `json:"status"`
	Result    map[string][]int `json:"result,omitempty"`
	Rate      float64          `json:"rate"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
	UserName  string           `json:"userName,omitempty"`
}

func getImportJobDetail(client *http.Client, host, jobID string) (*importJob, error) {
	reqURL := host + "/api/v1/devices/imports/" + jobID + "/detail"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result importJob `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result.Result, nil
}

func confirmImport(client *http.Client, host, jobID string) error {
	reqURL := host + "/api/v1/devices/imports/" + jobID
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("confirm failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("confirm failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// isTerminalStatus returns true if the import job has reached a final state.
func isTerminalStatus(status string) bool {
	switch status {
	case "success", "failed", "check_fail", "cancel":
		return true
	}
	return false
}

// waitForValidation polls until the job exits the "checking" state.
// The upload handler parses the file synchronously, so "init" means ready.
// Only "checking" requires waiting (rare, for large files with async validation).
func waitForValidation(client *http.Client, host, jobID string) (*importJob, error) {
	timeout := time.After(2 * time.Minute)

	for {
		job, err := getImportJobDetail(client, host, jobID)
		if err != nil {
			return nil, err
		}

		if job.Status != "checking" {
			return job, nil
		}

		select {
		case <-time.After(1 * time.Second):
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for file validation (job: %s)", jobID)
		}
	}
}

func pollImportJob(f *factory.Factory, client *http.Client, host, jobID string) (*importJob, error) {
	timeout := time.After(10 * time.Minute)

	for {
		job, err := getImportJobDetail(client, host, jobID)
		if err != nil {
			return nil, err
		}

		if isTerminalStatus(job.Status) {
			return job, nil
		}

		// Show progress
		if job.Rate > 0 && job.Rate < 1 {
			fmt.Fprintf(f.IO.ErrOut, "\rImporting... %.0f%%", job.Rate*100)
		}

		select {
		case <-time.After(2 * time.Second):
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for import to complete (job: %s)", jobID)
		}
	}
}

func showImportResult(f *factory.Factory, job *importJob) error {
	switch job.Status {
	case "success":
		fmt.Fprintf(f.IO.ErrOut, "Import completed: %d/%d device(s) imported successfully.\n", job.SuccessNo, job.Total)
	case "failed":
		fmt.Fprintf(f.IO.ErrOut, "Import finished with errors: %d succeeded, %d failed out of %d.\n", job.SuccessNo, job.FailNo, job.Total)
	case "check_fail":
		fmt.Fprintf(f.IO.ErrOut, "Import validation failed.\n")
	case "cancel":
		fmt.Fprintf(f.IO.ErrOut, "Import was cancelled.\n")
	default:
		fmt.Fprintf(f.IO.ErrOut, "Import ended with status: %s\n", job.Status)
	}

	if len(job.Result) > 0 {
		fmt.Fprintf(f.IO.ErrOut, "\nErrors:\n")
		for errCode, rows := range job.Result {
			fmt.Fprintf(f.IO.ErrOut, "  %s: rows %v\n", errCode, rows)
		}
	}

	if job.Status != "success" && job.FailNo > 0 {
		return fmt.Errorf("import completed with %d failure(s)", job.FailNo)
	}
	if job.Status == "check_fail" || job.Status == "cancel" {
		return fmt.Errorf("import %s", job.Status)
	}

	return nil
}
