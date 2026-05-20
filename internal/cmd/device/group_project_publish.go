package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupProjectPublish(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish <group-id> <project-id>",
		Short: "Publish a project version",
		Long:  "Publish a project version, making it immutable and eligible for deployment.",
		Example: `  # Publish a project
  incloud device group project publish 507f1f77bcf86cd799439011 653b1ff2a84e171614d88695`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devicegroups/"+args[0]+"/projects/"+args[1]+"/publish", nil)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Project %s published.\n", args[1])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
