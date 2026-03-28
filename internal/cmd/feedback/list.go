package feedback

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type listOptions struct {
	Page       int
	Limit      int
	Sort       string
	App        string
	Resolution string
	Type       string
	Fields     []string
	Count      bool
}

var defaultListFields = []string{"_id", "app", "type", "resolution", "content", "attachments", "createdAt"}

func NewCmdFeedbackList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List feedback entries",
		Long:    "List feedback entries with optional filtering and pagination. Use this command to review submitted feedback and check which entries have attachments.",
		Aliases: []string{"ls"},
		Example: `  # List all feedback
  incloud feedback list

  # List feedback with attachments visible (default fields include attachments)
  incloud feedback list

  # Filter by resolution status
  incloud feedback list --resolution new

  # Filter by type
  incloud feedback list --type suggestion

  # Filter by app
  incloud feedback list --app portal

  # Show only feedback that has attachments (using jq)
  incloud feedback list --jq '[.result[] | select(.attachments | length > 0)]'

  # Count total feedback entries
  incloud feedback list --count

  # Paginate results
  incloud feedback list --page 2 --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Count {
				q.Set("page", "0")
				q.Set("limit", "1")
			}
			if opts.App != "" {
				q.Set("app", opts.App)
			}
			if opts.Resolution != "" {
				q.Set("resolution", opts.Resolution)
			}
			if opts.Type != "" {
				q.Set("type", strings.ToUpper(opts.Type))
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/feedbacks", q)
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

			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithFormatters(iostreams.ColumnFormatters{
					"attachments": formatAttachments,
				}),
			)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.App, "app", "", "Filter by app (e.g. star, portal)")
	cmd.Flags().StringVar(&opts.Resolution, "resolution", "", "Filter by resolution status (new, resolved)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by feedback type (issue, question, comment, suggestion)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().BoolVar(&opts.Count, "count", false, "Only print the total count of matching feedback entries")

	return cmd
}

// formatAttachments renders the attachments array as a human-readable string for table output.
// The input is fmt.Sprint() of []interface{}, which produces "[path1 path2]".
func formatAttachments(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "[]" || s == "[ ]" {
		return ""
	}
	// Strip surrounding brackets from fmt.Sprint output
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	parts := strings.Fields(s)
	var names []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		// Extract filename from objectName path like "2026-03-25/abc123/file.png"
		segments := strings.Split(p, "/")
		name := segments[len(segments)-1]
		if name != "" {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
}
