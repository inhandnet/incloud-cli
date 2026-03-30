package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete a user",
		Example: `  # Delete a user (will prompt for confirmation)
  incloud user delete 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud user delete 507f1f77bcf86cd799439011 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/users/"+id, nil)
			if err != nil {
				return err
			}

			_, name := api.ResultIDName(body)
			if name == "" {
				name = id
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete user %q (%s)?", name, id))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Delete("/api/v1/users/" + id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "User %q (%s) deleted.\n", name, id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
