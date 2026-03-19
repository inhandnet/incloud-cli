package sdwan

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultConnectionFields = []string{"_id", "source", "target", "status"}

type networkConnectionsOptions struct {
	Page   int
	Limit  int
	Sort   string
	Name   string
	Fields []string
}

func newCmdNetworkConnections(f *factory.Factory) *cobra.Command {
	opts := &networkConnectionsOptions{}

	cmd := &cobra.Command{
		Use:   "connections <networkId>",
		Short: "List connections in an SD-WAN network",
		Example: `  # List all connections
  incloud sdwan network connections <id>

  # Filter by device name
  incloud sdwan network connections <id> --name ER805`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Name != "" {
				q.Set("deviceName", opts.Name)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultConnectionFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/connections", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
