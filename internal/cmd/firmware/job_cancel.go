package firmware

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdJobCancel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <jobId>",
		Short: "Cancel an OTA firmware upgrade job",
		Long: `Cancel a queued or in-progress OTA firmware upgrade job.

All pending executions under this job will also be canceled.`,
		Example: `  # Cancel an OTA job
  incloud firmware job cancel 20260318000001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			_, err = client.Put("/api/v1/jobs/"+url.PathEscape(jobID)+"/cancel", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Job %q canceled.\n", jobID)
			return nil
		},
	}

	return cmd
}
