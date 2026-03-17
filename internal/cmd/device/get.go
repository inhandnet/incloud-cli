package device

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get device details",
		Long:  "Get detailed information about a specific device by its ID.",
		Example: `  # Get device details (colorized JSON in TTY)
  incloud device get 507f1f77bcf86cd799439011

  # Only specific fields
  incloud device get 507f1f77bcf86cd799439011 -f name -f serialNumber -f online

  # Table output (KEY/VALUE pairs)
  incloud device get 507f1f77bcf86cd799439011 -o table

  # YAML output
  incloud device get 507f1f77bcf86cd799439011 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var q url.Values
			if len(opts.Fields) > 0 {
				q = url.Values{}
				q.Set("fields", strings.Join(opts.Fields, ","))
			}

			body, err := client.Get("/api/v1/devices/"+deviceID, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, opts.Fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
