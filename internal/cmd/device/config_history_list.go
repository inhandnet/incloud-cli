package device

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type configHistoryListOptions struct {
	cmdutil.ListFlags
	Module string
	After  string
	Before string
}

func newCmdConfigHistoryList(f *factory.Factory) *cobra.Command {
	opts := &configHistoryListOptions{}

	cmd := &cobra.Command{
		Use:   "list <device-id>",
		Short: "List configuration change history",
		Long: `List configuration change history snapshots for a device, with pagination and time range filtering.

The mergedConfig field is omitted by default to keep output concise.
Use 'incloud device config snapshots get' to view the full snapshot including merged configuration.`,
		Example: `  # List recent config history
  incloud device config snapshots list 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device config snapshots list 507f1f77bcf86cd799439011 --after 2024-01-01 --before 2024-02-01

  # Paginate
  incloud device config snapshots list 507f1f77bcf86cd799439011 --page 2 --limit 10`,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Module != "" {
				q.Set("module", opts.Module)
			}
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/config/history", q)
			if err != nil {
				return err
			}

			body = stripMergedConfig(body)

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "table"
			}
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Filter history after this time (ISO 8601)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter history before this time (ISO 8601)")

	return cmd
}

// stripMergedConfig removes the mergedConfig field from each record in the
// paginated response to reduce output size. Returns body unchanged on error.
func stripMergedConfig(body []byte) []byte {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		return body
	}
	raw, ok := envelope["result"]
	if !ok {
		return body
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return body
	}
	for _, item := range items {
		delete(item, "mergedConfig")
	}
	stripped, err := json.Marshal(items)
	if err != nil {
		return body
	}
	envelope["result"] = stripped
	out, err := json.Marshal(envelope)
	if err != nil {
		return body
	}
	return out
}
