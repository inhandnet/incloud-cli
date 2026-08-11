package org

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ReparentOptions struct {
	MovingOrgID       string
	NewParentID       string
	MoveBillingAssets bool
}

func NewCmdReparent(f *factory.Factory) *cobra.Command {
	opts := &ReparentOptions{}

	cmd := &cobra.Command{
		Use:   "reparent",
		Short: "Move an organization under a new parent",
		Long: `Move an organization (and its whole subtree) under a new parent organization.

This is a superadmin operation that re-parents the moving org while keeping its
oid stable. Billing assets (subscriptions/orders) are migrated only when
--move-billing-assets is set. The server runs precheck validations and rejects
the request on cycles, exceeding max depth, or exceeding the org count limit.`,
		Example: `  # Move an org under a new parent
  incloud org reparent --moving 61259f8f4be3e571fcfa4d75 --new-parent 6125a0114be3e571fcfa4d80

  # Also migrate billing assets to the new root tenant
  incloud org reparent --moving 61259f8f4be3e571fcfa4d75 --new-parent 6125a0114be3e571fcfa4d80 --move-billing-assets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"movingOrgId":       opts.MovingOrgID,
				"newParentId":       opts.NewParentID,
				"moveBillingAssets": opts.MoveBillingAssets,
			}

			respBody, err := client.Post("/api/v1/orgs/reparent", body)
			if err != nil {
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			id, name := api.ResultIDName(respBody)
			fmt.Fprintf(f.IO.ErrOut, "Organization %q reparented. (id: %s)\n", name, id)

			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.MovingOrgID, "moving", "", "ID of the organization to move (required)")
	cmd.Flags().StringVar(&opts.NewParentID, "new-parent", "", "ID of the new parent organization (required)")
	cmd.Flags().BoolVar(&opts.MoveBillingAssets, "move-billing-assets", false, "Migrate billing assets (subscriptions/orders) to the new root tenant")

	_ = cmd.MarkFlagRequired("moving")
	_ = cmd.MarkFlagRequired("new-parent")

	return cmd
}
