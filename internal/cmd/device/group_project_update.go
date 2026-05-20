package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupProjectUpdate(f *factory.Factory) *cobra.Command {
	var (
		dockerCompose string
		systemdJSON   string
		artifactKey   string
		layerfsID     string
		description   string
	)

	cmd := &cobra.Command{
		Use:   "update <group-id> <project-id>",
		Short: "Update a project version",
		Long:  "Update a project version's configuration. Published projects can only have their description updated.",
		Example: `  # Update docker-compose
  incloud device group project update 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --docker-compose "version: '3'\nservices: ..."

  # Update description only (works for published projects)
  incloud device group project update 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695 --description "Updated release"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{}
			if cmd.Flags().Changed("docker-compose") {
				reqBody["dockerCompose"] = dockerCompose
			}
			if cmd.Flags().Changed("systemd") {
				var systemd []interface{}
				if err := json.Unmarshal([]byte(systemdJSON), &systemd); err != nil {
					return fmt.Errorf("invalid --systemd JSON: %w", err)
				}
				reqBody["systemd"] = systemd
			}
			if cmd.Flags().Changed("artifact-key") {
				if artifactKey == "" {
					reqBody["artifactKey"] = nil
				} else {
					reqBody["artifactKey"] = artifactKey
				}
			}
			if cmd.Flags().Changed("layerfs-id") {
				if layerfsID == "" {
					reqBody["layerfsId"] = nil
				} else {
					reqBody["layerfsId"] = layerfsID
				}
			}
			if description != "" {
				reqBody["description"] = description
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devicegroups/"+args[0]+"/projects/"+args[1], reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Project %s updated.\n", args[1])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&dockerCompose, "docker-compose", "", "Docker Compose file content")
	cmd.Flags().StringVar(&systemdJSON, "systemd", "", `Systemd unit definitions as JSON array`)
	cmd.Flags().StringVar(&artifactKey, "artifact-key", "", "S3 artifact key (empty to clear)")
	cmd.Flags().StringVar(&layerfsID, "layerfs-id", "", "Layerfs snapshot ID (empty to clear)")
	cmd.Flags().StringVar(&description, "description", "", "Project description (1-256 chars)")

	return cmd
}
