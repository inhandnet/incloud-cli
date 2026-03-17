package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page     int
	Limit    int
	Sort     string
	After    string
	Before   string
	Status   string
	Priority int
	Device   string
	Group    string
	Type     []string
	Ack      string
	Query    string
	Fields   []string
	Count    bool
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
  incloud alert list --after 2024-01-01T00:00:00 --before 2024-01-31T23:59:59

  # Filter by alert type
  incloud alert list --type offline --type reboot

  # Filter by acknowledgement status
  incloud alert list --ack false

  # Search by entity name
  incloud alert list -q "router"

  # Sort results
  incloud alert list --sort "createdAt,desc"

  # Table output with selected fields
  incloud alert list -o table -f type -f status -f entityName

  # Count unacknowledged alerts
  incloud alert list --ack false --count

  # Count active alerts for a device
  incloud alert list --status ACTIVE --device 507f1f77bcf86cd799439011 --count`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/alerts")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
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
			var priority *int
			if cmd.Flags().Changed("priority") {
				priority = &opts.Priority
			}
			applyProbeParams(q, opts.After, opts.Before, opts.Status, priority, opts.Device, opts.Group, opts.Type, opts.Ack, opts.Query)

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
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

			switch output {
			case "table":
				if err := iostreams.FormatTable(body, f.IO, fields); err != nil {
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

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.After, "after", "", "Filter alerts after this time (e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter alerts before this time (e.g. 2024-01-31T23:59:59)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (ACTIVE/CLOSED)")
	cmd.Flags().IntVar(&opts.Priority, "priority", 0, "Filter by priority level")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Filter by device group ID")
	cmd.Flags().StringArrayVar(&opts.Type, "type", nil, "Filter by alert type (can be repeated)")
	cmd.Flags().StringVar(&opts.Ack, "ack", "", "Filter by acknowledgement status (true/false)")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by entity name (fuzzy match)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().BoolVar(&opts.Count, "count", false, "Only print the total count of matching alerts")

	return cmd
}
