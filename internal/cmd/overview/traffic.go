package overview

import (
	"encoding/json"
	"fmt"
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

var defaultTrafficFields = []string{"deviceName", "serialNumber", "tx", "rx", "total"}

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

  # JSON output
  incloud overview traffic -o json

  # Table output with selected fields
  incloud overview traffic -o table -f deviceName -f total`,
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

	overviewURL := buildURL(host+"/api/v1/datausage/overview", map[string]string{
		"after":  opts.After,
		"before": opts.Before,
	})

	topkParams := map[string]string{
		"n":      strconv.Itoa(opts.N),
		"after":  truncateToDate(opts.After),
		"before": truncateToDate(opts.Before),
		"type":   opts.Type,
	}
	topkURL := buildURL(host+"/api/v1/datausage/topk", topkParams)

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]json.RawMessage)
		apiErr  error
	)

	apis := []struct {
		name string
		url  string
	}{
		{"overview", overviewURL},
		{"topk", topkURL},
	}

	for _, a := range apis {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			body, err := doGet(client, url)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if apiErr == nil {
					apiErr = fmt.Errorf("%s: %w", name, err)
				}
				return
			}
			var envelope struct {
				Result json.RawMessage `json:"result"`
			}
			if json.Unmarshal(body, &envelope) == nil && envelope.Result != nil {
				results[name] = envelope.Result
			} else {
				results[name] = body
			}
		}(a.name, a.url)
	}
	wg.Wait()

	if apiErr != nil {
		return apiErr
	}

	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json", "jsonc":
		merged := map[string]json.RawMessage{
			"overview":   results["overview"],
			"topDevices": results["topk"],
		}
		b, _ := json.Marshal(merged)
		fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(b, f.IO, output))

	case "yaml":
		merged := map[string]json.RawMessage{
			"overview":   results["overview"],
			"topDevices": results["topk"],
		}
		b, _ := json.Marshal(merged)
		s, err := iostreams.FormatYAML(b)
		if err != nil {
			return err
		}
		fmt.Fprintln(f.IO.Out, s)

	case "table":
		fields := opts.Fields
		if len(fields) == 0 {
			fields = defaultTrafficFields
		}
		if err := iostreams.FormatTable(results["topk"], f.IO, fields); err != nil {
			return err
		}

	default:
		printTrafficDashboard(f.IO, results, opts.Fields)
	}

	return nil
}

func printTrafficDashboard(io *iostreams.IOStreams, data map[string]json.RawMessage, fields []string) {
	c := iostreams.NewColorizer(io.TermOutput())
	out := io.Out

	// --- Traffic Summary ---
	fmt.Fprintln(out, c.Bold("Traffic Summary"))
	var trafficData struct {
		Series []struct {
			Type   string          `json:"type"`
			Fields []string        `json:"fields"`
			Data   [][]interface{} `json:"data"`
		} `json:"series"`
	}
	if json.Unmarshal(data["overview"], &trafficData) == nil && len(trafficData.Series) > 0 {
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

	// --- Top Devices by Data Usage ---
	fmt.Fprintln(out, c.Bold("Top Devices by Data Usage"))
	var topDevices []struct {
		DeviceName   string  `json:"deviceName"`
		SerialNumber string  `json:"serialNumber"`
		TX           float64 `json:"tx"`
		RX           float64 `json:"rx"`
		Total        float64 `json:"total"`
	}
	if json.Unmarshal(data["topk"], &topDevices) == nil && len(topDevices) > 0 {
		tp := iostreams.NewTablePrinter(out, io.IsStdoutTTY())
		tp.AddRow(c.Bold("DEVICE"), c.Bold("SERIAL"), c.Bold("TX"), c.Bold("RX"), c.Bold("TOTAL"))
		for _, d := range topDevices {
			name := d.DeviceName
			if name == "" {
				name = d.SerialNumber
			}
			tp.AddRow(name, d.SerialNumber, formatBytes(d.TX), formatBytes(d.RX), formatBytes(d.Total))
		}
		_ = tp.Render()
	} else {
		fmt.Fprintln(out, c.Gray("  No top device data"))
	}
}
