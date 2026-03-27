package tunnel

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdTunnelGet(f *factory.Factory) *cobra.Command {
	var (
		output string
		fields []string
	)

	cmd := &cobra.Command{
		Use:   "get <tunnel-id>",
		Short: "Get tunnel details",
		Long:  "Show details of an active tunnel, including protocol, token, and metadata.",
		Example: `  # Get tunnel details
  incloud tunnel get nhddruohqziaxu6nvvibfwwn

  # Get as JSON
  incloud tunnel get nhddruohqziaxu6nvvibfwwn -o json

  # Get specific fields
  incloud tunnel get nhddruohqziaxu6nvvibfwwn -o json --fields id,protocol,token`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			endpoint := fmt.Sprintf("/api/v1/ngrok/tunnels/%s", args[0])
			body, err := client.Get(endpoint, nil)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format: json, yaml")
	cmd.Flags().StringSliceVar(&fields, "fields", nil, "Fields to display")

	return cmd
}
