package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdGroupProject(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage device group projects",
		Long:  "Create, list, update, delete, publish, and deploy versioned project configurations for edge device groups.",
	}

	cmd.AddCommand(newCmdGroupProjectCreate(f))
	cmd.AddCommand(newCmdGroupProjectList(f))
	cmd.AddCommand(newCmdGroupProjectGet(f))
	cmd.AddCommand(newCmdGroupProjectUpdate(f))
	cmd.AddCommand(newCmdGroupProjectDelete(f))
	cmd.AddCommand(newCmdGroupProjectPublish(f))
	cmd.AddCommand(newCmdGroupProjectDeploy(f))
	cmd.AddCommand(newCmdGroupProjectDevicesSummary(f))

	return cmd
}
