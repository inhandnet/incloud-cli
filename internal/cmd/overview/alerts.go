package overview

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type AlertsOptions struct {
	After  string
	Before string
	Group  []string
	N      int
	Fields []string
}

var (
	defaultTopAlertDevicesFields = []string{"deviceName", "serialNumber", "value"}
	defaultTopAlertTypesFields   = []string{"type", "value"}
)

func NewCmdAlerts(f *factory.Factory) *cobra.Command {
	opts := &AlertsOptions{}

	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Alert statistics and top rankings",
		Long:  "Show alert summary stats, top devices by alert count, and top alert types.",
		Example: `  # Show alert overview
  incloud overview alerts

  # Filter by time range
  incloud overview alerts --after 2024-01-01 --before 2024-01-31

  # Top 5 with device group filter
  incloud overview alerts --n 5 --group 507f1f77bcf86cd799439011

  # JSON output
  incloud overview alerts -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlerts(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2024-01-01 or 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2024-01-31 or 2024-01-31T23:59:59)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().IntVar(&opts.N, "n", 10, "Number of top items to show")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in tables")

	return cmd
}

func runAlerts(cmd *cobra.Command, f *factory.Factory, opts *AlertsOptions) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	actx, err := cfg.ActiveContext()
	if err != nil {
		return err
	}
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	host := actx.Host
	nStr := strconv.Itoa(opts.N)

	// Build URLs for top-alert-devices and top-alert-types with group support
	buildTopURL := func(path string) string {
		u, err := url.Parse(host + path)
		if err != nil {
			return host + path
		}
		q := u.Query()
		q.Set("n", nStr)
		if opts.After != "" {
			q.Set("after", opts.After)
		}
		if opts.Before != "" {
			q.Set("before", opts.Before)
		}
		for _, g := range opts.Group {
			q.Add("devicegroupId", g)
		}
		u.RawQuery = q.Encode()
		return u.String()
	}

	type apiReq struct {
		name string
		url  string
	}

	reqs := []apiReq{
		{"stats", host + "/api/v1/alerts/stats"},
		{"topDevices", buildTopURL("/api/v1/alert/top-alert-devices")},
		{"topTypes", buildTopURL("/api/v1/alert/top-alert-types")},
	}

	results := make(map[string]json.RawMessage)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for _, r := range reqs {
		wg.Add(1)
		go func(r apiReq) {
			defer wg.Done()
			body, err := doGet(client, r.url)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("%s: %w", r.name, err)
				}
				return
			}
			var envelope struct {
				Result json.RawMessage `json:"result"`
			}
			if json.Unmarshal(body, &envelope) == nil && envelope.Result != nil {
				results[r.name] = envelope.Result
			} else {
				results[r.name] = body
			}
		}(r)
	}
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json", "jsonc":
		merged := map[string]json.RawMessage{
			"stats":      results["stats"],
			"topDevices": results["topDevices"],
			"topTypes":   results["topTypes"],
		}
		b, _ := json.Marshal(merged)
		fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(b, f.IO, output))
	case "yaml":
		merged := map[string]json.RawMessage{
			"stats":      results["stats"],
			"topDevices": results["topDevices"],
			"topTypes":   results["topTypes"],
		}
		b, _ := json.Marshal(merged)
		s, err := iostreams.FormatYAML(b)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IO.Out, s)
	default:
		printAlertsDashboard(f.IO, results, opts.Fields)
	}

	return nil
}

func printAlertsDashboard(io *iostreams.IOStreams, data map[string]json.RawMessage, fields []string) {
	c := iostreams.NewColorizer(io.TermOutput())
	out := io.Out

	// --- Alert Summary ---
	fmt.Fprintln(out, c.Bold("Alert Summary"))
	var stats struct {
		Active int `json:"active"`
		Closed int `json:"closed"`
		Total  int `json:"total"`
	}
	if json.Unmarshal(data["stats"], &stats) == nil {
		fmt.Fprintf(out, "  Active: %s  Closed: %s  Total: %s\n",
			c.Red(strconv.Itoa(stats.Active)),
			c.Green(strconv.Itoa(stats.Closed)),
			c.Bold(strconv.Itoa(stats.Total)),
		)
	}
	fmt.Fprintln(out)

	// --- Top Devices by Alert Count ---
	fmt.Fprintln(out, c.Bold("Top Devices by Alert Count"))
	var topDevices []map[string]interface{}
	if json.Unmarshal(data["topDevices"], &topDevices) == nil && len(topDevices) > 0 {
		devFields := fields
		if len(devFields) == 0 {
			devFields = defaultTopAlertDevicesFields
		}
		tp := iostreams.NewTablePrinter(out, io.IsStdoutTTY())
		// Header
		headers := make([]string, len(devFields))
		for i, f := range devFields {
			headers[i] = c.Bold(f)
		}
		tp.AddRow(headers...)
		// Rows
		for _, d := range topDevices {
			row := make([]string, len(devFields))
			for i, f := range devFields {
				if v, ok := d[f]; ok {
					row[i] = formatValue(v)
				} else {
					row[i] = ""
				}
			}
			tp.AddRow(row...)
		}
		_ = tp.Render()
	} else {
		fmt.Fprintln(out, c.Gray("  No data"))
	}
	fmt.Fprintln(out)

	// --- Top Alert Types ---
	fmt.Fprintln(out, c.Bold("Top Alert Types"))
	var topTypes []map[string]interface{}
	if json.Unmarshal(data["topTypes"], &topTypes) == nil && len(topTypes) > 0 {
		typeFields := fields
		if len(typeFields) == 0 {
			typeFields = defaultTopAlertTypesFields
		}
		tp := iostreams.NewTablePrinter(out, io.IsStdoutTTY())
		// Header
		headers := make([]string, len(typeFields))
		for i, f := range typeFields {
			headers[i] = c.Bold(f)
		}
		tp.AddRow(headers...)
		// Rows
		for _, t := range topTypes {
			row := make([]string, len(typeFields))
			for i, f := range typeFields {
				if v, ok := t[f]; ok {
					row[i] = formatValue(v)
				} else {
					row[i] = ""
				}
			}
			tp.AddRow(row...)
		}
		_ = tp.Render()
	} else {
		fmt.Fprintln(out, c.Gray("  No data"))
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}
