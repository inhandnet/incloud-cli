package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type clientConnectionsOptions struct {
	cmdutil.ListFlags
	After  string
	Before string
}

func newCmdClientConnections(f *factory.Factory) *cobra.Command {
	opts := &clientConnectionsOptions{}

	cmd := &cobra.Command{
		Use:   "connections <client-id>",
		Short: "List connections for a touch client",
		Long:  "List connection history for a specific touch client.",
		Example: `  # List connections
  incloud touch client connections 507f1f77bcf86cd799439011

  # Filter by time range
  incloud touch client connections 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00Z --before 2024-12-31T23:59:59Z`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/touch/clients/"+args[0]+"/connections", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.After, "after", "", "Filter connections after this time (ISO 8601)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter connections before this time (ISO 8601)")

	return cmd
}
