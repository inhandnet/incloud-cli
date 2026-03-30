package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type endpointCreateOptions struct {
	DeviceID string
	Name     string
	LanIP    string
	VIP      string
}

func newCmdEndpointCreate(f *factory.Factory) *cobra.Command {
	opts := &endpointCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <network-id>",
		Short: "Create an endpoint in a connector network",
		Example: `  # Create an endpoint
  incloud connector endpoint create 66827b3ccfb1842140f4222f --device-id abc123 --name my-ep --lan-ip 192.168.1.0/24

  # With optional VIP
  incloud connector endpoint create 66827b3ccfb1842140f4222f --device-id abc123 --name my-ep --lan-ip 192.168.1.0/24 --vip 10.32.0.5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"deviceId": opts.DeviceID,
				"name":     opts.Name,
				"lanIp":    opts.LanIP,
			}
			if opts.VIP != "" {
				body["vip"] = opts.VIP
			}

			respBody, err := client.Post("/api/v1/connectors/"+networkID+"/endpoints", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			cmdutil.WriteCreated(f, "Connector endpoint", respBody)
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Device ID (required; use 'incloud connector device list <network-id>' to find IDs)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Endpoint name (required)")
	cmd.Flags().StringVar(&opts.LanIP, "lan-ip", "", "LAN IP or subnet (required, e.g. 192.168.1.0/24)")
	cmd.Flags().StringVar(&opts.VIP, "vip", "", "Virtual IP address")

	_ = cmd.MarkFlagRequired("device-id")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("lan-ip")

	return cmd
}
