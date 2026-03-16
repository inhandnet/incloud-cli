package product

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdProduct(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product",
		Short: "Manage products",
		Long:  "List, create, update, delete, and manage products on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdUpdate(f))
	cmd.AddCommand(NewCmdDelete(f))

	return cmd
}
