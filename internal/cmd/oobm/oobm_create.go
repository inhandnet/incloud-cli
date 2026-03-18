package oobm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmCreateOptions struct {
	DeviceID string
	Name     string
	ClientIP string
	Services []string
	IdleTime int
	ConnTime int
}

func parseService(s string) (map[string]any, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid service format %q: expected protocol:port[:usage]", s)
	}

	protocol := strings.ToLower(parts[0])
	validProtocols := map[string]bool{"http": true, "https": true, "tcp": true, "telnet": true, "ssh": true}
	if !validProtocols[protocol] {
		return nil, fmt.Errorf("unsupported protocol %q (supported: http, https, tcp, telnet, ssh)", protocol)
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port %q in service %q", parts[1], s)
	}

	svc := map[string]any{
		"protocol": protocol,
		"port":     port,
	}

	if len(parts) == 3 {
		usage := strings.ToLower(parts[2])
		if usage != "web" && usage != "cli" {
			return nil, fmt.Errorf("invalid usage %q in service %q (supported: web, cli)", usage, s)
		}
		svc["usage"] = usage
	}

	return svc, nil
}

func NewCmdOobmCreate(f *factory.Factory) *cobra.Command {
	opts := &OobmCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an OOBM resource",
		Long: `Create a new Out-of-Band Management resource to establish remote access tunnels.

Supported protocols: http, https, tcp, telnet, ssh
Service usage types: web (browser-based), cli (command-line)

Service format: protocol:port[:usage]
  When usage is omitted, the server determines the default based on protocol.`,
		Example: `  # Create a resource with SSH access
  incloud oobm create \
    --device-id 507f1f77bcf86cd799439011 \
    --name "Router SSH" \
    --client-ip 192.168.1.1 \
    --service ssh:22:cli

  # Create with multiple services
  incloud oobm create \
    --device-id 507f1f77bcf86cd799439011 \
    --name "Router Web+SSH" \
    --client-ip 192.168.1.1 \
    --service https:443 \
    --service ssh:22:web \
    --idle-time 600 --conn-time 7200`,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := parseServices(opts.Services)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"deviceId": opts.DeviceID,
				"name":     opts.Name,
				"clientIp": opts.ClientIP,
				"services": services,
				"idleTime": opts.IdleTime,
				"connTime": opts.ConnTime,
			}

			respBody, err := client.Post("/api/v1/oobm/resources", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			var resp struct {
				ID   string `json:"_id"`
				Name string `json:"name"`
			}
			_ = json.Unmarshal(respBody, &resp)
			fmt.Fprintf(f.IO.ErrOut, "OOBM resource %q created. (id: %s)\n", resp.Name, resp.ID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Device ID (required; use 'incloud device list' to find IDs)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Resource name (required, 1-32 chars)")
	cmd.Flags().StringVar(&opts.ClientIP, "client-ip", "", "Client IP address (required)")
	cmd.Flags().StringArrayVar(&opts.Services, "service", nil, "Service in protocol:port[:usage] format (required, can be repeated)")
	cmd.Flags().IntVar(&opts.IdleTime, "idle-time", 300, "Idle timeout in seconds (60-3600)")
	cmd.Flags().IntVar(&opts.ConnTime, "conn-time", 3600, "Connection timeout in seconds (3600-604800)")

	_ = cmd.MarkFlagRequired("device-id")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("client-ip")
	_ = cmd.MarkFlagRequired("service")

	return cmd
}
