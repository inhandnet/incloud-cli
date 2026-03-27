package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GroupCreateOptions struct {
	Name     string
	Product  string
	Firmware string
}

func newCmdGroupCreate(f *factory.Factory) *cobra.Command {
	opts := &GroupCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device group",
		Long:  "Create a new device group on the InCloud platform.",
		Example: `  # Create a device group
  incloud device group create --name "Edge Routers" --product ER805 --firmware V2.0.26

  # Output as JSON
  incloud device group create --name test --product IR915L --firmware V1.0.0 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"name":     opts.Name,
				"product":  opts.Product,
				"firmware": opts.Firmware,
			}

			respBody, err := client.Post("/api/v1/devicegroups", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			var resp struct {
				Result struct {
					ID   string `json:"_id"`
					Name string `json:"name"`
				} `json:"result"`
			}
			_ = json.Unmarshal(respBody, &resp)
			fmt.Fprintf(f.IO.ErrOut, "Device group %q created. (id: %s)\n", resp.Result.Name, resp.Result.ID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Group name (required, 1-128 chars)")
	cmd.Flags().StringVar(&opts.Product, "product", "", "Product model (required)")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Firmware version (required)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("product")
	_ = cmd.MarkFlagRequired("firmware")

	return cmd
}
