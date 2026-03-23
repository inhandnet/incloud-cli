package alert

import "github.com/spf13/cobra"

// NotifyOptions holds notification-related flags shared by rule create and update.
type NotifyOptions struct {
	Channels  []string
	Users     []string
	Webhooks  []string
	Days      []string
	StartTime string
	EndTime   string
}

// RegisterFlags registers notification flags on the given command.
func (o *NotifyOptions) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Channels, "channel", nil, "Notification channel (required, can be repeated: SMS/APP/EMAIL/WEBHOOK/SUBSCRIPTION)")
	cmd.Flags().StringArrayVar(&o.Users, "user", nil, "User ID to notify (can be repeated)")
	cmd.Flags().StringArrayVar(&o.Webhooks, "webhook", nil, "Webhook ID for notification (can be repeated)")
	cmd.Flags().StringArrayVar(&o.Days, "day", nil, "Active day of week (can be repeated: MONDAY..SUNDAY, default all)")
	cmd.Flags().StringVar(&o.StartTime, "start-time", "", "Active start time (HH:mm, default 00:00)")
	cmd.Flags().StringVar(&o.EndTime, "end-time", "", "Active end time (HH:mm, default 23:59)")
}

// ToMap converts NotifyOptions to the API request body format.
func (o *NotifyOptions) ToMap() map[string]any {
	m := map[string]any{
		"channels": o.Channels,
	}
	if len(o.Users) > 0 {
		m["users"] = o.Users
	}
	if len(o.Webhooks) > 0 {
		m["webhooks"] = o.Webhooks
	}
	if len(o.Days) > 0 {
		m["activeDayOfWeeks"] = o.Days
	}
	if o.StartTime != "" {
		m["startTime"] = o.StartTime
	}
	if o.EndTime != "" {
		m["endTime"] = o.EndTime
	}
	return m
}
