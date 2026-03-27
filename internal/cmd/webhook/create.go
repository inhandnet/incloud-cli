package webhook

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type CreateOptions struct {
	Name     string
	Webhook  string
	Provider string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		Long:  "Create a new message webhook configuration.",
		Example: `  # Create a WeChat webhook
  incloud webhook create --name "WeChat Alert" \
    --url https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx \
    --provider wechat

  # Create a generic webhook (receives JSON payload)
  incloud webhook create --name "Custom Alert" \
    --url https://example.com/webhook \
    --provider generic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]any{
				"name":     opts.Name,
				"webhook":  opts.Webhook,
				"provider": opts.Provider,
			}

			body, err := client.Post("/api/v1/message/webhooks", reqBody)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					ID string `json:"_id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(body, &resp); err == nil && resp.Result.ID != "" {
				fmt.Fprintf(f.IO.ErrOut, "Webhook created. (id: %s)\n", resp.Result.ID)
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Webhook name (required)")
	cmd.Flags().StringVar(&opts.Webhook, "url", "", "Webhook URL (required)")
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Webhook provider (required; supported: wechat, generic)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("provider")

	return cmd
}
