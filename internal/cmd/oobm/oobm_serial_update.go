package oobm

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmSerialUpdateOptions struct {
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

func NewCmdOobmSerialUpdate(f *factory.Factory) *cobra.Command {
	opts := &OobmSerialUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an OOBM serial port configuration",
		Long:  "Update an OOBM serial port configuration. This is a full replacement of all fields.",
		Example: `  # Update serial port settings
  incloud oobm serial update 507f1f77bcf86cd799439011 \
    --device-id 607f1f77bcf86cd799439022 \
    --name "Console" --speed 115200 --usage cli`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

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

			respBody, err := client.Put("/api/v1/oobm/serials/"+id, body)
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
			fmt.Fprintf(f.IO.ErrOut, "OOBM serial %q (%s) updated.\n", resp.Name, resp.ID)

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
