package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
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
  incloud device import-status 69c371131e0e4d15c8cb2b25 --wait`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportStatus(f, opts, args[0])
		},
	}

	cmd.Flags().BoolVarP(&opts.Wait, "wait", "w", false, "Wait for the import to complete if still running")

	return cmd
}

func runImportStatus(f *factory.Factory, opts *ImportStatusOptions, jobID string) error {
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

	return showImportResult(f, client, job)
}
