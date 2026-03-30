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

// ListOpts holds the common pagination/sorting/fields flags for list commands.
// Embed this struct in your command's options struct:
//
//	type MyListOptions struct {
//	    cmdutil.ListOpts
//	    MyField string
//	}
//
// Then call RegisterListFlags(cmd, &opts.ListOpts) in NewCmd.
type ListOpts struct {
	Page   int
	Limit  int
	Sort   string
	Fields []string
}

// RegisterListFlags registers the standard list flags (page, limit, sort, fields)
// on cmd, binding them to opts.
func RegisterListFlags(cmd *cobra.Command, opts *ListOpts) {
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
}

// RegisterExpandFlag registers the --expand flag on cmd, binding it to expand.
// Only call this for commands that support resource expansion.
func RegisterExpandFlag(cmd *cobra.Command, expand *[]string) {
	cmd.Flags().StringSliceVar(expand, "expand", nil, "Expand related resources")
}
