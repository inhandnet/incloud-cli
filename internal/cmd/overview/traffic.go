package overview

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TrafficOptions struct {
	After  string
	Before string
	Type   string
	N      int
	Fields []string
}

var trafficFormatters = iostreams.ColumnFormatters{
	"tx":    iostreams.FormatBytes,
	"rx":    iostreams.FormatBytes,
	"total": iostreams.FormatBytes,
}

func NewCmdTraffic(f *factory.Factory) *cobra.Command {
	opts := &TrafficOptions{}

	cmd := &cobra.Command{
		Use:   "traffic",
		Short: "Traffic overview and top devices",
		Long:  "Show global traffic statistics and top-K devices by data usage.",
		Example: `  # Show traffic overview
  incloud overview traffic

  # Filter by time range
  incloud overview traffic --after 2024-01-01 --before 2024-01-31

  # Filter by traffic type
  incloud overview traffic --type cellular

  # Top 5 devices
  incloud overview traffic --n 5

  # JSON output (summary + trend + topDevices)
  incloud overview traffic -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTraffic(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start date (e.g. 2024-01-01)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End date (e.g. 2024-01-31)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Traffic type: cellular|wifi|wired|all")
	cmd.Flags().IntVar(&opts.N, "n", 10, "Number of top devices to return")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table output")

	return cmd
}

// truncateToDate extracts the date portion from a string that may contain
// an ISO 8601 timestamp (e.g. "2024-01-01T00:00:00" -> "2024-01-01").
func truncateToDate(s string) string {
	if idx := strings.IndexByte(s, 'T'); idx > 0 {
		return s[:idx]
	}
	return s
}

func runTraffic(cmd *cobra.Command, f *factory.Factory, opts *TrafficOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	overviewQuery := makeQuery(map[string]string{
		"after":  opts.After,
		"before": opts.Before,
	})

	topkQuery := makeQuery(map[string]string{
		"n":      strconv.Itoa(opts.N),
		"after":  truncateToDate(opts.After),
		"before": truncateToDate(opts.Before),
		"type":   opts.Type,
	})

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]json.RawMessage)
		apiErr  error
	)

	apis := []struct {
		name  string
		path  string
		query url.Values
	}{
		{"overview", "/api/v1/datausage/overview", overviewQuery},
		{"topk", "/api/v1/datausage/topk", topkQuery},
	}

	for _, a := range apis {
		wg.Add(1)
		go func(name, path string, query url.Values) {
			defer wg.Done()
			body, err := client.Get(path, query)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if apiErr == nil {
					apiErr = fmt.Errorf("%s: %w", name, err)
				}
				return
			}
			results[name] = unwrapResult(body)
		}(a.name, a.path, a.query)
	}
	wg.Wait()

	if apiErr != nil {
		return apiErr
	}

	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json", "jsonc", "yaml":
		summary := buildTrafficSummary(results["overview"])
		trend := buildTrafficTrend(results["overview"])
		merged := map[string]json.RawMessage{
			"summary":    mustMarshalRaw(summary),
			"trend":      mustMarshalRaw(trend),
			"topDevices": results["topk"],
		}
		b, _ := json.Marshal(merged)
		return iostreams.FormatOutput(b, f.IO, output)

	default: // table and human-readable
		return printTrafficDashboard(f.IO, results)
	}
}

// overviewSeries is the parsed form of /api/v1/datausage/overview result.
type overviewSeries struct {
	Series []struct {
		Type   string          `json:"type"`
		Fields []string        `json:"fields"`
		Data   [][]interface{} `json:"data"`
	} `json:"series"`
}

func parseOverviewSeries(raw json.RawMessage) *overviewSeries {
	var s overviewSeries
	if json.Unmarshal(raw, &s) != nil {
		return nil
	}
	return &s
}

// buildTrafficSummary aggregates each series into a single row per traffic type.
func buildTrafficSummary(raw json.RawMessage) []map[string]any {
	s := parseOverviewSeries(raw)
	if s == nil {
		return nil
	}
	rows := make([]map[string]any, 0, len(s.Series))
	for _, series := range s.Series {
		txIdx := fieldIndex(series.Fields, "tx")
		rxIdx := fieldIndex(series.Fields, "rx")
		totalIdx := fieldIndex(series.Fields, "total")
		var txSum, rxSum, totalSum float64
		for _, row := range series.Data {
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
		rows = append(rows, map[string]any{
			"type":  series.Type,
			"tx":    txSum,
			"rx":    rxSum,
			"total": totalSum,
		})
	}
	return rows
}

// buildTrafficTrend flattens series data into one row per (type, timestamp).
func buildTrafficTrend(raw json.RawMessage) []map[string]any {
	s := parseOverviewSeries(raw)
	if s == nil {
		return nil
	}
	var rows []map[string]any
	for _, series := range s.Series {
		for _, row := range series.Data {
			obj := map[string]any{"type": series.Type}
			for i, field := range series.Fields {
				if i < len(row) {
					obj[field] = row[i]
				}
			}
			rows = append(rows, obj)
		}
	}
	return rows
}

func mustMarshalRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func printTrafficDashboard(io *iostreams.IOStreams, data map[string]json.RawMessage) error {
	c := iostreams.NewColorizer(io.TermOutput())
	out := io.Out

	summary := buildTrafficSummary(data["overview"])
	trend := buildTrafficTrend(data["overview"])

	// --- Traffic Summary ---
	fmt.Fprintln(out, c.Bold("Traffic Summary"))
	if len(summary) > 0 {
		for _, row := range summary {
			fmt.Fprintf(out, "  %s — TX: %s  RX: %s  Total: %s\n",
				row["type"],
				c.Bold(formatBytes(toFloat(row["tx"]))),
				c.Bold(formatBytes(toFloat(row["rx"]))),
				c.Bold(formatBytes(toFloat(row["total"]))),
			)
		}
	} else {
		fmt.Fprintln(out, c.Gray("  No traffic data"))
	}
	fmt.Fprintln(out)

	// --- Traffic Trend ---
	fmt.Fprintln(out, c.Bold("Traffic Trend"))
	if len(trend) > 0 {
		wrapped, _ := json.Marshal(map[string]any{"result": trend})
		if err := iostreams.FormatOutput(wrapped, io, "table",
			iostreams.WithColumns("type", "time", "tx", "rx", "total"),
			iostreams.WithFormatters(trafficFormatters),
		); err != nil {
			fmt.Fprintln(out, c.Gray("  No trend data"))
		}
	} else {
		fmt.Fprintln(out, c.Gray("  No trend data"))
	}
	fmt.Fprintln(out)

	// --- Top Devices by Data Usage ---
	fmt.Fprintln(out, c.Bold("Top Devices by Data Usage"))
	wrapped := []byte(`{"result":` + string(data["topk"]) + `}`)
	if err := iostreams.FormatOutput(wrapped, io, "table",
		iostreams.WithFormatters(trafficFormatters),
	); err != nil {
		fmt.Fprintln(out, c.Gray("  No top device data"))
	}

	return nil
}
