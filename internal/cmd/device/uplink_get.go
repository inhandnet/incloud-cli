package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type uplinkGetOptions struct {
	Fields []string
}

func newCmdUplinkGet(f *factory.Factory) *cobra.Command {
	opts := &uplinkGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <uplink-id>",
		Short: "Get uplink details",
		Long:  "Get detailed information for a specific uplink by its ID.",
		Example: `  # Get uplink details
  incloud device uplink get 69b27e3e6e65fb572c20fab4

  # Table output
  incloud device uplink get 69b27e3e6e65fb572c20fab4 -o table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uplinkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/uplinks/"+uplinkID, nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultUplinkDetailFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
