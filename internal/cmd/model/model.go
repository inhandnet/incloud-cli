package model

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdModel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage AI models",
		Long:  "Deploy and manage AI models on edge devices.",
	}

	cmd.AddCommand(NewCmdDeploy(f))

	return cmd
}
