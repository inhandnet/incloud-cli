package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get device group details",
		Example: `  # Get a device group
  incloud device group get 507f1f77bcf86cd799439011

  # Output as JSON
  incloud device group get 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/devicegroups/"+args[0], nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
