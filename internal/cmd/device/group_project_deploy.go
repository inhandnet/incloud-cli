package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdGroupProjectDeploy(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "deploy <group-id> <project-id>",
		Short: "Deploy a project version to the device group",
		Long:  "Deploy a published project version to all devices in the group. Pinned devices will not be affected.",
		Example: `  # Deploy a project
  incloud device group project deploy 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695

  # Skip confirmation
  incloud device group project deploy 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := args[0]
			projectID := args[1]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Deploy project %s to group %s?", projectID, groupID))
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

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devicegroups/"+groupID+"/projects/"+projectID+"/deploy", nil)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Project %s deployed to group %s.\n", projectID, groupID)
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
