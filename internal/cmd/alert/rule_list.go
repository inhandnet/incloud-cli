package alert

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleListOptions struct {
	Page   int
	Limit  int
	Sort   string
	Fields []string
}

var defaultRuleListFields = []string{"_id", "groupIds", "rules", "notify.channels", "createdAt"}

func NewCmdRuleList(f *factory.Factory) *cobra.Command {
	opts := &RuleListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List alert rules",
		Long:    "List alert rules with pagination.",
		Aliases: []string{"ls"},
		Example: `  # List alert rules
  incloud alert rule list

  # Paginate
  incloud alert rule list --page 2 --limit 50

  # Table output
  incloud alert rule list -o table

  # Table with selected fields
  incloud alert rule list -o table -f _id -f groupIds -f rules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := make(url.Values)
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultRuleListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/alerts/rules", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
