package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTest(f *factory.Factory) *cobra.Command {
	var webhookURL string

	cmd := &cobra.Command{
		Use:   "test [id]",
		Short: "Send a test message to a webhook",
		Long:  "Send a test message to verify the webhook is working correctly.\nEither provide a webhook ID as argument, or use --url to test a URL directly.",
		Example: `  # Test a saved webhook by ID
  incloud webhook test 507f1f77bcf86cd799439011

  # Test a webhook URL directly
  incloud webhook test --url https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				// Test by ID
				id := args[0]
				if _, err := client.Post(fmt.Sprintf("/api/v1/message/webhooks/%s/test", id), nil); err != nil {
					return err
				}
			} else if webhookURL != "" {
				// Test by URL (legacy)
				reqBody := map[string]any{
					"webhook": webhookURL,
				}
				if _, err := client.Post("/api/v1/message/webhooks/send", reqBody); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("either provide a webhook ID as argument or use --url flag")
			}

			fmt.Fprintf(f.IO.ErrOut, "Test message sent successfully.\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&webhookURL, "url", "", "Webhook URL to test directly")

	return cmd
}
