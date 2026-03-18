package org

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete an organization",
		Long:    "Delete an organization and cascade-delete its roles, invitations, and customers.",
		Example: `  # Delete an organization (will prompt for confirmation)
  incloud org delete 61259f8f4be3e571fcfa4d75

  # Skip confirmation
  incloud org delete 61259f8f4be3e571fcfa4d75 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/orgs/"+id, nil)
			if err != nil {
				return err
			}

			_, name := resultIDName(body)
			if name == "" {
				name = id
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete organization %q (%s)?", name, id))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Delete("/api/v1/orgs/" + id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Organization %q (%s) deleted.\n", name, id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
