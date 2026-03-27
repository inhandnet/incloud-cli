package oobm

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmUpdateOptions struct {
	DeviceID string
	Name     string
	ClientIP string
	Services []string
	IdleTime int
	ConnTime int
}

func NewCmdOobmUpdate(f *factory.Factory) *cobra.Command {
	opts := &OobmUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an OOBM resource",
		Long: `Update an Out-of-Band Management resource.

All fields are required because the API performs a full replacement.
Service format: protocol:port[:usage]`,
		Example: `  # Update a resource (all fields required — full replacement)
  incloud oobm update 507f1f77bcf86cd799439011 \
    --device-id 607f1f77bcf86cd799439022 \
    --name "Router SSH" \
    --client-ip 192.168.1.1 \
    --service ssh:22:cli

  # Update with multiple services
  incloud oobm update 507f1f77bcf86cd799439011 \
    --device-id 607f1f77bcf86cd799439022 \
    --name "Router Web+SSH" \
    --client-ip 192.168.1.1 \
    --service https:443 \
    --service ssh:22:web \
    --idle-time 600 --conn-time 7200`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

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

			respBody, err := client.Put("/api/v1/oobm/resources/"+id, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			var resp struct {
				ID   string `json:"_id"`
				Name string `json:"name"`
			}
			_ = json.Unmarshal(respBody, &resp)
			fmt.Fprintf(f.IO.ErrOut, "OOBM resource %q (%s) updated.\n", resp.Name, resp.ID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
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
