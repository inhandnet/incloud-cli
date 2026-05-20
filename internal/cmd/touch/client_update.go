package touch

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdClientUpdate(f *factory.Factory) *cobra.Command {
	var (
		name       string
		ip         string
		serialJSON string
	)

	cmd := &cobra.Command{
		Use:   "update <client-id>",
		Short: "Update a touch client",
		Long:  "Update a remote access client's name, IP address, or serial configuration.",
		Example: `  # Update name
  incloud touch client update 507f1f77bcf86cd799439011 --name new-name

  # Update IP address
  incloud touch client update 507f1f77bcf86cd799439011 --ip 192.168.1.200`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{}
			if name != "" {
				reqBody["name"] = name
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

			body, err := apiClient.Put("/api/v1/touch/clients/"+args[0], reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Touch client %s updated.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New client name (1-128 chars)")
	cmd.Flags().StringVar(&ip, "ip", "", "New IP address for ETHERNET type")
	cmd.Flags().StringVar(&serialJSON, "serial", "", "New serial configuration as JSON")

	return cmd
}
