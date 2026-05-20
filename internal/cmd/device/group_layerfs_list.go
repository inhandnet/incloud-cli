package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type groupLayerfsListOptions struct {
	cmdutil.ListFlags
	Status string
	Expand string
}

func newCmdGroupLayerfsList(f *factory.Factory) *cobra.Command {
	opts := &groupLayerfsListOptions{}

	cmd := &cobra.Command{
		Use:     "list <group-id>",
		Short:   "List filesystem snapshots in a device group",
		Long:    "List filesystem snapshots (layerfs) within a device group.",
		Aliases: []string{"ls"},
		Example: `  # List layerfs in a group
  incloud device group layerfs list 507f1f77bcf86cd799439011

  # Filter by status
  incloud device group layerfs list 507f1f77bcf86cd799439011 --status SUCCEEDED

  # Expand related resources
  incloud device group layerfs list 507f1f77bcf86cd799439011 --expand device,creator`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devicegroups/"+args[0]+"/layerfs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (QUEUED|INPROGRESS|SUCCEEDED|FAILED|CANCELED)")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", "Expand related resources (e.g. device,creator)")

	return cmd
}
