package firmware

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecCancel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <executionId>",
		Short: "Cancel an OTA job execution",
		Long:  "Cancel a queued or in-progress OTA job execution for a specific device.",
		Example: `  # Cancel an execution
  incloud firmware job executions cancel 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			execID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			_, err = client.Put("/api/v1/job/executions/"+url.PathEscape(execID)+"/cancel", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Execution %q canceled.\n", execID)
			return nil
		},
	}

	return cmd
}
