package overview

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdOverview(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Platform overview dashboard",
		Long:  "Show platform overview statistics: devices, alerts, traffic, and offline.",
	}

	cmd.AddCommand(NewCmdDevices(f))
	cmd.AddCommand(NewCmdAlerts(f))
	cmd.AddCommand(NewCmdTraffic(f))
	cmd.AddCommand(NewCmdOffline(f))

	return cmd
}

// applyDefaultTimeRange2 sets after/before to the last 7 days when not specified.
func applyDefaultTimeRange2(after, before *string) {
	now := time.Now()
	if *before == "" {
		*before = now.Format("2006-01-02")
	}
	if *after == "" {
		*after = now.AddDate(0, 0, -7).Format("2006-01-02")
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func fieldIndex(fields []string, name string) int {
	for i, f := range fields {
		if f == name {
			return i
		}
	}
	return -1
}

func formatBytes(b float64) string {
	return iostreams.FormatBytes(strconv.FormatFloat(b, 'f', -1, 64))
}

// makeQuery builds url.Values from a map, skipping empty values.
func makeQuery(params map[string]string) url.Values {
	q := make(url.Values)
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	return q
}

// makeQueryWithGroups builds url.Values from a map plus repeated group IDs.
func makeQueryWithGroups(params map[string]string, groups []string) url.Values {
	q := makeQuery(params)
	for _, g := range groups {
		q.Add("devicegroupId", g)
	}
	return q
}

// unwrapResult extracts the "result" field from API response envelope, or returns body as-is.
func unwrapResult(body []byte) json.RawMessage {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if json.Unmarshal(body, &envelope) == nil && envelope.Result != nil {
		return envelope.Result
	}
	return body
}
