package feedback

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type listOptions struct {
	cmdutil.ListFlags
	App        string
	Resolution string
	Type       string
}

var defaultListFields = []string{"_id", "app", "type", "resolution", "reply", "content", "attachments", "createdAt"}

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

  # Paginate results
  incloud feedback list --page 2 --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
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

			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithFormatters(iostreams.ColumnFormatters{
					"attachments": formatAttachments,
				}),
			)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.App, "app", "", "Filter by app (e.g. star, portal)")
	cmd.Flags().StringVar(&opts.Resolution, "resolution", "", "Filter by resolution status (new, resolved)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by feedback type (issue, question, comment, suggestion)")
	opts.ListFlags.RegisterExpand(cmd, "user", "org")

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
