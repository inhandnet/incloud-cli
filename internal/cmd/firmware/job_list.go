package firmware

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type JobListOptions struct {
	Page     int
	Limit    int
	Sort     string
	Firmware string
	Module   string
	Status   string
	Fields   []string
}

var defaultJobListFields = []string{
	"_id", "status", "document.module", "document.version",
	"progress", "creator.name", "createdAt",
}

// flattenJobList flattens nested jobProcessDetails into a human-readable "progress" field
// for table display: "succeeded/total (failed:N)".
func flattenJobList(body []byte) ([]byte, error) {
	var envelope struct {
		Result []map[string]interface{} `json:"result"`
		Total  int                      `json:"total"`
		Page   int                      `json:"page"`
		Limit  int                      `json:"limit"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing job list response: %w", err)
	}

	for _, item := range envelope.Result {
		details, ok := item["jobProcessDetails"].(map[string]interface{})
		if !ok {
			continue
		}
		total := toInt(details["total"])
		succeeded := toInt(details["succeeded"])
		failed := toInt(details["failed"])
		s := fmt.Sprintf("%d/%d", succeeded, total)
		if failed > 0 {
			s += fmt.Sprintf(" (failed:%d)", failed)
		}
		item["progress"] = s
	}

	return json.Marshal(envelope)
}

func toInt(v interface{}) int {
	n, ok := v.(float64)
	if !ok {
		return 0
	}
	return int(n)
}

func NewCmdJobList(f *factory.Factory) *cobra.Command {
	opts := &JobListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OTA firmware upgrade jobs",
		Long:    "List OTA firmware upgrade jobs with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List recent OTA jobs
  incloud firmware job list

  # Filter by firmware ID
  incloud firmware job list --firmware 6989afd5eeb72121455dc104

  # Filter by status
  incloud firmware job list --status succeeded

  # Filter by module
  incloud firmware job list --module default

  # Paginate and sort
  incloud firmware job list --page 2 --limit 50 --sort createdAt,desc

  # Select fields
  incloud firmware job list -f _id -f status -f document.version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			q.Set("expand", "creator,jobProcessDetails")
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Firmware != "" {
				q.Set("firmwareId", opts.Firmware)
			}
			if opts.Module != "" {
				q.Set("module", opts.Module)
			}
			if opts.Status != "" {
				q.Set("status", strings.ToUpper(opts.Status))
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultJobListFields
			}

			body, err := client.Get("/api/v1/ota/jobs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(flattenJobList))
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware ID")
	cmd.Flags().StringVar(&opts.Module, "module", "", "Filter by module name")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (queued|inprogress|succeeded|canceled)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
