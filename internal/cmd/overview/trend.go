package overview

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TrendOptions struct {
	After  string
	Before string
	Fields []string
}

var defaultTrendFields = []string{"date", "online", "total"}

func NewCmdTrend(f *factory.Factory) *cobra.Command {
	opts := &TrendOptions{}

	cmd := &cobra.Command{
		Use:   "trend",
		Short: "Device online count trend",
		Long:  "Show daily online and total device count trend over time.",
		Example: `  # Show trend for the last 30 days (default)
  incloud overview trend

  # Custom time range
  incloud overview trend --after 2024-01-01 --before 2024-03-31

  # JSON output
  incloud overview trend -o json

  # Show only online count
  incloud overview trend -f date -f online`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrend(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start date (e.g. 2024-01-01 or 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End date (e.g. 2024-03-31 or 2024-03-31T23:59:59Z)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}

func runTrend(cmd *cobra.Command, f *factory.Factory, opts *TrendOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	applyDefaultTrendRange(&opts.After, &opts.Before)

	// Fetch both metrics concurrently
	var onlineBody, totalBody []byte
	var onlineErr, totalErr error
	var wg sync.WaitGroup

	q := makeQuery(map[string]string{
		"after":      opts.After,
		"before":     opts.Before,
		"aggregator": "LAST",
		"interval":   "1d",
		"fill":       "zero",
	})

	wg.Add(2)
	go func() {
		defer wg.Done()
		onlineBody, onlineErr = client.Get("/api/v1/stats/nezha_daily_global_devices_online_count/data", q)
	}()
	go func() {
		defer wg.Done()
		totalBody, totalErr = client.Get("/api/v1/stats/nezha_daily_global_devices_total_count/data", q)
	}()
	wg.Wait()

	if onlineErr != nil {
		return fmt.Errorf("fetching online count: %w", onlineErr)
	}
	if totalErr != nil {
		return fmt.Errorf("fetching total count: %w", totalErr)
	}

	merged, err := mergeStatsSeries(onlineBody, totalBody)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	if !cmd.Flags().Changed("output") {
		output = "table"
	}
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultTrendFields
	}
	return iostreams.FormatOutput(merged, f.IO, output, fields)
}

// applyDefaultTrendRange sets after/before to the last 30 days when not specified.
func applyDefaultTrendRange(after, before *string) {
	now := time.Now()
	if *before == "" {
		*before = now.Format(time.RFC3339)
	}
	if *after == "" {
		*after = now.AddDate(0, 0, -30).Format(time.RFC3339)
	}
}

// mergeStatsSeries merges two stats API responses (online + total) into a flat
// JSON array with columns: date, online, total.
//
// Stats API response format:
//
//	{"result":{"series":[{"name":"...","values":[{"timestamp":"...","value":N}]}]}}
func mergeStatsSeries(onlineBody, totalBody []byte) ([]byte, error) {
	onlineMap, err := parseStatsValues(onlineBody)
	if err != nil {
		return nil, fmt.Errorf("parsing online series: %w", err)
	}
	totalMap, err := parseStatsValues(totalBody)
	if err != nil {
		return nil, fmt.Errorf("parsing total series: %w", err)
	}

	// Collect all dates
	dateSet := make(map[string]struct{})
	for d := range onlineMap {
		dateSet[d] = struct{}{}
	}
	for d := range totalMap {
		dateSet[d] = struct{}{}
	}

	// Sort dates
	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	// Build rows
	rows := make([]any, 0, len(dates))
	for _, d := range dates {
		rows = append(rows, map[string]any{
			"date":   d,
			"online": onlineMap[d],
			"total":  totalMap[d],
		})
	}

	return json.Marshal(map[string]any{"result": rows})
}

// parseStatsValues extracts timestamp->value map from a stats API response.
// Timestamps are truncated to date (YYYY-MM-DD).
func parseStatsValues(body []byte) (map[string]float64, error) {
	var resp struct {
		Result struct {
			Series []struct {
				Values []struct {
					Timestamp string  `json:"timestamp"`
					Value     float64 `json:"value"`
				} `json:"values"`
			} `json:"series"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	m := make(map[string]float64)
	for _, s := range resp.Result.Series {
		for _, v := range s.Values {
			date := truncateToDate(v.Timestamp)
			m[date] = v.Value
		}
	}
	return m, nil
}

