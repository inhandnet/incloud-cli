package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type signalListOptions struct {
	After  string
	Before string
	Order  string
	Fields []string
}

var defaultSignalFields = []string{"time", "type", "rsrp", "rsrq", "sinr", "networkType", "carrier", "band"}

func newCmdSignalList(f *factory.Factory) *cobra.Command {
	opts := &signalListOptions{}

	cmd := &cobra.Command{
		Use:   "list <device-id>",
		Short: "Show signal quality over time",
		Long:  "Display signal quality metrics (RSRP, RSRQ, SINR, etc.) for a device over time.",
		Example: `  # Show signal data for a device
  incloud device signal list 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device signal list 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z

  # Select specific fields
  incloud device signal list 507f1f77bcf86cd799439011 -f time -f rsrp -f sinr

  # JSON output
  incloud device signal list 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Order != "asc" && opts.Order != "desc" {
				return fmt.Errorf("invalid --order value %q: must be asc or desc", opts.Order)
			}
			deviceID := args[0]

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

			body, err := client.Get("/api/v1/devices/"+deviceID+"/signal", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if !cmd.Flags().Changed("output") {
				output = "table"
			}
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultSignalFields
			}
			transform := iostreams.TransformFunc(iostreams.FlattenSeries)
			if opts.Order == "desc" {
				transform = iostreams.ChainTransforms(iostreams.FlattenSeries, iostreams.ReverseJSONArray)
			}
			return iostreams.FormatOutput(body, f.IO, output, iostreams.WithTransform(transform))
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Order, "order", "desc", "Sort order: asc (oldest first) or desc (newest first)")
	_ = cmd.RegisterFlagCompletionFunc("order", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"asc", "desc"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
