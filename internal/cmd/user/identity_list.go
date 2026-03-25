package user

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultIdentityListFields = []string{"oid", "orgName", "type", "roles", "expiresAt"}

func NewCmdIdentityList(f *factory.Factory) *cobra.Command {
	var (
		page    int
		limit   int
		orgName string
		fields  []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List identities across organizations",
		Long: `List the current user's identities (roles) across all accessible organizations.

Each identity represents the user's membership in an organization, including
the assigned roles and optional expiration date for external organizations.`,
		Example: `  # List all identities
  incloud user identity list

  # Filter by organization name
  incloud user identity list --org-name "Acme"

  # JSON output
  incloud user identity list -o json

  # Table with custom fields
  incloud user identity list -f oid -f orgName -f roles`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(page-1))
			q.Set("limit", strconv.Itoa(limit))
			if orgName != "" {
				q.Set("orgName", orgName)
			}

			body, err := client.Get("/api/v1/user/identities", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			cols := fields
			if len(cols) == 0 && output == "table" {
				cols = defaultIdentityListFields
			}

			return iostreams.FormatOutput(body, f.IO, output, cols,
				iostreams.WithTransform(flattenIdentities))
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&orgName, "org-name", "", "Filter by organization name")
	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}

// flattenIdentities transforms the identities response for table rendering.
// It adds a "type" field (internal/external) and flattens roles to a comma-separated string.
func flattenIdentities(data []byte) ([]byte, error) {
	// Parse envelope: body may be {"result": [...]} or a bare array.
	var envelope struct {
		Result []map[string]any `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Result == nil {
		return data, nil
	}

	for _, item := range envelope.Result {
		oid, okOid := item["oid"].(string)
		userOid, okUserOid := item["userOid"].(string)
		if okOid && okUserOid && oid == userOid {
			item["type"] = "internal"
		} else {
			item["type"] = "external"
		}

		if roles, ok := item["roles"].([]any); ok {
			names := make([]string, 0, len(roles))
			for _, r := range roles {
				if rm, ok := r.(map[string]any); ok {
					if name, ok := rm["roleName"].(string); ok {
						names = append(names, name)
					}
				}
			}
			item["roles"] = strings.Join(names, ", ")
		}
	}

	return json.Marshal(map[string]any{"result": envelope.Result})
}
