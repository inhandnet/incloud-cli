package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupProjectGet(f *factory.Factory) *cobra.Command {
	var expand string

	cmd := &cobra.Command{
		Use:   "get <group-id> <project-id>",
		Short: "Get a project version",
		Long:  "Get details of a project version in a device group.",
		Example: `  # Get project details
  incloud device group project get 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695

  # With expanded resources
  incloud device group project get 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --expand creator,edge-layerfs`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if expand != "" {
				q.Set("expand", expand)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devicegroups/"+args[0]+"/projects/"+args[1], q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&expand, "expand", "", "Expand related resources (e.g. creator,edge-layerfs)")

	return cmd
}
