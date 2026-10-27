package alert

import (
	"fmt"
	"slices"
	"time"
	_ "time/tzdata" // embed IANA zone data so --timezone validation works on Windows

	"github.com/spf13/cobra"
)

// NotifyOptions holds notification-related flags shared by rule create and update.
type NotifyOptions struct {
	Channels  []string
	Users     []string
	Webhooks  []string
	Days      []string
	StartTime string
	EndTime   string
	Timezone  string
}

// notifyFields are the persisted notify fields carried over by rule update.
// Response-only fields (userDetails, webhookDetails) are dropped.
var notifyFields = []string{"channels", "users", "webhooks", "emails", "activeDayOfWeeks", "startTime", "endTime", "timezone"}

// RegisterFlags registers notification flags on the given command.
func (o *NotifyOptions) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVar(&o.Channels, "channel", nil, "Notification channel (required on create, can be repeated: SMS/APP/EMAIL/WEBHOOK/SUBSCRIPTION)")
	cmd.Flags().StringArrayVar(&o.Users, "user", nil, "User ID to notify (can be repeated)")
	cmd.Flags().StringArrayVar(&o.Webhooks, "webhook", nil, "Webhook ID for notification (can be repeated)")
	cmd.Flags().StringArrayVar(&o.Days, "day", nil, "Active day of week (can be repeated: MONDAY..SUNDAY, default all)")
	cmd.Flags().StringVar(&o.StartTime, "start-time", "", "Active start time (HH:mm, default 00:00)")
	cmd.Flags().StringVar(&o.EndTime, "end-time", "", "Active end time (HH:mm, default 23:59)")
	cmd.Flags().StringVar(&o.Timezone, "timezone", "", "IANA timezone in which --day/--start-time/--end-time are evaluated (e.g. Asia/Shanghai; default: server timezone)")
}

// Validate checks flag values locally before sending the request.
func (o *NotifyOptions) Validate() error {
	if o.Timezone == "" {
		return nil
	}
	// time.LoadLocation accepts "Local", which is not an IANA name the server understands.
	if o.Timezone == "Local" {
		return fmt.Errorf("invalid --timezone %q: must be an IANA timezone name such as Asia/Shanghai", o.Timezone)
	}
	if _, err := time.LoadLocation(o.Timezone); err != nil {
		return fmt.Errorf("invalid --timezone %q: must be an IANA timezone name such as Asia/Shanghai", o.Timezone)
	}
	return nil
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
	if o.Timezone != "" {
		m["timezone"] = o.Timezone
	}
	return m
}

// MergeInto builds the notify body for rule update: it starts from the rule's
// existing notify and overrides only the flags explicitly set on cmd. A flag
// set to an empty value clears that field. The backend replaces notify as a
// whole, so fields not carried over here (e.g. emails set in the portal) would
// otherwise be lost.
func (o *NotifyOptions) MergeInto(existing map[string]any, cmd *cobra.Command) map[string]any {
	m := make(map[string]any, len(notifyFields))
	for _, k := range notifyFields {
		if v, ok := existing[k]; ok && v != nil {
			m[k] = v
		}
	}

	setList := func(flag, key string, v []string) {
		if !cmd.Flags().Changed(flag) {
			return
		}
		v = slices.DeleteFunc(slices.Clone(v), func(s string) bool { return s == "" })
		if len(v) == 0 {
			delete(m, key)
		} else {
			m[key] = v
		}
	}
	setString := func(flag, key, v string) {
		if !cmd.Flags().Changed(flag) {
			return
		}
		if v == "" {
			delete(m, key)
		} else {
			m[key] = v
		}
	}

	setList("channel", "channels", o.Channels)
	setList("user", "users", o.Users)
	setList("webhook", "webhooks", o.Webhooks)
	setList("day", "activeDayOfWeeks", o.Days)
	setString("start-time", "startTime", o.StartTime)
	setString("end-time", "endTime", o.EndTime)
	setString("timezone", "timezone", o.Timezone)
	return m
}

// hasItems reports whether v is a non-empty list, either from flags ([]string)
// or decoded from JSON ([]any).
func hasItems(v any) bool {
	switch l := v.(type) {
	case []string:
		return len(l) > 0
	case []any:
		return len(l) > 0
	}
	return false
}
