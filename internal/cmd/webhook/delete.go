package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <webhook-id>",
		Short: "Delete a webhook",
		Long:  "Delete a message webhook by ID.",
		Example: `  # Delete a webhook (will prompt for confirmation)
  incloud webhook delete 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud webhook delete 507f1f77bcf86cd799439011 --yes`,
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete webhook %s?", webhookID))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			if _, err := client.Delete("/api/v1/message/webhooks/" + webhookID); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Deleted webhook %s.\n", webhookID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
