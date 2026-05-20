package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupProjectCreate(f *factory.Factory) *cobra.Command {
	var (
		dockerCompose string
		systemdJSON   string
		artifactKey   string
		layerfsID     string
	)

	cmd := &cobra.Command{
		Use:   "create <group-id>",
		Short: "Create a project version",
		Long:  "Create a new project version for a device group with docker-compose and/or systemd configurations.",
		Example: `  # Create with docker-compose
  incloud device group project create 507f1f77bcf86cd799439011 --docker-compose "version: '3'\nservices:\n  app:\n    image: myapp:latest"

  # Create with systemd units
  incloud device group project create 507f1f77bcf86cd799439011 --systemd '[{"name":"myservice","content":"[Unit]\nDescription=My Service\n[Service]\nExecStart=/usr/bin/myapp"}]'

  # Create with artifact and layerfs
  incloud device group project create 507f1f77bcf86cd799439011 --docker-compose "..." --artifact-key "path/to/artifact" --layerfs-id 653b1ff2a84e171614d88695`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{}
			if dockerCompose != "" {
				reqBody["dockerCompose"] = dockerCompose
			}
			if systemdJSON != "" {
				var systemd []interface{}
				if err := json.Unmarshal([]byte(systemdJSON), &systemd); err != nil {
					return fmt.Errorf("invalid --systemd JSON: %w", err)
				}
				reqBody["systemd"] = systemd
			}
			if artifactKey != "" {
				reqBody["artifactKey"] = artifactKey
			}
			if layerfsID != "" {
				reqBody["layerfsId"] = layerfsID
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devicegroups/"+args[0]+"/projects", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Project version created in group %s.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&dockerCompose, "docker-compose", "", "Docker Compose file content")
	cmd.Flags().StringVar(&systemdJSON, "systemd", "", `Systemd unit definitions as JSON array (e.g. '[{"name":"svc","content":"..."}]')`)
	cmd.Flags().StringVar(&artifactKey, "artifact-key", "", "S3 artifact key")
	cmd.Flags().StringVar(&layerfsID, "layerfs-id", "", "Layerfs snapshot ID to use as base")

	return cmd
}
