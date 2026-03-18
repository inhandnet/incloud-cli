package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdShadowDelete(f *factory.Factory) *cobra.Command {
	var (
		name string
		yes  bool
	)

	cmd := &cobra.Command{
		Use:   "delete <device-id>",
		Short: "Delete a shadow document",
		Long:  "Delete a named shadow document from a device.",
		Example: `  # Delete a shadow (with confirmation)
  incloud device shadow delete 507f1f77bcf86cd799439011 --name test

  # Skip confirmation
  incloud device shadow delete 507f1f77bcf86cd799439011 --name test --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if !yes {
				confirmed, err := confirmPrompt(f, fmt.Sprintf("Delete shadow %q from device %s?", name, deviceID))
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IO.ErrOut, "Aborted.")
					return nil
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/api/v1/devices/%s/shadow?name=%s", deviceID, name)
			body, err := client.Delete(path)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				fmt.Fprintf(f.IO.ErrOut, "Shadow %q deleted from device %s.\n", name, deviceID)
				return nil
			}
			return iostreams.FormatOutput(body, f.IO, output, nil,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Shadow name (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
