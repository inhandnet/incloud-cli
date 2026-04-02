package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type assetListOptions struct {
	cmdutil.ListFlags
	Name     string
	MAC      string
	Number   string
	Category []string
	Status   []string
}

var defaultAssetListFields = []string{"_id", "name", "mac", "number", "category", "status", "warrantyExpiration", "createdAt"}

func newCmdAssetList(f *factory.Factory) *cobra.Command {
	opts := &assetListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List network assets",
		Long:    "List all network assets in the current organization.",
		Aliases: []string{"ls"},
		Example: `  # List all assets
  incloud device asset list

  # Filter by category
  incloud device asset list --category router,ap

  # Filter by status
  incloud device asset list --status in_use,in_stock

  # Search by name
  incloud device asset list --name "office"

  # Search by MAC address
  incloud device asset list --mac "00:18:05"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultAssetListFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.MAC != "" {
				q.Set("mac", opts.MAC)
			}
			if opts.Number != "" {
				q.Set("number", opts.Number)
			}
			for _, c := range opts.Category {
				q.Add("category", c)
			}
			for _, s := range opts.Status {
				q.Add("status", s)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/network/assets", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by asset name (partial match)")
	cmd.Flags().StringVar(&opts.MAC, "mac", "", "Filter by MAC address (partial match)")
	cmd.Flags().StringVar(&opts.Number, "number", "", "Filter by asset number (partial match)")
	cmd.Flags().StringSliceVar(&opts.Category, "category", nil, "Filter by category ("+assetCategories+")")
	cmd.Flags().StringSliceVar(&opts.Status, "status", nil, "Filter by status ("+assetStatuses+")")

	return cmd
}
