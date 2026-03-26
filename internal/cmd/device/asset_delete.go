package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdAssetDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <asset-id> [asset-id...]",
		Aliases: []string{"rm"},
		Short:   "Delete network assets",
		Long:    "Delete one or more network assets. Supports both single and batch deletion.",
		Args:    cobra.MinimumNArgs(1),
		Example: `  # Delete a single asset
  incloud device asset delete 507f1f77bcf86cd799439011

  # Delete multiple assets
  incloud device asset delete id1 id2 id3

  # Skip confirmation
  incloud device asset delete 507f1f77bcf86cd799439011 -y`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if !yes {
				prompt := fmt.Sprintf("Delete %d asset(s) (%s)?", len(args), summarizeIDs(args))
				confirmed, err := ui.Confirm(f, prompt)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			if len(args) == 1 {
				_, err = client.Delete("/api/v1/network/assets/" + args[0])
				if err != nil {
					return err
				}
				fmt.Fprintf(f.IO.ErrOut, "Asset %s deleted.\n", args[0])
			} else {
				body := map[string]any{
					"ids": args,
				}
				_, err = client.Post("/api/v1/network/assets/remove", body)
				if err != nil {
					return err
				}
				fmt.Fprintf(f.IO.ErrOut, "%d assets deleted.\n", len(args))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// summarizeIDs returns a shortened display string for a list of IDs.
func summarizeIDs(ids []string) string {
	if len(ids) <= 3 {
		return strings.Join(ids, ", ")
	}
	return strings.Join(ids[:3], ", ") + fmt.Sprintf(" and %d more", len(ids)-3)
}
