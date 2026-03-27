package firmware

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdExecRetry(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry <executionId>",
		Short: "Retry a failed OTA job execution",
		Long: `Retry a failed OTA job execution. This creates a new OTA job targeting
the same device with the same firmware.

Only executions in FAILED status can be retried.`,
		Example: `  # Retry a failed execution
  incloud firmware job executions retry 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			execID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Put("/api/v1/job/executions/"+url.PathEscape(execID)+"/retry", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Execution %q retried, new job created.\n", execID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
