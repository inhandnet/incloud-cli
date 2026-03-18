package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdOobmDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id> [<id>...]",
		Aliases: []string{"rm"},
		Short:   "Delete OOBM resources",
		Example: `  # Delete a single resource
  incloud network oobm delete 507f1f77bcf86cd799439011

  # Delete multiple resources
  incloud network oobm delete 507f1f77bcf86cd799439011 507f1f77bcf86cd799439012

  # Skip confirmation
  incloud network oobm delete 507f1f77bcf86cd799439011 -y`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				prompt := fmt.Sprintf("Delete %d OOBM resource(s)?", len(args))
				confirmed, err := ui.Confirm(f, prompt)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			_, err = client.Do("DELETE", "/api/v1/oobm/resources/by-ids", &api.RequestOptions{
				Body: map[string]any{"ids": args},
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Deleted %d OOBM resource(s).\n", len(args))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
