package device

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultConfigHistoryFields = []string{"_id", "trigger", "version", "createdAt"}

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
		Long:  "List configuration change history snapshots for a device, with pagination and time range filtering.",
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
			q.Set("size", strconv.Itoa(limit))
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

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "table"
			}
			displayFields := fields
			if len(displayFields) == 0 {
				// Always use default fields in table mode to avoid
				// expanding the huge mergedConfig nested object.
				displayFields = defaultConfigHistoryFields
			}

			return iostreams.FormatOutput(body, f.IO, output, displayFields)
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
