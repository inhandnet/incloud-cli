package oobm

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmConnectOptions struct {
	Services []string
}

func NewCmdOobmConnect(f *factory.Factory) *cobra.Command {
	opts := &OobmConnectOptions{}

	cmd := &cobra.Command{
		Use:   "connect <id>",
		Short: "Connect an OOBM resource",
		Long: `Connect an Out-of-Band Management resource to establish remote access tunnels.

Without --service, all services defined on the resource are connected.
Use --service to connect only specific services (protocol:port[:usage] format).

The response includes tunnelId and token. Use 'incloud tunnel forward' to
forward the tunnel to a local port for direct ssh/telnet/curl access.`,
		Example: `  # Connect all services on the resource
  incloud oobm connect 507f1f77bcf86cd799439011

  # Connect only SSH service
  incloud oobm connect 507f1f77bcf86cd799439011 --service ssh:22:cli

  # Forward to local port for direct access
  incloud oobm connect 507f1f77bcf86cd799439011 --service ssh:22:cli -o json
  incloud tunnel forward <tunnelId> --token <token> --port 2222
  ssh root@localhost -p 2222`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var services []map[string]any
			if len(opts.Services) > 0 {
				services, err = parseServices(opts.Services)
				if err != nil {
					return err
				}
			} else {
				// Auto-resolve: fetch resource and use all its services.
				res, err := getOobmResource(client, id)
				if err != nil {
					return err
				}
				services = res.Services
				if len(services) == 0 {
					return fmt.Errorf("resource %q has no services defined", id)
				}
			}

			endpoint := fmt.Sprintf("/api/v1/oobm/resources/%s/connect", id)
			results := make([]json.RawMessage, 0, len(services))

			for _, svc := range services {
				respBody, err := client.Post(endpoint, svc)
				if err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Failed to connect %s: %v\n", serviceLabel(svc), err)
					continue
				}
				results = append(results, respBody)

				var status struct {
					Result struct {
						Protocol string `json:"protocol"`
						URL      string `json:"url"`
					} `json:"result"`
				}
				_ = json.Unmarshal(respBody, &status)
				fmt.Fprintf(f.IO.ErrOut, "Connected %s: %s\n", status.Result.Protocol, status.Result.URL)
			}

			if len(results) == 0 {
				return fmt.Errorf("no services connected successfully")
			}

			combined, _ := json.Marshal(results)
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(combined, f.IO, output)
		},
	}

	cmd.Flags().StringArrayVar(&opts.Services, "service", nil, "Service in protocol:port[:usage] format (omit to connect all)")

	return cmd
}
