package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdLock(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "lock <id>",
		Short: "Lock a user account",
		Long:  "Lock a user account to prevent login. The user's data is preserved.",
		Example: `  # Lock a user (will prompt for confirmation)
  incloud user lock 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud user lock 507f1f77bcf86cd799439011 --yes`,
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

			_, name := resultIDName(body)
			if name == "" {
				name = id
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Lock user %q (%s)?", name, id))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Put("/api/v1/users/"+id+"/lock", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "User %q (%s) locked.\n", name, id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
