package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdClientGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <client-id>",
		Short: "Get a touch client",
		Long:  "Get details of a remote access client by its ID.",
		Example: `  # Get client details
  incloud touch client get 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/touch/clients/"+args[0], q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
