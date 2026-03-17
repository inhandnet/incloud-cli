package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdAssign(f *factory.Factory) *cobra.Command {
	var group string

	cmd := &cobra.Command{
		Use:   "assign <device-id>",
		Short: "Assign a device to a device group",
		Example: `  # Assign device to a group
  incloud device assign 507f1f77bcf86cd799439011 --group 653b1ff2a84e171614d88695`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"deviceId":      deviceID,
						"deviceGroupId": group,
					},
				},
			}

			_, err = client.Put("/api/v1/devices/move", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Device %s assigned to group %s.\n", deviceID, group)
			return nil
		},
	}

	cmd.Flags().StringVar(&group, "group", "", "Target device group ID (required)")
	_ = cmd.MarkFlagRequired("group")

	return cmd
}
