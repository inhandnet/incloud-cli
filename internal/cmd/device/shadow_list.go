package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdShadowList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <device-id>",
		Short: "List shadow document names",
		Long:  "List all named shadow documents for a device.",
		Example: `  # List shadow names
  incloud device shadow list 507f1f77bcf86cd799439011

  # Output as JSON
  incloud device shadow list 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/shadow/names", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")

			switch output {
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			case "table":
				// result is a string array; print one name per line
				var wrapper struct {
					Result []string `json:"result"`
				}
				if err := json.Unmarshal(body, &wrapper); err != nil {
					return fmt.Errorf("parsing response: %w", err)
				}
				if len(wrapper.Result) == 0 {
					fmt.Fprintln(f.IO.ErrOut, "No shadow documents found.")
					return nil
				}
				for _, name := range wrapper.Result {
					fmt.Fprintln(f.IO.Out, name)
				}
			default:
				fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
			}
			return nil
		},
	}

	return cmd
}
