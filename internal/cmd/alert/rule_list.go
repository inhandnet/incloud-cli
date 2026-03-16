package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleListOptions struct {
	Page    int
	Limit   int
	Sort    string
	Columns []string
}

var defaultRuleListColumns = []string{"_id", "groupIds", "rules", "notify.channels", "createdAt"}

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
  incloud alert rule list --page 1 --limit 50

  # Table output
  incloud alert rule list -o table

  # Table with selected columns
  incloud alert rule list -o table -c _id -c groupIds -c rules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			u, err := url.Parse(ctx.Host + "/api/v1/alerts/rules")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			q.Set("page", strconv.Itoa(opts.Page))
			q.Set("size", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "table":
				columns := opts.Columns
				if len(columns) == 0 && f.IO.IsStdoutTTY() {
					columns = defaultRuleListColumns
				}
				if err := iostreams.FormatTable(body, f.IO, columns); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 0, "Page number (default 0)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page (default 20)")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringArrayVarP(&opts.Columns, "column", "c", nil, "Columns to show in table output")

	return cmd
}
