package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTest(f *factory.Factory) *cobra.Command {
	var webhookURL string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Send a test message to a webhook",
		Long:  "Send a test message to verify the webhook URL is working correctly.",
		Example: `  # Test a webhook URL
  incloud webhook test --url https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]any{
				"webhook": webhookURL,
			}

			if _, err := client.Post("/api/v1/message/webhooks/send", reqBody); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Test message sent successfully.\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&webhookURL, "url", "", "Webhook URL to test (required)")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}
