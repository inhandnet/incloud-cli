package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdUnlock(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unlock <id>",
		Short:   "Unlock a user account",
		Long:    "Unlock a previously locked user account to restore login access.",
		Example: `  incloud user unlock 507f1f77bcf86cd799439011`,
		Args:    cobra.ExactArgs(1),
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

			_, err = client.Put("/api/v1/users/"+id+"/unlock", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "User %q (%s) unlocked.\n", name, id)
			return nil
		},
	}

	return cmd
}
