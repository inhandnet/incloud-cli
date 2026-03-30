package user

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type IdentityListOptions struct {
	cmdutil.ListFlags
	OrgName string
}

func NewCmdIdentityList(f *factory.Factory) *cobra.Command {
	opts := &IdentityListOptions{}

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

			q := cmdutil.NewQuery(cmd, nil)
			if opts.OrgName != "" {
				q.Set("orgName", opts.OrgName)
			}

			body, err := client.Get("/api/v1/user/identities", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(flattenIdentities))
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.OrgName, "org-name", "", "Filter by organization name")
	opts.ListFlags.RegisterExpand(cmd)

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
