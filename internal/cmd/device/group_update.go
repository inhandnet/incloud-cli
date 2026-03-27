package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GroupUpdateOptions struct {
	Name     string
	Firmware string
	Tags     []string
}

func newCmdGroupUpdate(f *factory.Factory) *cobra.Command {
	opts := &GroupUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a device group",
		Example: `  # Rename a device group
  incloud device group update 507f1f77bcf86cd799439011 --name "New Name"

  # Update firmware
  incloud device group update 507f1f77bcf86cd799439011 --firmware V2.0.30

  # Set tags
  incloud device group update 507f1f77bcf86cd799439011 --tag env=prod --tag region=us`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}

			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("firmware") {
				body["firmware"] = opts.Firmware
			}
			if cmd.Flags().Changed("tag") {
				tags, err := parseKeyValues(opts.Tags)
				if err != nil {
					return err
				}
				body["tags"] = tags
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one field must be specified (--name, --firmware, or --tag)")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			respBody, err := client.Put("/api/v1/devicegroups/"+args[0], body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			var resp struct {
				Result struct {
					ID   string `json:"_id"`
					Name string `json:"name"`
				} `json:"result"`
			}
			_ = json.Unmarshal(respBody, &resp)
			fmt.Fprintf(f.IO.ErrOut, "Device group %q (%s) updated.\n", resp.Result.Name, resp.Result.ID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Group name (1-128 chars)")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Firmware version")
	cmd.Flags().StringArrayVar(&opts.Tags, "tag", nil, "Tag in key=value format (can be repeated, replaces all tags)")

	return cmd
}
