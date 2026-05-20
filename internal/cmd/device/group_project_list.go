package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type groupProjectListOptions struct {
	cmdutil.ListFlags
	Published string
	Version   string
	Expand    string
}

func newCmdGroupProjectList(f *factory.Factory) *cobra.Command {
	opts := &groupProjectListOptions{}

	cmd := &cobra.Command{
		Use:     "list <group-id>",
		Short:   "List project versions",
		Long:    "List project versions in a device group.",
		Aliases: []string{"ls"},
		Example: `  # List all projects
  incloud device group project list 507f1f77bcf86cd799439011

  # Filter by published status
  incloud device group project list 507f1f77bcf86cd799439011 --published true

  # Filter by version
  incloud device group project list 507f1f77bcf86cd799439011 --version 0.1

  # Expand related resources
  incloud device group project list 507f1f77bcf86cd799439011 --expand creator,edge-layerfs`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Published != "" {
				q.Set("published", opts.Published)
			}
			if opts.Version != "" {
				q.Set("version", opts.Version)
			}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devicegroups/"+args[0]+"/projects", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Published, "published", "", "Filter by published status (true|false)")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Filter by version (LIKE search)")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", "Expand related resources (e.g. creator,edge-layerfs)")

	return cmd
}
