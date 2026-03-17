package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleUpdateOptions struct {
	Types     []string
	Channels  []string
	Users     []string
	Webhooks  []string
	Days      []string
	StartTime string
	EndTime   string
}

func NewCmdRuleUpdate(f *factory.Factory) *cobra.Command {
	opts := &RuleUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <rule-id>",
		Short: "Update an alert rule",
		Long: `Update an existing alert rule. This is a full replacement of rules and notification
settings — all flags must be provided (use "alert rule get" to view current values).

Target bindings (type and targetIds) are preserved from the existing rule.`,
		Example: `  # Update rule types and channels
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type REBOOT --type FIRMWARE_UPGRADE \
    --channel EMAIL --channel APP

  # Update with active time window
  incloud alert rule update 507f1f77bcf86cd799439011 \
    --type DISCONNECTED \
    --channel EMAIL \
    --day MONDAY --day TUESDAY --day WEDNESDAY --day THURSDAY --day FRIDAY \
    --start-time 09:00 --end-time 18:00`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleID := args[0]

			if len(opts.Types) == 0 {
				return fmt.Errorf("--type is required")
			}
			if len(opts.Channels) == 0 {
				return fmt.Errorf("--channel is required")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			// GET existing rule to preserve type and targetIds
			getURL := ctx.Host + "/api/v1/alerts/rules/" + ruleID
			getReq, err := http.NewRequestWithContext(context.Background(), "GET", getURL, http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			getResp, err := client.Do(getReq)
			if err != nil {
				return fmt.Errorf("fetching existing rule: %w", err)
			}
			defer getResp.Body.Close()

			getBody, err := io.ReadAll(getResp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if getResp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", getResp.StatusCode, string(getBody))
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

			// Build update body
			rules := make([]map[string]any, len(opts.Types))
			for i, t := range opts.Types {
				rules[i] = map[string]any{
					"type":  t,
					"param": map[string]any{},
				}
			}

			notify := map[string]any{
				"channels": opts.Channels,
			}
			if len(opts.Users) > 0 {
				notify["users"] = opts.Users
			}
			if len(opts.Webhooks) > 0 {
				notify["webhooks"] = opts.Webhooks
			}
			if len(opts.Days) > 0 {
				notify["activeDayOfWeeks"] = opts.Days
			}
			if opts.StartTime != "" {
				notify["startTime"] = opts.StartTime
			}
			if opts.EndTime != "" {
				notify["endTime"] = opts.EndTime
			}

			reqBody := map[string]any{
				"type":      existing.Result.Type,
				"targetIds": existing.Result.TargetIDs,
				"rules":     rules,
				"notify":    notify,
			}

			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("marshaling request body: %w", err)
			}

			putURL := ctx.Host + "/api/v1/alerts/rules/" + ruleID
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, putURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, nil)
		},
	}

	cmd.Flags().StringArrayVar(&opts.Types, "type", nil, "Alert type (required, can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Channels, "channel", nil, "Notification channel (required, can be repeated: SMS/APP/EMAIL/WEBHOOK/SUBSCRIPTION)")
	cmd.Flags().StringArrayVar(&opts.Users, "user", nil, "User ID to notify (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Webhooks, "webhook", nil, "Webhook ID for notification (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Days, "day", nil, "Active day of week (can be repeated: MONDAY..SUNDAY, default all)")
	cmd.Flags().StringVar(&opts.StartTime, "start-time", "", "Active start time (HH:mm, default 00:00)")
	cmd.Flags().StringVar(&opts.EndTime, "end-time", "", "Active end time (HH:mm, default 23:59)")

	return cmd
}
