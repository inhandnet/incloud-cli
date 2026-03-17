package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
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

			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			reqURL := actx.Host + "/api/v1/devices/move"
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqURL, bytes.NewReader(jsonBytes))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				fmt.Fprintln(f.IO.ErrOut, string(respBody))
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}

			fmt.Fprintf(f.IO.ErrOut, "Device %s assigned to group %s.\n", deviceID, group)
			return nil
		},
	}

	cmd.Flags().StringVar(&group, "group", "", "Target device group ID (required)")
	_ = cmd.MarkFlagRequired("group")

	return cmd
}
