package product

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
		Short:   "Delete a product",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete product %s?", id))
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

			_, err = client.Delete("/api/v1/products/" + id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Product %s deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
