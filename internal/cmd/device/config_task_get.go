package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigTaskGet(f *factory.Factory) *cobra.Command {
	var expand string

	cmd := &cobra.Command{
		Use:   "get <job-id>",
		Short: "Get a CLI configuration task",
		Long:  "Get details of a CLI configuration task by its job ID.",
		Example: `  # Get task details
  incloud device config task get 507f1f77bcf86cd799439011

  # Get task details with job expansion
  incloud device config task get 507f1f77bcf86cd799439011 --expand cliJob`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if expand != "" {
				q.Set("expand", expand)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/cli-configs/"+args[0], q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&expand, "expand", "", "Expand related resources (e.g. cliJob)")

	return cmd
}
