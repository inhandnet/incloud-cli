package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClientSetPosReady(f *factory.Factory) *cobra.Command {
	var mac string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "set-pos-ready <device-id>",
		Short: "Set POS Ready status for a client",
		Long:  "Enable or disable POS Ready status for a client on the specified device.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Enable POS Ready for a client
  incloud device client set-pos-ready DEVICE_ID --mac FC:5C:EE:8C:90:93 --enabled

  # Disable POS Ready
  incloud device client set-pos-ready DEVICE_ID --mac FC:5C:EE:8C:90:93 --enabled=false

  # Short form
  incloud dev client set-pos-ready DEVICE_ID --mac FC:5C:EE:8C:90:93 --enabled`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			deviceID := args[0]
			body := map[string]any{
				"mac":     mac,
				"enabled": enabled,
			}
			_, err = client.Post("/api/v1/network/devices/"+deviceID+"/clients/pos-ready", body)
			if err != nil {
				return err
			}

			state := "enabled"
			if !enabled {
				state = "disabled"
			}
			fmt.Fprintf(f.IO.ErrOut, "POS Ready %s for client %s on device %s.\n", state, mac, deviceID)
			return nil
		},
	}

	cmd.Flags().StringVar(&mac, "mac", "", "Client MAC address (required)")
	cmd.Flags().BoolVar(&enabled, "enabled", false, "Enable or disable POS Ready")
	_ = cmd.MarkFlagRequired("mac")
	_ = cmd.MarkFlagRequired("enabled")

	return cmd
}
