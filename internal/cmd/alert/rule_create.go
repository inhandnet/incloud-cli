package alert

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleCreateOptions struct {
	TargetType string
	Targets    []string
	Types      []string
	Notify     NotifyOptions
}

func NewCmdRuleCreate(f *factory.Factory) *cobra.Command {
	opts := &RuleCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an alert rule",
		Long: `Create a new alert rule with specified targets, alert types, and notification settings.

Target types: org, group, device

The --type flag accepts three formats:
  - Type name only:       --type reboot
  - With parameters:      --type "disconnected,retention=600"
  - JSON object:          --type '{"type":"disconnected","param":{"retention":600}}'

Use 'incloud alert rule types' to see all supported types and their parameters.

Supported notification channels:
  SMS, APP, EMAIL, WEBHOOK, SUBSCRIPTION`,
		Example: `  # Create a rule for an org
  incloud alert rule create \
    --target-type org --target 507f1f77bcf86cd799439011 \
    --type disconnected \
    --channel EMAIL --channel APP

  # Create a rule with parameters
  incloud alert rule create \
    --target-type org --target 507f1f77bcf86cd799439011 \
    --type "disconnected,retention=600" \
    --channel EMAIL

  # Create with JSON format
  incloud alert rule create \
    --target-type device --target 507f1f77bcf86cd799439011 \
    --type '{"type":"high_average_cpu_utilization","param":{"retention":600,"threshold":80}}' \
    --channel APP

  # Create a rule for device groups with multiple types
  incloud alert rule create \
    --target-type group --target 507f1f77bcf86cd799439011 \
    --type disconnected --type reboot \
    --channel EMAIL \
    --user 607f1f77bcf86cd799439022

  # Create a rule with active time window
  incloud alert rule create \
    --target-type device --target 507f1f77bcf86cd799439011 \
    --type disconnected \
    --channel EMAIL \
    --day MONDAY --day TUESDAY --day WEDNESDAY --day THURSDAY --day FRIDAY \
    --start-time 09:00 --end-time 18:00`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := ParseTypeFlags(opts.Types)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]any{
				"type":      opts.TargetType,
				"targetIds": opts.Targets,
				"rules":     RulesToRequestBody(rules),
				"notify":    opts.Notify.ToMap(),
			}

			body, err := client.Post("/api/v1/alerts/rules", reqBody)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					ID string `json:"_id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(body, &resp); err == nil && resp.Result.ID != "" {
				fmt.Fprintf(f.IO.ErrOut, "Alert rule created. (id: %s)\n", resp.Result.ID)
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.TargetType, "target-type", "", "Target type: org, group, or device (required)")
	cmd.Flags().StringArrayVar(&opts.Targets, "target", nil, "Target ID (required, can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Types, "type", nil, "Alert type with optional params (required, can be repeated; use 'incloud alert rule types' to list)")
	opts.Notify.RegisterFlags(cmd)

	_ = cmd.MarkFlagRequired("target-type")
	_ = cmd.MarkFlagRequired("target")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("channel")

	return cmd
}
