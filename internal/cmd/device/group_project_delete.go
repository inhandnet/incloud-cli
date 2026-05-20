package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdGroupProjectDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <group-id> <project-id>[,<id2>,...]",
		Short:   "Delete project versions",
		Long:    "Delete one or more project versions from a device group.",
		Aliases: []string{"rm"},
		Example: `  # Delete a project
  incloud device group project delete 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695

  # Delete multiple
  incloud device group project delete 507f1f77bcf86cd799439011 id1,id2

  # Skip confirmation
  incloud device group project delete 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID := args[0]
			idsArg := args[1]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Delete project(s) %s from group %s?", idsArg, groupID))
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

			ids := strings.Split(idsArg, ",")
			reqBody := map[string]interface{}{
				"ids": ids,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devicegroups/"+groupID+"/projects/bulk-remove", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Project(s) deleted from group %s.\n", groupID)
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
