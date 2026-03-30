package cmdutil

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// NewQuery builds url.Values from well-known list flags (page, limit, sort, fields, expand).
// Only flags that are registered on the command are read; missing flags are silently skipped.
// defaultFields is used when the user doesn't specify --fields and output is table mode.
func NewQuery(cmd *cobra.Command, defaultFields []string) url.Values {
	q := url.Values{}
	flags := cmd.Flags()

	// page: CLI 1-based → API 0-based
	if page, err := flags.GetInt("page"); err == nil {
		q.Set("page", strconv.Itoa(page-1))
	}

	// limit
	if limit, err := flags.GetInt("limit"); err == nil {
		q.Set("limit", strconv.Itoa(limit))
	}

	// sort: only set if explicitly provided
	if flags.Changed("sort") {
		if sort, err := flags.GetString("sort"); err == nil {
			q.Set("sort", sort)
		}
	}

	// fields: user-specified takes precedence, otherwise apply defaults for table output
	if flags.Changed("fields") {
		if fields, err := flags.GetStringSlice("fields"); err == nil && len(fields) > 0 {
			q.Set("fields", strings.Join(fields, ","))
		}
	} else if len(defaultFields) > 0 {
		output, _ := flags.GetString("output")
		if output == "" || output == "table" {
			q.Set("fields", strings.Join(defaultFields, ","))
		}
	}

	// expand: set if non-empty (covers both Changed and non-empty defaults)
	if expand, err := flags.GetStringSlice("expand"); err == nil && len(expand) > 0 {
		q.Set("expand", strings.Join(expand, ","))
	}

	return q
}

// ListFlags holds the common pagination/sorting/fields/expand flags for list commands.
// Embed this struct in your command's options struct:
//
//	type ListOptions struct {
//	    cmdutil.ListFlags
//	    Query  string
//	    Status string
//	}
//
// Then call opts.ListFlags.Register(cmd) to register page/limit/sort/fields,
// and opts.ListFlags.RegisterExpand(cmd) for commands that also support --expand.
type ListFlags struct {
	Page   int
	Limit  int
	Sort   string
	Fields []string
	Expand []string
}

// Register registers the standard list flags (page, limit, sort, fields) on cmd.
func (lf *ListFlags) Register(cmd *cobra.Command) {
	cmd.Flags().IntVar(&lf.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&lf.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&lf.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringSliceVarP(&lf.Fields, "fields", "f", nil, "Fields to return and display")
}

// RegisterExpand registers the --expand flag on cmd.
// Call this in addition to Register for commands that support resource expansion.
func (lf *ListFlags) RegisterExpand(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&lf.Expand, "expand", nil, "Expand related resources")
}
