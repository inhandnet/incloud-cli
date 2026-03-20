package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleCreateOptions struct {
	TargetType string
	Targets    []string
	Types      []string
	Channels   []string
	Users      []string
	Webhooks   []string
	Days       []string
	StartTime  string
	EndTime    string
}

func (o *RuleCreateOptions) toRequestBody() map[string]any {
	rules := make([]map[string]any, len(o.Types))
	for i, t := range o.Types {
		rules[i] = map[string]any{
			"type":  t,
			"param": map[string]any{},
		}
	}

	notify := map[string]any{
		"channels": o.Channels,
	}
	if len(o.Users) > 0 {
		notify["users"] = o.Users
	}
	if len(o.Webhooks) > 0 {
		notify["webhooks"] = o.Webhooks
	}
	if len(o.Days) > 0 {
		notify["activeDayOfWeeks"] = o.Days
	}
	if o.StartTime != "" {
		notify["startTime"] = o.StartTime
	}
	if o.EndTime != "" {
		notify["endTime"] = o.EndTime
	}

	return map[string]any{
		"type":      o.TargetType,
		"targetIds": o.Targets,
		"rules":     rules,
		"notify":    notify,
	}
}

func NewCmdRuleCreate(f *factory.Factory) *cobra.Command {
	opts := &RuleCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an alert rule",
		Long: `Create a new alert rule with specified targets, alert types, and notification settings.

Target types: org, group, device

Supported alert types:
  CONNECTED, DISCONNECTED, CONFIG_SYNC_FAILED, SIM_SWITCH,
  LOCAL_CONFIG_UPDATE, REBOOT, FIRMWARE_UPGRADE, LICENSE_EXPIRING,
  LICENSE_EXPIRED, UPLINK_SWITCH, ETHERNET_WAN_CONNECTED,
  ETHERNET_WAN_DISCONNECTED, MODEM_WAN_CONNECTED, MODEM_WAN_DISCONNECTED,
  WWAN_CONNECTED, WWAN_DISCONNECTED, CLIENT_CONNECTED, CLIENT_DISCONNECTED,
  CELL_OPERATOR_SWITCH, BRIDGE_LOOP_DETECT, CELL_TRAFFIC_REACH_THRESHOLD,
  DEVICE_POWER_OFF

Supported notification channels:
  SMS, APP, EMAIL, WEBHOOK, SUBSCRIPTION`,
		Example: `  # Create a rule for an org
  incloud alert rule create \
    --target-type org --target 507f1f77bcf86cd799439011 \
    --type DISCONNECTED \
    --channel EMAIL --channel APP

  # Create a rule for device groups
  incloud alert rule create \
    --target-type group --target 507f1f77bcf86cd799439011 \
    --type DISCONNECTED --type REBOOT \
    --channel EMAIL \
    --user 607f1f77bcf86cd799439022

  # Create a rule for specific devices with active time window
  incloud alert rule create \
    --target-type device --target 507f1f77bcf86cd799439011 \
    --type DISCONNECTED \
    --channel EMAIL \
    --day MONDAY --day TUESDAY --day WEDNESDAY --day THURSDAY --day FRIDAY \
    --start-time 09:00 --end-time 18:00`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Post("/api/v1/alerts/rules", opts.toRequestBody())
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&opts.TargetType, "target-type", "", "Target type: org, group, or device (required)")
	cmd.Flags().StringArrayVar(&opts.Targets, "target", nil, "Target ID (required, can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Types, "type", nil, "Alert type (required, can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Channels, "channel", nil, "Notification channel (required, can be repeated: SMS/APP/EMAIL/WEBHOOK/SUBSCRIPTION)")
	cmd.Flags().StringArrayVar(&opts.Users, "user", nil, "User ID to notify (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Webhooks, "webhook", nil, "Webhook ID for notification (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Days, "day", nil, "Active day of week (can be repeated: MONDAY..SUNDAY, default all)")
	cmd.Flags().StringVar(&opts.StartTime, "start-time", "", "Active start time (HH:mm, default 00:00)")
	cmd.Flags().StringVar(&opts.EndTime, "end-time", "", "Active end time (HH:mm, default 23:59)")

	_ = cmd.MarkFlagRequired("target-type")
	_ = cmd.MarkFlagRequired("target")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("channel")

	return cmd
}
