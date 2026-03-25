package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ImportStatusOptions struct {
	Wait bool
}

func NewCmdImportStatus(f *factory.Factory) *cobra.Command {
	opts := &ImportStatusOptions{}

	cmd := &cobra.Command{
		Use:   "import-status <job-id>",
		Short: "Check the status of a device import job",
		Long: `Check the status of a device import job by its ID.

If the job is still running, use --wait to poll until completion.
Failed rows are shown with their serial numbers and failure reasons.`,
		Example: `  # Check import status
  incloud device import-status 69c371131e0e4d15c8cb2b25

  # Wait for a running import to complete
  incloud device import-status 69c371131e0e4d15c8cb2b25 --wait

  # Get structured output
  incloud device import-status 69c371131e0e4d15c8cb2b25 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportStatus(cmd, f, opts, args[0])
		},
	}

	cmd.Flags().BoolVarP(&opts.Wait, "wait", "w", false, "Wait for the import to complete if still running")

	return cmd
}

// importStatusResult is the structured output for import-status.
type importStatusResult struct {
	ID        string         `json:"_id"`
	FileName  string         `json:"fileName"`
	Status    string         `json:"status"`
	Total     int            `json:"total"`
	SuccessNo int            `json:"successNo"`
	FailNo    int            `json:"failNo"`
	Rate      float64        `json:"rate"`
	CreatedAt string         `json:"createdAt"`
	Failed    []importDetail `json:"failed,omitempty"`
}

func runImportStatus(cmd *cobra.Command, f *factory.Factory, opts *ImportStatusOptions, jobID string) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	job, err := getImportJobDetail(client, jobID)
	if err != nil {
		return err
	}

	if !isTerminalStatus(job.Status) {
		if opts.Wait {
			fmt.Fprintf(f.IO.ErrOut, "Import job %s is %s, waiting for completion...\n", jobID, job.Status)
			job, err = pollImportJob(f, client, jobID)
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintf(f.IO.ErrOut, "Import job %s is %s (%.0f%%). Use --wait to follow progress.\n", jobID, job.Status, job.Rate*100)
			return nil
		}
	}

	// Build structured result
	result := importStatusResult{
		ID:        job.ID,
		FileName:  job.FileName,
		Status:    job.Status,
		Total:     job.Total,
		SuccessNo: job.SuccessNo,
		FailNo:    job.FailNo,
		Rate:      job.Rate,
		CreatedAt: job.CreatedAt,
	}
	if job.FailNo > 0 {
		if details, detailErr := getFailedDetails(client, job.ID); detailErr == nil {
			result.Failed = details
		}
	}

	// Output structured data to stdout
	output, _ := cmd.Flags().GetString("output")
	body, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return marshalErr
	}
	if err := iostreams.FormatOutput(body, f.IO, output, nil); err != nil {
		return err
	}

	// Human-readable summary to stderr
	showImportResultSummary(f, job)
	if job.FailNo > 0 && len(result.Failed) > 0 {
		fmt.Fprintf(f.IO.ErrOut, "\nFailed rows:\n")
		for _, d := range result.Failed {
			fmt.Fprintf(f.IO.ErrOut, "  Row %d: %s (%s) — %s\n", d.Row, d.SerialNumber, d.DeviceName, d.FailReason)
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
