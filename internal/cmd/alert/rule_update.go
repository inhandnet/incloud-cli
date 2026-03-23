package alert

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleUpdateOptions struct {
	Types  []string
	Notify NotifyOptions
}

func NewCmdRuleUpdate(f *factory.Factory) *cobra.Command {
	opts := &RuleUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <rule-id>",
		Short: "Update an alert rule",
		Long: `Update an existing alert rule. This is a full replacement of rules and notification
settings — all flags must be provided (use "alert rule get" to view current values).

Target bindings (type and targetIds) are preserved from the existing rule.

The --type flag accepts three formats:
  - Type name only:       --type reboot
  - With parameters:      --type "disconnected,retention=600"
  - JSON object:          --type '{"type":"disconnected","param":{"retention":600}}'

Use 'incloud alert rule types' to see all supported types and their parameters.`,
		Example: `  # Update rule types and channels
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type reboot --type firmware_upgrade \
    --channel EMAIL --channel APP

  # Update with type parameters
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type "disconnected,retention=600" \
    --channel EMAIL

  # Update with JSON format
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type '{"type":"high_average_cpu_utilization","param":{"retention":600,"threshold":80}}' \
    --channel APP

  # Update with active time window
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type disconnected \
    --channel EMAIL \
    --day MONDAY --day TUESDAY --day WEDNESDAY --day THURSDAY --day FRIDAY \
    --start-time 09:00 --end-time 18:00`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleID := args[0]

			rules, err := ParseTypeFlags(opts.Types)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// GET existing rule to preserve type and targetIds
			getBody, err := client.Get("/api/v1/alerts/rules/"+ruleID, nil)
			if err != nil {
				return err
			}

			var existing struct {
				Result struct {
					Type      string   `json:"type"`
					TargetIDs []string `json:"targetIds"`
				} `json:"result"`
			}
			if err := json.Unmarshal(getBody, &existing); err != nil {
				return fmt.Errorf("parsing existing rule: %w", err)
			}

			reqBody := map[string]any{
				"type":      existing.Result.Type,
				"targetIds": existing.Result.TargetIDs,
				"rules":     RulesToRequestBody(rules),
				"notify":    opts.Notify.ToMap(),
			}

			body, err := client.Put("/api/v1/alerts/rules/"+ruleID, reqBody)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Alert rule (%s) updated.\n", ruleID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, nil)
		},
	}

	cmd.Flags().StringArrayVar(&opts.Types, "type", nil, `Alert type (required, can be repeated; use 'incloud alert rule types' to list all)`)
	opts.Notify.RegisterFlags(cmd)

	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("channel")

	return cmd
}
