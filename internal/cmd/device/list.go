package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page    int
	Limit   int
	Sort    string
	Query   string
	Online  string
	Status  string
	Product []string
	Group   []string
	Fields  []string
	Expand  []string
}

var defaultListFields = []string{"_id", "name", "serialNumber", "online", "product", "firmware"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List devices",
		Long:    "List devices on the InCloud platform with optional filtering, searching, and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List devices with default pagination
  incloud device list

  # Paginate
  incloud device list --page 2 --limit 50

  # Filter by online status
  incloud device list --online true

  # Search by name or serial number
  incloud device list -q "router"

  # Filter by product
  incloud device list --product IR615

  # Sort results
  incloud device list --sort "name,asc"

  # Expand related resources (e.g. org info, firmware upgrade status)
  incloud device list --expand org,firmwareUpgradeStatus

  # Table output with selected fields
  incloud device list -o table -f name -f serialNumber -f online

  # Extract names with jq
  incloud device list --jq '.result[].name'

  # Export offline devices as CSV
  incloud device list --online false --jq '.result[] | [.name, .serialNumber] | @csv'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Query != "" {
				q.Set("q", opts.Query)
			}
			if opts.Online != "" {
				q.Set("online", opts.Online)
			}
			if opts.Status != "" {
				switch opts.Status {
				case "online":
					q.Set("online", "true")
				case "offline":
					q.Set("online", "false")
				}
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}
			for _, g := range opts.Group {
				q.Add("devicegroupId", g)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/devices", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort order (e.g. \"createdAt,desc\")")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by name or serial number")
	cmd.Flags().StringVar(&opts.Online, "online", "", "Filter by online status (true/false)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (online/offline)")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related resources (e.g. org,firmwareUpgradeStatus,compatibilities)")

	return cmd
}
