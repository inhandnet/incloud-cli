package device

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgeList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <id>[,<id2>,...]",
		Short:   "List edge devices by IDs",
		Long:    "Retrieve edge-specific properties for one or more devices by their IDs.",
		Aliases: []string{"ls"},
		Example: `  # Get edge info for a single device
  incloud device edge list 507f1f77bcf86cd799439011

  # Get edge info for multiple devices
  incloud device edge list 507f1f77bcf86cd799439011,653b1ff2a84e171614d88695`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			ids := strings.Split(args[0], ",")
			reqBody := map[string]interface{}{
				"ids": ids,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devices/list", reqBody)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
