package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupRegistryUpdate(f *factory.Factory) *cobra.Command {
	var registriesJSON string

	cmd := &cobra.Command{
		Use:   "update <group-id>",
		Short: "Update registry configuration for a device group",
		Long:  "Update the container registry configurations for an edge device group. Provide registries as a JSON array.",
		Example: `  # Set a single registry
  incloud device group registry update 507f1f77bcf86cd799439011 --registries '[{"url":"registry.example.com","username":"user","password":"pass"}]'

  # Clear all registries
  incloud device group registry update 507f1f77bcf86cd799439011 --registries '[]'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var registries []interface{}
			if err := json.Unmarshal([]byte(registriesJSON), &registries); err != nil {
				return fmt.Errorf("invalid --registries JSON: %w", err)
			}

			reqBody := map[string]interface{}{
				"registries": registries,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devicegroups/"+args[0], reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Registry configuration updated for group %s.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&registriesJSON, "registries", "", `Registry configurations as JSON array (e.g. '[{"url":"...","username":"...","password":"..."}]')`)
	_ = cmd.MarkFlagRequired("registries")

	return cmd
}
