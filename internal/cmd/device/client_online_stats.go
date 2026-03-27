package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultClientOnlineStatsListFields = []string{"eventType", "mode", "networkName", "timestamp", "online"}

func newCmdClientOnlineStats(f *factory.Factory) *cobra.Command {
	var after, before string

	cmd := &cobra.Command{
		Use:   "online-stats <client-id>",
		Short: "Client online statistics",
		Long:  "Display online time, offline count, online rate, and event timeline for a client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if after != "" {
				q.Set("after", after)
			}
			if before != "" {
				q.Set("before", before)
			}

			body, err := client.Get("/api/v1/network/clients/"+args[0]+"/online-events-chart/statistics", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			return renderOnlineStatsTable(body, f.IO)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start time (ISO 8601)")
	cmd.Flags().StringVar(&before, "before", "", "End time (ISO 8601)")

	return cmd
}

func renderOnlineStatsTable(body []byte, io *iostreams.IOStreams) error {
	parsed := gjson.ParseBytes(body)
	result := parsed.Get("result")

	// Build summary object (exclude list field)
	summary := map[string]interface{}{
		"onlineTime":   result.Get("onlineTime").Value(),
		"offlineCount": result.Get("offlineCount").Value(),
		"onlineRate":   result.Get("onlineRate").Value(),
	}
	summaryJSON, err := json.Marshal(map[string]interface{}{"result": summary})
	if err != nil {
		return err
	}
	if err := iostreams.FormatTable(summaryJSON, io, nil); err != nil {
		return err
	}

	// Print events section header
	c := iostreams.NewColorizer(io.TermOutput())
	fmt.Fprintln(io.Out)
	if io.IsStdoutTTY() {
		fmt.Fprintln(io.Out, c.Bold("Events"))
	} else {
		fmt.Fprintln(io.Out, "Events")
	}

	// Build list array
	listItems := result.Get("list").Array()
	if len(listItems) == 0 {
		fmt.Fprintln(io.Out, "No events.")
		return nil
	}

	listJSON, err := json.Marshal(map[string]interface{}{"result": result.Get("list").Value()})
	if err != nil {
		return err
	}
	return iostreams.FormatTable(listJSON, io, defaultClientOnlineStatsListFields)
}
