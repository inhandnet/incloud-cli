package device

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigHistoryList(f *factory.Factory) *cobra.Command {
	var (
		module string
		page   int
		limit  int
		sort   string
		after  string
		before string
		fields []string
	)

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

			q := url.Values{}
			q.Set("page", strconv.Itoa(page-1))
			q.Set("limit", strconv.Itoa(limit))
			if module != "" {
				q.Set("module", module)
			}
			if sort != "" {
				q.Set("sort", sort)
			}
			if after != "" {
				q.Set("after", after)
			}
			if before != "" {
				q.Set("before", before)
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

	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&after, "after", "", "Filter history after this time (ISO 8601)")
	cmd.Flags().StringVar(&before, "before", "", "Filter history before this time (ISO 8601)")
	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to display in table mode")

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
