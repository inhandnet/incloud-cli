package connector

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultAccountFields = []string{"_id", "name", "vip", "staticIp", "connected", "createdAt"}

type accountListOptions struct {
	Page      int
	Limit     int
	Sort      string
	Name      string
	Connected string
	Search    string
	Fields    []string
}

func newCmdAccountList(f *factory.Factory) *cobra.Command {
	opts := &accountListOptions{}

	cmd := &cobra.Command{
		Use:     "list <network-id>",
		Aliases: []string{"ls"},
		Short:   "List accounts in a connector network",
		Example: `  # List accounts in a network
  incloud connector account list 66827b3ccfb1842140f4222f

  # Filter connected only
  incloud connector account list 66827b3ccfb1842140f4222f --connected true

  # Search by name
  incloud connector account list 66827b3ccfb1842140f4222f -q admin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

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
				q.Set("name", opts.Name)
			}
			if opts.Connected != "" {
				q.Set("connected", opts.Connected)
			}
			if opts.Search != "" {
				q.Set("nameOrSn", opts.Search)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultAccountFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/connectors/"+networkID+"/accounts", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by account name")
	cmd.Flags().StringVar(&opts.Connected, "connected", "", "Filter by connected status (true/false)")
	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
