package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupLayerfsUpdate(f *factory.Factory) *cobra.Command {
	var (
		name        string
		description string
	)

	cmd := &cobra.Command{
		Use:   "update <group-id> <layerfs-id>",
		Short: "Update a filesystem snapshot",
		Long:  "Update the name or description of a filesystem snapshot (layerfs).",
		Example: `  # Update name
  incloud device group layerfs update 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --name new-name

  # Update description
  incloud device group layerfs update 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --description "Updated desc"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{}
			if name != "" {
				reqBody["name"] = name
			}
			if description != "" {
				reqBody["description"] = description
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devicegroups/"+args[0]+"/layerfs/"+args[1], reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Layerfs %s updated.\n", args[1])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New name (1-64 chars)")
	cmd.Flags().StringVar(&description, "description", "", "New description (max 128 chars)")

	return cmd
}
