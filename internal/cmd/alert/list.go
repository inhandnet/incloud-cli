package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	After    string
	Before   string
	Status   string
	Priority int
	Device   string
	Group    string
	Type     []string
	Ack      string
	Query    string
}

var defaultListFields = []string{"_id", "type", "priority", "status", "entityName", "ack", "createdAt"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List alerts",
		Long:    "List alerts on the InCloud platform with optional filtering, searching, and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List alerts with default pagination
  incloud alert list

  # Paginate
  incloud alert list --page 2 --limit 50

  # Filter by status
  incloud alert list --status ACTIVE

  # Filter by priority
  incloud alert list --priority 1

  # Filter by device
  incloud alert list --device 507f1f77bcf86cd799439011

  # Filter by time range
  incloud alert list --after 2024-01-01T00:00:00Z --before 2024-01-31T23:59:59Z

  # Filter by alert type (use 'incloud alert rule types' to list available types)
  incloud alert list --type disconnected --type reboot

  # Filter by acknowledgement status
  incloud alert list --ack false

  # Search by entity name
  incloud alert list -q "router"

  # Sort results
  incloud alert list --sort "createdAt,desc"

  # Table output with selected fields
  incloud alert list -o table -f type -f status -f entityName

  # Aggregate alert types with jq
  incloud alert list --limit 100 --jq '[.result[].type] | group_by(.) | map({type: .[0], count: length})'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			var priority *int
			if cmd.Flags().Changed("priority") {
				priority = &opts.Priority
			}
			applyProbeParams(q, opts.After, opts.Before, opts.Status, priority, opts.Device, opts.Group, opts.Type, opts.Ack, opts.Query)

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/alerts", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (ACTIVE/CLOSED)")
	cmd.Flags().IntVar(&opts.Priority, "priority", 0, "Filter by priority level")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Filter by device group ID")
	cmd.Flags().StringArrayVar(&opts.Type, "type", nil, "Filter by alert type (use 'incloud alert rule types' to list available types; can be repeated)")
	cmd.Flags().StringVar(&opts.Ack, "ack", "", "Filter by acknowledgement status (true/false)")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by entity name (fuzzy match)")

	return cmd
}
