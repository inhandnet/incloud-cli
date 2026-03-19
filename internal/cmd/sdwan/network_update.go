package sdwan

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type networkUpdateOptions struct {
	Name            string
	TunnelMode      string
	ForceAllTraffic bool
	Hubs            []string
	Spokes          []string
}

func newCmdNetworkUpdate(f *factory.Factory) *cobra.Command {
	opts := &networkUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an SD-WAN network",
		Long: `Update an SD-WAN network. Fetches the current state first and merges your changes,
so you only need to specify the fields you want to change.

When --hub or --spoke is specified, the device list is replaced entirely.
Existing device configurations (subnets, tunnel ports, etc.) are preserved
for devices that remain in the network.`,
		Example: `  # Update name
  incloud sdwan network update <id> --name new-name

  # Change tunnel mode
  incloud sdwan network update <id> --tunnel-mode symmetric

  # Replace hub devices
  incloud sdwan network update <id> --hub new-hub-id

  # Add spokes (replaces existing spoke list)
  incloud sdwan network update <id> --spoke spoke1 --spoke spoke2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hasChange := cmd.Flags().Changed("name") || cmd.Flags().Changed("tunnel-mode") ||
				cmd.Flags().Changed("force-all-traffic") || cmd.Flags().Changed("hub") || cmd.Flags().Changed("spoke")
			if !hasChange {
				return fmt.Errorf("no fields to update; specify at least one of --name, --tunnel-mode, --force-all-traffic, --hub, --spoke")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			id := args[0]

			// Fetch current network to build baseline
			currentBody, err := client.Get(apiBase+"/networks/"+id, nil)
			if err != nil {
				return err
			}
			var current struct {
				Result struct {
					Name            string `json:"name"`
					TunnelMode      string `json:"tunnelCreationMode"`
					ForceAllTraffic bool   `json:"forceSendAllTraffic"`
				} `json:"result"`
			}
			if err := json.Unmarshal(currentBody, &current); err != nil {
				return fmt.Errorf("parsing current network: %w", err)
			}

			// Build update body with current values as baseline
			body := map[string]interface{}{
				"name":                current.Result.Name,
				"forceSendAllTraffic": current.Result.ForceAllTraffic,
			}
			if current.Result.TunnelMode != "" {
				body["tunnelCreationMode"] = current.Result.TunnelMode
			}

			// Override with user-specified values
			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("tunnel-mode") {
				body["tunnelCreationMode"] = opts.TunnelMode
			}
			if cmd.Flags().Changed("force-all-traffic") {
				body["forceSendAllTraffic"] = opts.ForceAllTraffic
			}

			// Handle hubs: use user-provided or preserve current devices
			if cmd.Flags().Changed("hub") {
				body["hubs"] = toMembers(opts.Hubs)
			} else {
				hubs, err := fetchCurrentMembers(client, id, "hub")
				if err != nil {
					return fmt.Errorf("fetching current hubs: %w", err)
				}
				body["hubs"] = hubs
			}

			// Handle spokes
			if cmd.Flags().Changed("spoke") {
				body["spokes"] = toMembers(opts.Spokes)
			} else {
				spokes, err := fetchCurrentMembers(client, id, "spoke")
				if err != nil {
					return fmt.Errorf("fetching current spokes: %w", err)
				}
				body["spokes"] = spokes
			}

			respBody, err := client.Put(apiBase+"/networks/"+id, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			writeUpdated(f, "SD-WAN network", respBody)
			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Network name (max 64 chars)")
	cmd.Flags().StringVar(&opts.TunnelMode, "tunnel-mode", "", "Tunnel creation mode: mesh or symmetric")
	cmd.Flags().BoolVar(&opts.ForceAllTraffic, "force-all-traffic", false, "Force send all traffic through tunnels")
	cmd.Flags().StringArrayVar(&opts.Hubs, "hub", nil, "Hub device ID (repeatable, replaces existing hubs)")
	cmd.Flags().StringArrayVar(&opts.Spokes, "spoke", nil, "Spoke device ID (repeatable, replaces existing spokes)")

	return cmd
}

// fetchCurrentMembers fetches devices of a given role and returns them as
// NetworkMember payloads preserving subnets, tunnelPorts, ifacePublicIpMappings,
// and preferredHub so that an update doesn't lose existing device configuration.
func fetchCurrentMembers(client *api.APIClient, networkID, role string) ([]map[string]interface{}, error) {
	q := url.Values{}
	q.Set("role", role)
	q.Set("page", "0")
	q.Set("limit", "500")

	body, err := client.Get(apiBase+"/networks/"+networkID+"/devices", q)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	members := make([]map[string]interface{}, 0, len(resp.Result))
	for _, raw := range resp.Result {
		var device map[string]interface{}
		if err := json.Unmarshal(raw, &device); err != nil {
			return nil, err
		}

		member := map[string]interface{}{
			"deviceId": device["_id"],
		}

		// Preserve enabled subnets
		if subnets, ok := device["subnets"].([]interface{}); ok {
			var selected []map[string]interface{}
			for _, s := range subnets {
				sub, ok := s.(map[string]interface{})
				if !ok {
					continue
				}
				selected = append(selected, map[string]interface{}{
					"id":   sub["id"],
					"cidr": sub["subnet"],
				})
			}
			if len(selected) > 0 {
				member["subnets"] = selected
			}
		}

		// Preserve tunnel ports (hubs)
		if tp, ok := device["tunnelPorts"].(map[string]interface{}); ok {
			member["tunnelPorts"] = tp
		}

		// Preserve public IP mappings (hubs)
		if mappings, ok := device["ifacePublicIpMappings"].([]interface{}); ok && len(mappings) > 0 {
			member["ifacePublicIpMappings"] = mappings
		}

		// Preserve preferred hub (spokes)
		if ph, ok := device["preferredHub"].(string); ok && ph != "" {
			member["preferredHub"] = ph
		}

		members = append(members, member)
	}
	return members, nil
}
