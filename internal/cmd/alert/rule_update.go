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
		Long: `Update an existing alert rule.

Alert types (--type) are a full replacement: pass every type the rule should keep.

Notification settings are merged into the existing ones: only the notification flags
you pass are changed, everything else (including timezone and external email
recipients configured in the portal) is preserved. Pass a flag with an empty value
to clear that setting, e.g. --timezone "" or --user "".

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

  # Only change the notification timezone, keep everything else
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type disconnected \
    --timezone Europe/Berlin

  # Update with active time window
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type disconnected \
    --channel EMAIL \
    --day MONDAY --day TUESDAY --day WEDNESDAY --day THURSDAY --day FRIDAY \
    --start-time 09:00 --end-time 18:00`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleID := args[0]

			if err := opts.Notify.Validate(); err != nil {
				return err
			}

			rules, err := ParseTypeFlags(opts.Types)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// GET existing rule to preserve type, targetIds and notify
			getBody, err := client.Get("/api/v1/alerts/rules/"+ruleID, nil)
			if err != nil {
				return err
			}

			var existing struct {
				Result struct {
					Type      string         `json:"type"`
					TargetIDs []string       `json:"targetIds"`
					Notify    map[string]any `json:"notify"`
				} `json:"result"`
			}
			if err := json.Unmarshal(getBody, &existing); err != nil {
				return fmt.Errorf("parsing existing rule: %w", err)
			}

			notify := opts.Notify.MergeInto(existing.Result.Notify, cmd)
			// An update without channels would null out notify and silently stop notifications.
			if !hasItems(notify["channels"]) {
				return fmt.Errorf("rule has no notification channel; pass --channel")
			}

			reqBody := map[string]any{
				"type":      existing.Result.Type,
				"targetIds": existing.Result.TargetIDs,
				"rules":     RulesToRequestBody(rules),
				"notify":    notify,
			}

			body, err := client.Put("/api/v1/alerts/rules/"+ruleID, reqBody)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Alert rule (%s) updated.\n", ruleID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringArrayVar(&opts.Types, "type", nil, `Alert type (required, can be repeated; use 'incloud alert rule types' to list all)`)
	opts.Notify.RegisterFlags(cmd)

	_ = cmd.MarkFlagRequired("type")

	return cmd
}
