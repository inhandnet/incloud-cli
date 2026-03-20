package device

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdExecSpeedtest(f *factory.Factory) *cobra.Command {
	var (
		iface      string
		serverNode string
	)

	cmd := &cobra.Command{
		Use:   "speedtest <device-id>",
		Short: "Run speed test on a device",
		Long: `Run speed test from a remote device and stream results in real time.

If --interface or --server-node is not specified, the command fetches available
options from the device and prompts you to select. Press Ctrl+C to cancel.`,
		Example: `  # Run speed test (prompts for interface and server node)
  incloud device exec speedtest 507f1f77bcf86cd799439011

  # With specific interface and server node (no prompts)
  incloud device exec speedtest 507f1f77bcf86cd799439011 --interface eth0 --server-node node1

  # View available interfaces and server nodes
  incloud device exec speedtest-config 507f1f77bcf86cd799439011

  # View historical results
  incloud device exec speedtest-history 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if iface == "" || serverNode == "" {
				selected, err := promptSpeedtestConfig(f, deviceID, iface)
				if err != nil {
					return err
				}
				if iface == "" {
					iface = selected.iface
				}
				if serverNode == "" {
					serverNode = selected.serverNode
				}
			}

			return runDiagnosisStreamReplace(f, cmd, deviceID, "speedtest", map[string]any{
				"interface":  iface,
				"serverNode": serverNode,
			})
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "",
		"Network interface to use (use 'incloud device exec speedtest-config <device-id>' to list)")
	cmd.Flags().StringVar(&serverNode, "server-node", "",
		"Speed test server node ID (use 'incloud device exec speedtest-config <device-id>' to list)")

	return cmd
}

type speedtestSelection struct {
	iface      string
	serverNode string
}

// promptSpeedtestConfig fetches available interfaces and server nodes from the
// device, then prompts the user to select. If ifaceHint is non-empty, it is
// used to filter server nodes for that interface.
func promptSpeedtestConfig(f *factory.Factory, deviceID, ifaceHint string) (speedtestSelection, error) {
	client, err := f.APIClient()
	if err != nil {
		return speedtestSelection{}, err
	}

	q := url.Values{}
	if ifaceHint != "" {
		q.Set("interface", ifaceHint)
	}

	body, err := client.Get("/api/v1/devices/"+deviceID+"/diagnosis/speedtest/config", q)
	if err != nil {
		return speedtestSelection{}, err
	}

	result := gjson.GetBytes(body, "result")

	// Build interface options
	var sel speedtestSelection
	ifaces := result.Get("uplinkInterfaces").Array()
	if len(ifaces) == 0 {
		return sel, fmt.Errorf("no uplink interfaces available on this device")
	}

	if ifaceHint == "" {
		if len(ifaces) == 1 {
			sel.iface = ifaces[0].Get("name").String()
			fmt.Fprintf(f.IO.ErrOut, "Using interface: %s\n", formatIfaceLabel(&ifaces[0]))
		} else {
			opts := make([]huh.Option[string], 0, len(ifaces))
			for idx := range ifaces {
				opts = append(opts, huh.NewOption(formatIfaceLabel(&ifaces[idx]), ifaces[idx].Get("name").String()))
			}
			sel.iface, err = ui.Select(f, "Select interface", opts)
			if err != nil {
				return sel, err
			}
		}

		// Re-fetch config with selected interface to get matching server nodes
		q.Set("interface", sel.iface)
		body, err = client.Get("/api/v1/devices/"+deviceID+"/diagnosis/speedtest/config", q)
		if err != nil {
			return sel, err
		}
		result = gjson.GetBytes(body, "result")
	}

	// Build server node options
	nodes := result.Get("serverNodes").Array()
	if len(nodes) == 0 {
		return sel, fmt.Errorf("no server nodes available for interface %q", sel.iface)
	}

	if len(nodes) == 1 {
		sel.serverNode = nodes[0].Get("id").String()
		fmt.Fprintf(f.IO.ErrOut, "Using server: %s\n", formatNodeLabel(&nodes[0]))
	} else {
		opts := make([]huh.Option[string], 0, len(nodes))
		for idx := range nodes {
			opts = append(opts, huh.NewOption(formatNodeLabel(&nodes[idx]), nodes[idx].Get("id").String()))
		}
		sel.serverNode, err = ui.Select(f, "Select server node", opts)
		if err != nil {
			return sel, err
		}
	}

	return sel, nil
}

func formatIfaceLabel(i *gjson.Result) string {
	name := i.Get("name").String()
	label := i.Get("label").String()
	if label != "" && label != name {
		return fmt.Sprintf("%s (%s)", label, name)
	}
	return name
}

func formatNodeLabel(n *gjson.Result) string {
	name := n.Get("name").String()
	city := n.Get("city").String()
	country := n.Get("country").String()
	if city != "" || country != "" {
		return fmt.Sprintf("%s (%s, %s)", name, city, country)
	}
	return name
}
