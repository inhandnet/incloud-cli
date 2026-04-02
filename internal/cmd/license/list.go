package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Status   string
	Type     string
	Attached string
	OrderID  string
	OrgID    string
}

var defaultListFields = []string{"_id", "type", "status", "deviceId", "activatedDate", "expiresAt"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List licenses",
		Long:    "List licenses with optional filtering by status, type, attachment, and organization.",
		Aliases: []string{"ls"},
		Example: `  # List all activated licenses
  incloud license list --status activated

  # List unattached licenses of a specific type
  incloud license list --status inactivated --attached false --type basic

  # List licenses for a specific organization (admin)
  incloud license list --org-id 64a1b2c3d4e5f6a7b8c9d0e1

  # List expiring licenses
  incloud license list --status to_be_expired -o yaml

  # Filter by order
  incloud license list --order-id 64a1b2c3d4e5f6a7b8c9d0e1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if opts.Attached != "" {
				q.Set("attached", opts.Attached)
			}
			if opts.OrderID != "" {
				q.Set("orderId", opts.OrderID)
			}
			if opts.OrgID != "" {
				q.Set("oid", opts.OrgID)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/billing/licenses", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (activated/inactivated/to_be_expired/expired)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by license type slug")
	cmd.Flags().StringVar(&opts.Attached, "attached", "", "Filter by attachment status (true/false)")
	cmd.Flags().StringVar(&opts.OrderID, "order-id", "", "Filter by order ID")
	cmd.Flags().StringVar(&opts.OrgID, "org-id", "", "Filter by organization ID")
	opts.RegisterExpand(cmd, "type", "device", "org")

	return cmd
}
