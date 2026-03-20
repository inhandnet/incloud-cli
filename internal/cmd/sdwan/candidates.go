package sdwan

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultCandidateFields = []string{"_id", "deviceName", "serialNumber", "product"}

type candidatesOptions struct {
	Role      string
	NameOrSn  string
	NetworkID string
	Exclude   []string
	Page      int
	Limit     int
	Fields    []string
}

func newCmdCandidates(f *factory.Factory) *cobra.Command {
	opts := &candidatesOptions{}

	cmd := &cobra.Command{
		Use:   "candidates",
		Short: "Find candidate devices for SD-WAN networks",
		Example: `  # Find hub candidates
  incloud sdwan candidates --role hub

  # Find spoke candidates, filter by name or serial number
  incloud sdwan candidates --role spoke --name-or-sn ER805

  # Exclude specific devices
  incloud sdwan candidates --role hub --exclude <id1> --exclude <id2>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))

			body := map[string]interface{}{
				"role": opts.Role,
			}
			if opts.NameOrSn != "" {
				body["nameOrSn"] = opts.NameOrSn
			}
			if opts.NetworkID != "" {
				body["networkId"] = opts.NetworkID
			}
			if len(opts.Exclude) > 0 {
				body["exclusion"] = opts.Exclude
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultCandidateFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			respBody, err := client.Do("POST", apiBase+"/networks/devices/candidates", &api.RequestOptions{
				Query: q,
				Body:  body,
			})
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(respBody, f.IO, output, fields)
		},
	}

	cmd.Flags().StringVar(&opts.Role, "role", "", "Device role: hub or spoke (required)")
	cmd.Flags().StringVar(&opts.NameOrSn, "name-or-sn", "", "Filter by device name or serial number")
	cmd.Flags().StringVar(&opts.NetworkID, "network-id", "", "Filter by network ID")
	cmd.Flags().StringArrayVar(&opts.Exclude, "exclude", nil, "Device IDs to exclude (repeatable)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	_ = cmd.MarkFlagRequired("role")

	return cmd
}
