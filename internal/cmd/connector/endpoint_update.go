package connector

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type endpointUpdateOptions struct {
	Name  string
	LanIP string
	VIP   string
}

func newCmdEndpointUpdate(f *factory.Factory) *cobra.Command {
	opts := &endpointUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <network-id> <endpoint-id>",
		Short: "Update an endpoint in a connector network",
		Example: `  # Update endpoint name
  incloud connector endpoint update 66827b3ccfb1842140f4222f ep123 --name new-name

  # Update LAN IP
  incloud connector endpoint update 66827b3ccfb1842140f4222f ep123 --lan-ip 192.168.2.0/24

  # Update VIP
  incloud connector endpoint update 66827b3ccfb1842140f4222f ep123 --vip 10.32.0.10`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})
			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("lan-ip") {
				body["lanIp"] = opts.LanIP
			}
			if cmd.Flags().Changed("vip") {
				body["vip"] = opts.VIP
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --lan-ip, --vip")
			}

			networkID := args[0]
			endpointID := args[1]
			respBody, err := client.Put("/api/v1/connectors/"+networkID+"/endpoints/"+endpointID, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			writeUpdated(f, "Connector endpoint", respBody)
			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Endpoint name")
	cmd.Flags().StringVar(&opts.LanIP, "lan-ip", "", "LAN IP or subnet (e.g. 192.168.1.0/24)")
	cmd.Flags().StringVar(&opts.VIP, "vip", "", "Virtual IP address")

	return cmd
}
