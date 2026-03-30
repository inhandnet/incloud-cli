package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type networkCreateOptions struct {
	Name        string
	Description string
	Subnet      string
}

func newCmdNetworkCreate(f *factory.Factory) *cobra.Command {
	opts := &networkCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a connector network",
		Example: `  # Create with name only
  incloud connector network create --name my-vpn

  # With subnet and description
  incloud connector network create --name my-vpn --subnet 10.32.0.0/12 --description "Office VPN"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"name": opts.Name,
			}
			if opts.Description != "" {
				body["description"] = opts.Description
			}
			if opts.Subnet != "" {
				body["subnet"] = opts.Subnet
			}

			respBody, err := client.Post("/api/v1/connectors", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			cmdutil.WriteCreated(f, "Connector network", respBody)
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Network name (required, max 128 chars)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Network description (max 256 chars)")
	cmd.Flags().StringVar(&opts.Subnet, "subnet", "", "Network subnet (e.g. 10.32.0.0/12)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
