package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type UpdateOptions struct {
	Webhook  string
	Provider string
}

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <webhook-id>",
		Short: "Update a webhook",
		Long:  "Update an existing webhook's URL or provider.",
		Example: `  # Update webhook URL
  incloud webhook update 507f1f77bcf86cd799439011 --url https://example.com/new-hook

  # Update provider
  incloud webhook update 507f1f77bcf86cd799439011 --provider wechat`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := make(map[string]any)
			if cmd.Flags().Changed("url") {
				reqBody["webhook"] = opts.Webhook
			}
			if cmd.Flags().Changed("provider") {
				reqBody["provider"] = opts.Provider
			}

			if len(reqBody) == 0 {
				return fmt.Errorf("at least one of --url or --provider must be specified")
			}

			body, err := client.Put("/api/v1/message/webhooks/"+webhookID, reqBody)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Webhook (%s) updated.\n", webhookID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Webhook, "url", "", "Webhook URL")
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Webhook provider (supported: wechat)")

	return cmd
}
