package device

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

type ImportOptions struct {
	Yes     bool
	NoWait  bool
	GroupID string
	OrgID   string
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

  # Import and assign devices to a group
  incloud device import devices.csv --group 507f1f77bcf86cd799439011

  # Import devices under a sub-organization
  incloud device import devices.csv --org 507f1f77bcf86cd799439022

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
	cmd.Flags().StringVar(&opts.GroupID, "group", "", "Assign imported devices to a group (use 'incloud device group list' to find IDs)")
	cmd.Flags().StringVar(&opts.OrgID, "org", "", "Create devices under a sub-organization (use 'incloud org list' to find IDs)")

	return cmd
}

func runImport(f *factory.Factory, opts *ImportOptions, filePath string) error {
	client, err := f.APIClient()
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
	jobID, err := uploadImportFile(client, uploadPath, opts.GroupID, opts.OrgID)
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IO.ErrOut, "Import job: %s\n", jobID)

	// Step 2: Wait for file validation to complete, then show parsed result
	job, err := waitForValidation(client, jobID)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IO.ErrOut, "Parsed %d device(s) from %s\n", job.Total, job.FileName)

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
		confirmed, err := ui.Confirm(f, fmt.Sprintf("Proceed with importing %d device(s)?", job.Total))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	// Step 4: Confirm import
	fmt.Fprintf(f.IO.ErrOut, "Starting import...\n")
	if err := confirmImport(client, jobID); err != nil {
		return err
	}

	if opts.NoWait {
		fmt.Fprintf(f.IO.ErrOut, "Import started. Track progress with: incloud device import-status %s\n", jobID)
		return nil
	}

	// Step 5: Poll for completion
	job, err = pollImportJob(f, client, jobID)
	if err != nil {
		return err
	}

	// Step 6: Show final result
	return showImportResult(f, client, job)
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
// groupID and orgID are optional; when non-empty they are sent as multipart form
// fields so the backend assigns imported devices to the given group / sub-org.
func uploadImportFile(client *api.APIClient, filePath, groupID, orgID string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fields := make(map[string]string)
	if groupID != "" {
		fields["groupId"] = groupID
	}
	if orgID != "" {
		fields["oid"] = orgID
	}

	var body []byte
	if len(fields) > 0 {
		body, err = client.UploadWithFields("/api/v1/devices/imports", "file", filepath.Base(filePath), file, fields)
	} else {
		body, err = client.Upload("/api/v1/devices/imports", "file", filepath.Base(filePath), file)
	}
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
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

func getImportJobDetail(client *api.APIClient, jobID string) (*importJob, error) {
	body, err := client.Get("/api/v1/devices/imports/"+jobID+"/detail", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result importJob `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result.Result, nil
}

func confirmImport(client *api.APIClient, jobID string) error {
	_, err := client.Post("/api/v1/devices/imports/"+jobID, nil)
	return err
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
func waitForValidation(client *api.APIClient, jobID string) (*importJob, error) {
	timeout := time.After(2 * time.Minute)

	for {
		job, err := getImportJobDetail(client, jobID)
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

func pollImportJob(f *factory.Factory, client *api.APIClient, jobID string) (*importJob, error) {
	timeout := time.After(10 * time.Minute)

	for {
		job, err := getImportJobDetail(client, jobID)
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

func showImportResult(f *factory.Factory, client *api.APIClient, job *importJob) error {
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

	// Show per-row failure details from the details API;
	// fall back to the job-level error codes if unavailable.
	if client != nil && job.FailNo > 0 {
		if !showFailedDetails(f, client, job.ID) {
			showJobErrors(f, job)
		}
	} else {
		showJobErrors(f, job)
	}

	if job.Status != "success" && job.FailNo > 0 {
		return fmt.Errorf("import completed with %d failure(s)", job.FailNo)
	}
	if job.Status == "check_fail" || job.Status == "cancel" {
		return fmt.Errorf("import %s", job.Status)
	}

	return nil
}

// importDetail represents a single row in the import detail list.
type importDetail struct {
	Row          int    `json:"row"`
	DeviceName   string `json:"deviceName"`
	SerialNumber string `json:"serialNumber"`
	Status       string `json:"status"`
	FailReason   string `json:"failReason"`
}

// getFailedDetails fetches import details with FAILED status for a given job.
func getFailedDetails(client *api.APIClient, jobID string) ([]importDetail, error) {
	q := make(url.Values)
	q.Set("status", "FAILED")
	q.Set("page", "0")
	q.Set("size", "100")

	body, err := client.Get("/api/v1/devices/imports/"+jobID+"/details", q)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result []importDetail `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing details response: %w", err)
	}

	return result.Result, nil
}

// showFailedDetails prints per-row failure reasons to stderr.
// Returns true if details were successfully fetched and displayed.
func showFailedDetails(f *factory.Factory, client *api.APIClient, jobID string) bool {
	details, err := getFailedDetails(client, jobID)
	if err != nil || len(details) == 0 {
		return false
	}

	fmt.Fprintf(f.IO.ErrOut, "\nFailed rows:\n")
	for _, d := range details {
		fmt.Fprintf(f.IO.ErrOut, "  Row %d: %s (%s) — %s\n", d.Row, d.SerialNumber, d.DeviceName, d.FailReason)
	}
	return true
}

// showJobErrors prints job-level error codes (fallback when details API is unavailable).
func showJobErrors(f *factory.Factory, job *importJob) {
	if len(job.Result) == 0 {
		return
	}
	fmt.Fprintf(f.IO.ErrOut, "\nErrors:\n")
	for errCode, rows := range job.Result {
		fmt.Fprintf(f.IO.ErrOut, "  %s: rows %v\n", errCode, rows)
	}
}
