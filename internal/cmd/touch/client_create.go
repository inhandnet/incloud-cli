package touch

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdClientCreate(f *factory.Factory) *cobra.Command {
	var (
		name       string
		clientType string
		deviceID   string
		ip         string
		serialJSON string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a touch client",
		Long:  "Create a new remote access client (downstream endpoint) on an edge device. Supports ETHERNET and SERIAL types.",
		Example: `  # Create an Ethernet client
  incloud touch client create --name my-plc --type ETHERNET --device-id 507f1f77bcf86cd799439011 --ip 192.168.1.100

  # Create a Serial client
  incloud touch client create --name my-serial --type SERIAL --device-id 507f1f77bcf86cd799439011 --serial '{"name":"ttyS0","type":"rs232","baudRate":"9600","dataBit":8,"stopBit":1,"parityBit":"NONE"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"name":     name,
				"type":     clientType,
				"deviceId": deviceID,
			}

			if ip != "" {
				reqBody["ethernet"] = map[string]interface{}{
					"ip": ip,
				}
			}
			if serialJSON != "" {
				var serial interface{}
				if err := json.Unmarshal([]byte(serialJSON), &serial); err != nil {
					return fmt.Errorf("invalid --serial JSON: %w", err)
				}
				reqBody["serial"] = serial
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/touch/clients", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Touch client created.\n")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Client name (required, 1-128 chars)")
	cmd.Flags().StringVar(&clientType, "type", "", "Client type: ETHERNET or SERIAL (required)")
	cmd.Flags().StringVar(&deviceID, "device-id", "", "Edge device ID (required)")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address for ETHERNET type")
	cmd.Flags().StringVar(&serialJSON, "serial", "", "Serial configuration as JSON for SERIAL type")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("device-id")

	return cmd
}
