package device

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigGet(f *factory.Factory) *cobra.Command {
	var (
		module string
		layers []string
	)

	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get device configuration",
		Long: `Get the device configuration. By default returns the fully merged configuration
(combining default, group, and individual layers). Use --layer to view specific
configuration layers instead (actual, target, pending, group, individual).`,
		Example: `  # Get merged configuration (default)
  incloud device config get 507f1f77bcf86cd799439011

  # Get only the actual layer
  incloud device config get 507f1f77bcf86cd799439011 --layer actual

  # Get actual and pending layers
  incloud device config get 507f1f77bcf86cd799439011 --layer actual --layer pending

  # YAML output
  incloud device config get 507f1f77bcf86cd799439011 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if module != "" {
				q.Set("module", module)
			}

			var body []byte
			if len(layers) > 0 {
				// Layered view: GET /devices/{id}/config?fields=...
				q.Set("fields", strings.Join(layers, ","))
				body, err = client.Get("/api/v1/devices/"+deviceID+"/config", q)
			} else {
				// Merged view (default): GET /devices/{id}/merge-config
				body, err = client.Get("/api/v1/devices/"+deviceID+"/merge-config", q)
			}
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, nil,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().StringArrayVar(&layers, "layer", nil, "Config layers to return: actual, target, pending, group, individual (can be repeated)")

	return cmd
}
