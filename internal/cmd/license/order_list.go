package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OrderListOptions struct {
	cmdutil.ListFlags
	Status string
	Type   string
	After  string
	Before string
}

var defaultOrderListFields = []string{"_id", "type", "status", "totalAmount", "currency", "createdAt", "paidAt"}

func NewCmdOrderList(f *factory.Factory) *cobra.Command {
	opts := &OrderListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List orders",
		Long:    "List license orders with optional filtering by status, type, and date range.",
		Aliases: []string{"ls"},
		Example: `  # List completed orders
  incloud license order list --status complete

  # List renewal orders in a date range
  incloud license order list --type license_renewal --after 2026-01-01 --before 2026-03-31

  # List all orders as YAML
  incloud license order list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultOrderListFields)
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if opts.After != "" {
				q.Set("after", cmdutil.ParseTimeFlag(opts.After))
			}
			if opts.Before != "" {
				q.Set("before", cmdutil.ParseTimeFlag(opts.Before))
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/billing/orders", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (open/complete/cancelled)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by order type (license_purchase/license_renewal/sim_bill/service_purchase/service_renewal/service_upgrade)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start date (e.g. 2025-01-01)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End date (e.g. 2025-01-31)")
	opts.RegisterExpand(cmd, "creator", "org")

	return cmd
}
