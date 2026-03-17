package activity

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page   int
	Limit  int
	Sort   string
	After  string
	Before string
	App    string
	Action string
	Actor  string
	Fields []string
	Count  bool
}

var defaultListFields = []string{"_id", "app", "action", "actor.name", "entity.type", "entity.name", "ipAddress", "timestamp"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List activity logs",
		Long:    "List audit activity logs on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List recent activity logs
  incloud activity list

  # Paginate
  incloud activity list --page 2 --limit 50

  # Filter by action type
  incloud activity list --action device_created

  # Filter by actor
  incloud activity list --actor 5f1e5605cf562757b857a7b9

  # Filter by time range
  incloud activity list --after 2024-01-01T00:00:00 --before 2024-01-31T23:59:59

  # Filter by application
  incloud activity list --app nezha

  # Sort by timestamp ascending
  incloud activity list --sort "timestamp,asc"

  # Table output with selected fields
  incloud activity list -o table -f action -f actor.name -f entity.name

  # Count activity logs in a time range
  incloud activity list --after 2024-01-01T00:00:00 --count`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := make(url.Values)
			if opts.Count {
				q.Set("page", "0")
				q.Set("limit", "1")
			} else {
				q.Set("page", strconv.Itoa(opts.Page-1))
				q.Set("limit", strconv.Itoa(opts.Limit))
			}
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.After != "" {
				q.Set("from", opts.After)
			}
			if opts.Before != "" {
				q.Set("to", opts.Before)
			}
			if opts.App != "" {
				q.Set("app", opts.App)
			}
			if opts.Action != "" {
				q.Set("action", opts.Action)
			}
			if opts.Actor != "" {
				q.Set("actor", opts.Actor)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultListFields
			}

			body, err := client.Get("/api/v1/audit/logs", q)
			if err != nil {
				return err
			}

			if opts.Count {
				var envelope struct {
					Total int64 `json:"total"`
				}
				if err := json.Unmarshal(body, &envelope); err != nil {
					return fmt.Errorf("parsing response: %w", err)
				}
				fmt.Fprintln(f.IO.Out, envelope.Total)
				return nil
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "timestamp,desc")`)
	cmd.Flags().StringVar(&opts.After, "after", "", "Filter logs after this time (e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter logs before this time (e.g. 2024-01-31T23:59:59)")
	cmd.Flags().StringVar(&opts.App, "app", "", "Filter by application name")
	cmd.Flags().StringVar(&opts.Action, "action", "", "Filter by action type (e.g. device_created, device_deleted)")
	cmd.Flags().StringVar(&opts.Actor, "actor", "", "Filter by actor ID")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().BoolVar(&opts.Count, "count", false, "Only print the total count of matching logs")

	return cmd
}
