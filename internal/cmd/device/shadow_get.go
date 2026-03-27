package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdShadowGet(f *factory.Factory) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get a shadow document",
		Long:  "Get the full state of a named shadow document, including desired, reported, and delta states.",
		Example: `  # Get shadow document
  incloud device shadow get 507f1f77bcf86cd799439011 --name default

  # Output as YAML
  incloud device shadow get 507f1f77bcf86cd799439011 --name default -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{"name": {name}}
			body, err := client.Get(fmt.Sprintf("/api/v1/devices/%s/shadow", deviceID), q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Shadow name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
