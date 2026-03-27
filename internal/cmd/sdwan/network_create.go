package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type networkCreateOptions struct {
	Name            string
	Type            string
	TunnelMode      string
	ForceAllTraffic bool
	Hubs            []string
	Spokes          []string
	LoopbackCidr    string
	TunnelCidr      string
}

func newCmdNetworkCreate(f *factory.Factory) *cobra.Command {
	opts := &networkCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an SD-WAN network",
		Example: `  # Create a hub-spoke network
  incloud sdwan network create --name my-sdwan --hub <deviceId>

  # Multiple hubs and spokes
  incloud sdwan network create --name office-vpn \
    --hub hub1-id --hub hub2-id --spoke spoke1-id --spoke spoke2-id

  # Symmetric tunnel mode with custom CIDR
  incloud sdwan network create --name site-vpn \
    --tunnel-mode symmetric --hub hub-id \
    --loopback-cidr 10.0.0.0/17 --tunnel-cidr 10.0.128.0/17`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"name":   opts.Name,
				"type":   opts.Type,
				"hubs":   toMembers(opts.Hubs),
				"spokes": toMembers(opts.Spokes),
			}

			if opts.TunnelMode != "" {
				body["tunnelCreationMode"] = opts.TunnelMode
			}
			if cmd.Flags().Changed("force-all-traffic") {
				body["forceSendAllTraffic"] = opts.ForceAllTraffic
			}
			if cmd.Flags().Changed("loopback-cidr") {
				body["loopbackCidr"] = opts.LoopbackCidr
			}
			if cmd.Flags().Changed("tunnel-cidr") {
				body["tunnelCidr"] = opts.TunnelCidr
			}

			respBody, err := client.Post(apiBase+"/networks", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			writeCreated(f, "SD-WAN network", respBody)
			return formatOutput(cmd, f.IO, respBody)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Network name (required, max 64 chars)")
	cmd.Flags().StringVar(&opts.Type, "type", "hub_spoke", "Network type")
	cmd.Flags().StringVar(&opts.TunnelMode, "tunnel-mode", "", "Tunnel creation mode: mesh (default) or symmetric")
	cmd.Flags().BoolVar(&opts.ForceAllTraffic, "force-all-traffic", false, "Force send all traffic through tunnels")
	cmd.Flags().StringArrayVar(&opts.Hubs, "hub", nil,
		"Hub device ID (required, repeatable, max 5; use 'incloud sdwan candidates --role hub' to find IDs)")
	cmd.Flags().StringArrayVar(&opts.Spokes, "spoke", nil,
		"Spoke device ID (repeatable, max 500; use 'incloud sdwan candidates --role spoke' to find IDs)")
	cmd.Flags().StringVar(&opts.LoopbackCidr, "loopback-cidr", "10.113.0.0/17", "Loopback address CIDR pool")
	cmd.Flags().StringVar(&opts.TunnelCidr, "tunnel-cidr", "10.113.128.0/17", "Tunnel address CIDR pool")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("hub")

	return cmd
}
