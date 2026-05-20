package touch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdClientDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <client-id>",
		Short:   "Delete a touch client",
		Long:    "Delete a remote access client by its ID.",
		Aliases: []string{"rm"},
		Example: `  # Delete a client
  incloud touch client delete 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud touch client delete 507f1f77bcf86cd799439011 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete touch client %s?", args[0]))
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

			_, err = client.Delete("/api/v1/touch/clients/" + args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Touch client %s deleted.\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
