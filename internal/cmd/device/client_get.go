package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdClientGet(f *factory.Factory) *cobra.Command {
	var fields []string

	cmd := &cobra.Command{
		Use:   "get <client-id>",
		Short: "Get client details",
		Long:  "Display detailed information about a specific connected client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/network/clients/"+args[0], nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
