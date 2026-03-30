package overview

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OfflineOptions struct {
	After                          string
	Before                         string
	Group                          []string
	N                              int
	Page                           int
	Limit                          int
	Fields                         []string
	Query                          string
	OfflineTimesGreaterThan        int
	MaxOfflineTimesGreaterThan     int
	MaxOfflineDurationGreaterThan  int
}

var (
	offlineFormatters = iostreams.ColumnFormatters{
		"totalOfflineDuration": iostreams.FormatDuration,
		"avgOfflineDuration":   iostreams.FormatDuration,
		"maxOfflineDuration":   iostreams.FormatDuration,
	}
)

func NewCmdOffline(f *factory.Factory) *cobra.Command {
	opts := &OfflineOptions{}

	cmd := &cobra.Command{
		Use:   "offline",
		Short: "Offline analysis and top devices",
		Long:  "Show top-N offline devices and offline statistics list.",
		Example: `  # Show offline dashboard
  incloud overview offline

  # Custom time range
  incloud overview offline --after 2024-01-01 --before 2024-01-31

  # Top 5 devices, page 2 of statistics
  incloud overview offline --n 5 --page 2

  # Filter by device group
  incloud overview offline --group 507f1f77bcf86cd799439011

  # JSON output
  incloud overview offline -o json

  # Table output (statistics only)
  incloud overview offline -o table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOffline(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2024-01-01 or 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2024-01-31 or 2024-01-31T23:59:59Z)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().IntVar(&opts.N, "n", 10, "Number of top devices to show")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Statistics list page number (1-based)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Statistics list page size")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Filter by device name or serial number")
	cmd.Flags().IntVar(&opts.OfflineTimesGreaterThan, "min-offline-times", 0, "Filter devices with total offline times >= N")
	cmd.Flags().IntVar(&opts.MaxOfflineTimesGreaterThan, "min-max-offline-times", 0, "Filter devices with daily max offline times >= N")
	cmd.Flags().IntVar(&opts.MaxOfflineDurationGreaterThan, "min-max-offline-duration", 0, "Filter devices with daily max offline duration >= N seconds")

	return cmd
}

func runOffline(cmd *cobra.Command, f *factory.Factory, opts *OfflineOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	applyDefaultTimeRange2(&opts.After, &opts.Before)

	// Build topn query
	topnQuery := makeQueryWithGroups(map[string]string{
		"topN":   strconv.Itoa(opts.N),
		"after":  opts.After,
		"before": opts.Before,
	}, opts.Group)

	// Build statistics query (page is 1-based in CLI, 0-based in API)
	statsParams := map[string]string{
		"page":   strconv.Itoa(opts.Page - 1),
		"limit":  strconv.Itoa(opts.Limit),
		"after":  opts.After,
		"before": opts.Before,
	}
	if opts.Query != "" {
		statsParams["q"] = opts.Query
	}
	if opts.OfflineTimesGreaterThan > 0 {
		statsParams["offlineTimesGreaterThan"] = strconv.Itoa(opts.OfflineTimesGreaterThan)
	}
	if opts.MaxOfflineTimesGreaterThan > 0 {
		statsParams["dailyMaxOfflineTimesGreaterThan"] = strconv.Itoa(opts.MaxOfflineTimesGreaterThan)
	}
	if opts.MaxOfflineDurationGreaterThan > 0 {
		statsParams["dailyMaxOfflineDurationGreaterThan"] = strconv.Itoa(opts.MaxOfflineDurationGreaterThan)
	}
	statsQuery := makeQueryWithGroups(statsParams, opts.Group)

	// Concurrent fetch
	var topnBody, statsBody []byte
	var topnErr, statsErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		topnBody, topnErr = client.Get("/api/v1/devices/offline/topn", topnQuery)
	}()
	go func() {
		defer wg.Done()
		statsBody, statsErr = client.Get("/api/v1/devices/offline/statistics", statsQuery)
	}()
	wg.Wait()

	if topnErr != nil {
		return fmt.Errorf("topn: %w", topnErr)
	}
	if statsErr != nil {
		return fmt.Errorf("statistics: %w", statsErr)
	}

	// Unwrap topn result
	topnData := unwrapResult(topnBody)

	output, _ := cmd.Flags().GetString("output")

	switch output {
	case "json", "jsonc", "yaml":
		merged := buildMergedOutput(topnData, statsBody)
		b, _ := json.Marshal(merged)
		return iostreams.FormatOutput(b, f.IO, output)
	default:
		printOfflineDashboard(f.IO, opts, topnData, statsBody)
	}

	return nil
}

func buildMergedOutput(topnData json.RawMessage, statsBody []byte) map[string]json.RawMessage {
	merged := make(map[string]json.RawMessage)
	merged["topDevices"] = topnData

	// Parse statistics to build the structured output
	var statsEnvelope struct {
		Result json.RawMessage `json:"result"`
		Total  json.RawMessage `json:"total"`
		Page   json.RawMessage `json:"page"`
		Limit  json.RawMessage `json:"limit"`
	}
	if json.Unmarshal(statsBody, &statsEnvelope) == nil && statsEnvelope.Result != nil {
		statsObj := map[string]json.RawMessage{
			"result": statsEnvelope.Result,
			"total":  statsEnvelope.Total,
			"page":   statsEnvelope.Page,
			"limit":  statsEnvelope.Limit,
		}
		b, _ := json.Marshal(statsObj)
		merged["statistics"] = b
	} else {
		merged["statistics"] = statsBody
	}

	return merged
}

func printOfflineDashboard(streams *iostreams.IOStreams, opts *OfflineOptions, topnData json.RawMessage, statsBody []byte) {
	c := iostreams.NewColorizer(streams.TermOutput())
	out := streams.Out

	fmt.Fprintln(out, c.Gray(fmt.Sprintf("Period: %s ~ %s", opts.After, opts.Before)))
	fmt.Fprintln(out)

	// --- Top Offline Devices ---
	fmt.Fprintln(out, c.Bold("Top Offline Devices"))
	topWrapper, _ := json.Marshal(map[string]json.RawMessage{"result": topnData})
	if err := iostreams.FormatOutput(topWrapper, streams, "table"); err != nil {
		fmt.Fprintln(out, c.Gray("  No data"))
	}
	fmt.Fprintln(out)

	// --- Offline Statistics ---
	fmt.Fprintln(out, c.Bold("Offline Statistics"))
	if err := iostreams.FormatOutput(statsBody, streams, "table",
		iostreams.WithFormatters(offlineFormatters),
	); err != nil {
		fmt.Fprintln(out, c.Gray("  No data"))
	}
}
