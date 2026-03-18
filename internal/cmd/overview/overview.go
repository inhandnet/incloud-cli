package overview

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OverviewOptions struct {
	After  string
	Before string
	N      int
}

func NewCmdOverview(f *factory.Factory) *cobra.Command {
	opts := &OverviewOptions{}

	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Platform overview dashboard",
		Long:  "Show a summary dashboard with device status, alerts, traffic, and offline statistics.",
		Example: `  # Show overview dashboard
  incloud overview

  # With custom time range
  incloud overview --after 2024-01-01 --before 2024-01-31

  # JSON output
  incloud overview -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOverview(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2024-01-01 or 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2024-01-31 or 2024-01-31T23:59:59)")
	cmd.Flags().IntVar(&opts.N, "n", 3, "Number of top items to show")

	cmd.AddCommand(NewCmdDevices(f))
	cmd.AddCommand(NewCmdAlerts(f))
	cmd.AddCommand(NewCmdTraffic(f))
	cmd.AddCommand(NewCmdOffline(f))

	return cmd
}

func runOverview(cmd *cobra.Command, f *factory.Factory, opts *OverviewOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	applyDefaultTimeRange(opts)

	nStr := strconv.Itoa(opts.N)
	timeQuery := makeQuery(map[string]string{
		"after": opts.After, "before": opts.Before,
	})
	topQuery := makeQuery(map[string]string{
		"n": nStr, "after": opts.After, "before": opts.Before,
	})
	offlineQuery := makeQuery(map[string]string{
		"topN": nStr, "after": opts.After, "before": opts.Before,
	})

	type apiReq struct {
		name  string
		path  string
		query url.Values
	}

	reqs := []apiReq{
		{"summary", "/api/v1/devices/summary", nil},
		{"alertStats", "/api/v1/alerts/stats", nil},
		{"topTypes", "/api/v1/alert/top-alert-types", topQuery},
		{"traffic", "/api/v1/datausage/overview", timeQuery},
		{"offline", "/api/v1/devices/offline/topn", offlineQuery},
	}

	results := make(map[string]json.RawMessage)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for _, r := range reqs {
		wg.Add(1)
		go func(r apiReq) {
			defer wg.Done()
			body, err := client.Get(r.path, r.query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("%s: %w", r.name, err)
				}
				return
			}
			results[r.name] = unwrapResult(body)
		}(r)
	}
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json", "jsonc", "yaml":
		merged := map[string]json.RawMessage{
			"devices":    results["summary"],
			"alerts":     results["alertStats"],
			"alertTypes": results["topTypes"],
			"traffic":    results["traffic"],
			"offline":    results["offline"],
		}
		b, _ := json.Marshal(merged)
		if output == "yaml" {
			s, err := iostreams.FormatYAML(b)
			if err != nil {
				return err
			}
			fmt.Fprintln(f.IO.Out, s)
		} else {
			fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(b, f.IO, output))
		}
	default:
		printDashboard(f.IO, opts, results)
	}

	return nil
}

// applyDefaultTimeRange sets After/Before to the last 7 days when not specified.
func applyDefaultTimeRange(opts *OverviewOptions) {
	applyDefaultTimeRange2(&opts.After, &opts.Before)
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

func printDashboard(streams *iostreams.IOStreams, opts *OverviewOptions, data map[string]json.RawMessage) {
	c := iostreams.NewColorizer(streams.TermOutput())
	out := streams.Out

	// --- Time Range ---
	fmt.Fprintf(out, "%s %s ~ %s\n\n", c.Bold("Period:"), opts.After, opts.Before)

	// --- Devices ---
	fmt.Fprintln(out, c.Bold("Devices"))
	var summary struct {
		Count struct {
			Total    int `json:"total"`
			Online   int `json:"online"`
			Offline  int `json:"offline"`
			Inactive int `json:"inactive"`
		} `json:"count"`
	}
	if json.Unmarshal(data["summary"], &summary) == nil {
		cnt := summary.Count
		fmt.Fprintf(out, "  Total: %s  Online: %s  Offline: %s  Inactive: %s\n",
			c.Bold(strconv.Itoa(cnt.Total)),
			c.Green(strconv.Itoa(cnt.Online)),
			c.Red(strconv.Itoa(cnt.Offline)),
			c.Yellow(strconv.Itoa(cnt.Inactive)),
		)
	}
	fmt.Fprintln(out)

	// --- Alerts ---
	fmt.Fprintln(out, c.Bold("Alerts"))
	var alertStats struct {
		Active int `json:"active"`
		Closed int `json:"closed"`
		Total  int `json:"total"`
	}
	if json.Unmarshal(data["alertStats"], &alertStats) == nil {
		fmt.Fprintf(out, "  Active: %s  Closed: %s  Total: %s\n",
			c.Red(strconv.Itoa(alertStats.Active)),
			c.Green(strconv.Itoa(alertStats.Closed)),
			c.Bold(strconv.Itoa(alertStats.Total)),
		)
	}

	var topTypes []struct {
		Type  string `json:"type"`
		Value int    `json:"value"`
	}
	if json.Unmarshal(data["topTypes"], &topTypes) == nil && len(topTypes) > 0 {
		fmt.Fprint(out, "  Top types: ")
		for i, t := range topTypes {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}
			fmt.Fprintf(out, "%s(%d)", t.Type, t.Value)
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out)

	// --- Traffic ---
	fmt.Fprintln(out, c.Bold("Traffic"))
	var trafficData struct {
		Series []struct {
			Type   string          `json:"type"`
			Fields []string        `json:"fields"`
			Data   [][]interface{} `json:"data"`
		} `json:"series"`
	}
	if json.Unmarshal(data["traffic"], &trafficData) == nil && len(trafficData.Series) > 0 {
		for _, s := range trafficData.Series {
			txIdx := fieldIndex(s.Fields, "tx")
			rxIdx := fieldIndex(s.Fields, "rx")
			totalIdx := fieldIndex(s.Fields, "total")
			var txSum, rxSum, totalSum float64
			for _, row := range s.Data {
				if txIdx >= 0 && txIdx < len(row) {
					txSum += toFloat(row[txIdx])
				}
				if rxIdx >= 0 && rxIdx < len(row) {
					rxSum += toFloat(row[rxIdx])
				}
				if totalIdx >= 0 && totalIdx < len(row) {
					totalSum += toFloat(row[totalIdx])
				}
			}
			fmt.Fprintf(out, "  %s — TX: %s  RX: %s  Total: %s\n",
				s.Type,
				c.Bold(formatBytes(txSum)),
				c.Bold(formatBytes(rxSum)),
				c.Bold(formatBytes(totalSum)),
			)
		}
	} else {
		fmt.Fprintln(out, c.Gray("  No traffic data"))
	}
	fmt.Fprintln(out)

	// --- Offline ---
	fmt.Fprintln(out, c.Bold("Top Offline Devices"))
	var offlineDevices []struct {
		DeviceName   string `json:"deviceName"`
		SerialNumber string `json:"serialNumber"`
		OfflineTimes int    `json:"offlineTimes"`
	}
	if json.Unmarshal(data["offline"], &offlineDevices) == nil && len(offlineDevices) > 0 {
		tp := iostreams.NewTablePrinter(out, streams.IsStdoutTTY())
		tp.AddRow(c.Bold("DEVICE"), c.Bold("SERIAL"), c.Bold("OFFLINE TIMES"))
		for _, d := range offlineDevices {
			name := d.DeviceName
			if name == "" {
				name = d.SerialNumber
			}
			tp.AddRow(name, d.SerialNumber, strconv.Itoa(d.OfflineTimes))
		}
		_ = tp.Render()
	} else {
		fmt.Fprintln(out, c.Gray("  No offline data"))
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
