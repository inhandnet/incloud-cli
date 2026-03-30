package connector

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type networkUpdateOptions struct {
	Name        string
	Description string
}

func newCmdNetworkUpdate(f *factory.Factory) *cobra.Command {
	opts := &networkUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a connector network",
		Example: `  # Update name
  incloud connector network update 66827b3ccfb1842140f4222f --name new-name

  # Update description
  incloud connector network update 66827b3ccfb1842140f4222f --description "Updated desc"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})
			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("description") {
				body["description"] = opts.Description
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --description")
			}

			respBody, err := client.Put("/api/v1/connectors/"+args[0], body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			cmdutil.WriteUpdated(f, "Connector network", respBody)
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Network name (max 128 chars)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Network description (max 256 chars)")

	return cmd
}
