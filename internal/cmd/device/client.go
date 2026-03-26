package device

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdClient(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Connected clients (Wi-Fi/LAN)",
		Long:  "List, inspect, and monitor clients connected to your devices.",
	}

	cmd.AddCommand(newCmdClientList(f))
	cmd.AddCommand(newCmdClientGet(f))
	cmd.AddCommand(newCmdClientUpdate(f))
	cmd.AddCommand(newCmdClientThroughput(f))
	cmd.AddCommand(newCmdClientRSSI(f))
	cmd.AddCommand(newCmdClientSINR(f))
	cmd.AddCommand(newCmdClientDatausageHourly(f))
	cmd.AddCommand(newCmdClientDatausageDaily(f))
	cmd.AddCommand(newCmdClientOnlineEvents(f))
	cmd.AddCommand(newCmdClientOnlineStats(f))
	cmd.AddCommand(newCmdClientMarkAsset(f))

	return cmd
}

// clientSeriesOptions holds flags shared by time-series subcommands (throughput,
// rssi, sinr, datausage-hourly, datausage-daily).
type clientSeriesOptions struct {
	After  string
	Before string
	Month  string
	Fields []string
}

// fetchClientSeries is the shared RunE for client time-series commands.
// endpoint is the URL path suffix (e.g. "throughput", "rssi").
func fetchClientSeries(f *factory.Factory, cmd *cobra.Command, clientID, endpoint string, opts *clientSeriesOptions, defaultFields []string, fmts ...iostreams.FormatOption) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	if opts.After != "" {
		q.Set("after", opts.After)
	}
	if opts.Before != "" {
		q.Set("before", opts.Before)
	}
	if opts.Month != "" {
		q.Set("month", opts.Month)
	}

	body, err := client.Get("/api/v1/network/clients/"+clientID+"/"+endpoint, q)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	if !cmd.Flags().Changed("output") {
		output = "table"
	}
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultFields
	}

	formatOpts := []iostreams.FormatOption{iostreams.WithTransform(iostreams.FlattenSeries)}
	formatOpts = append(formatOpts, fmts...)
	return iostreams.FormatOutput(body, f.IO, output, fields, formatOpts...)
}
