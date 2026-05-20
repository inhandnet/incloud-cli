package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgePin(f *factory.Factory) *cobra.Command {
	var version string
	var unpin bool

	cmd := &cobra.Command{
		Use:   "pin <device-id>",
		Short: "Pin a device to a specific project version",
		Long: `Pin an edge device to a specific project version so it does not follow group-wide deployments.
Use --unpin to remove the pin and let the device follow group deployments again.`,
		Example: `  # Pin device to version 0.1.3
  incloud device edge pin 507f1f77bcf86cd799439011 --version 0.1.3

  # Unpin device (follow group deployments)
  incloud device edge pin 507f1f77bcf86cd799439011 --unpin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{}
			if unpin {
				reqBody["version"] = nil
			} else {
				reqBody["version"] = version
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devices/"+args[0]+"/pin", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			if unpin {
				fmt.Fprintf(f.IO.ErrOut, "Device %s unpinned.\n", args[0])
			} else {
				fmt.Fprintf(f.IO.ErrOut, "Device %s pinned to version %s.\n", args[0], version)
			}
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Project version to pin to")
	cmd.Flags().BoolVar(&unpin, "unpin", false, "Remove pin and follow group deployments")
	cmd.MarkFlagsMutuallyExclusive("version", "unpin")

	return cmd
}
