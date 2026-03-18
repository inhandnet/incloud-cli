package oobm

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmSerialCreateOptions struct {
	DeviceID string
	Name     string
	Speed    int
	DataBits int
	StopBits int
	Parity   int
	Xonxoff  bool
	IdleTime int
	ConnTime int
	Usage    string
}

func NewCmdOobmSerialCreate(f *factory.Factory) *cobra.Command {
	opts := &OobmSerialCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an OOBM serial port configuration",
		Long: `Create a new OOBM serial port configuration for remote console access.

Parity values: 0=None, 1=Odd, 2=Even
Usage types: web (browser-based terminal), cli (SSH tunnel)`,
		Example: `  # Create a serial port with defaults
  incloud oobm serial create \
    --device-id 507f1f77bcf86cd799439011 \
    --name "Console Port"

  # Create with custom serial settings
  incloud oobm serial create \
    --device-id 507f1f77bcf86cd799439011 \
    --name "RS232" \
    --speed 115200 --data-bits 8 --stop-bits 1 --parity 0 \
    --usage cli --idle-time 600`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"deviceId": opts.DeviceID,
				"name":     opts.Name,
				"speed":    opts.Speed,
				"dataBits": opts.DataBits,
				"stopBits": opts.StopBits,
				"parity":   opts.Parity,
				"xonxoff":  opts.Xonxoff,
				"idleTime": opts.IdleTime,
				"connTime": opts.ConnTime,
				"usage":    opts.Usage,
			}

			respBody, err := client.Post("/api/v1/oobm/serials", body)
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
			fmt.Fprintf(f.IO.ErrOut, "OOBM serial %q created. (id: %s)\n", resp.Name, resp.ID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Device ID (required; use 'incloud device list' to find IDs)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name (required, 1-32 chars)")
	cmd.Flags().IntVar(&opts.Speed, "speed", 9600, "Baud rate (300-230400)")
	cmd.Flags().IntVar(&opts.DataBits, "data-bits", 8, "Data bits (5-8)")
	cmd.Flags().IntVar(&opts.StopBits, "stop-bits", 1, "Stop bits (1-2)")
	cmd.Flags().IntVar(&opts.Parity, "parity", 0, "Parity: 0=None, 1=Odd, 2=Even")
	cmd.Flags().BoolVar(&opts.Xonxoff, "xonxoff", false, "XON/XOFF flow control")
	cmd.Flags().IntVar(&opts.IdleTime, "idle-time", 300, "Idle timeout in seconds (60-3600)")
	cmd.Flags().IntVar(&opts.ConnTime, "conn-time", 3600, "Connection timeout in seconds (3600-604800)")
	cmd.Flags().StringVar(&opts.Usage, "usage", "web", "Usage type: web or cli")

	_ = cmd.MarkFlagRequired("device-id")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
