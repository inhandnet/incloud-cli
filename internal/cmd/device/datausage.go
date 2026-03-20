package device

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdDatausage(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "datausage",
		Aliases: []string{"du"},
		Short:   "Device data usage statistics",
		Long:    "View device data usage (traffic) statistics at hourly, daily, or monthly granularity.",
	}

	cmd.AddCommand(newCmdDatausageHourly(f))
	cmd.AddCommand(newCmdDatausageDaily(f))
	cmd.AddCommand(newCmdDatausageMonthly(f))
	cmd.AddCommand(newCmdDatausageList(f))

	return cmd
}

var defaultDatausageFields = []string{"time", "type", "tx", "rx", "total"}

// datausageSeriesOptions holds flags shared by hourly/daily/monthly subcommands.
// Each subcommand only registers the flags it needs (e.g. daily registers --month
// but not --after/--before).
type datausageSeriesOptions struct {
	Type   string
	After  string
	Before string
	Month  string
	Year   string
	Fields []string
}

// fetchDatausageSeries is the shared RunE for hourly/daily/monthly commands.
// endpoint is the URL path suffix (e.g. "datausage-hourly").
func fetchDatausageSeries(f *factory.Factory, cmd *cobra.Command, deviceID, endpoint string, opts *datausageSeriesOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body, err := client.Get("/api/v1/devices/"+deviceID+"/"+endpoint, url.Values{
		"type":   {opts.Type},
		"after":  {opts.After},
		"before": {opts.Before},
		"month":  {opts.Month},
		"year":   {opts.Year},
	})
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	if !cmd.Flags().Changed("output") {
		output = "table"
	}
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultDatausageFields
	}
	return iostreams.FormatOutput(body, f.IO, output, fields,
		iostreams.WithTransform(iostreams.FlattenSeries),
		iostreams.WithFormatters(iostreams.ColumnFormatters{
			"tx":    iostreams.FormatBytes,
			"rx":    iostreams.FormatBytes,
			"total": iostreams.FormatBytes,
		}),
	)
}
