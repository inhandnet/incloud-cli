package user

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type UpdateOptions struct {
	Name    string
	Email   string
	Contact string
	RoleID  string
	Locale  string
	Labels  []string
}

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user",
		Long:  "Update an existing user on the InCloud platform.",
		Example: `  # Update display name
  incloud user update 507f1f77bcf86cd799439011 --name "New Name"

  # Update email and contact
  incloud user update 507f1f77bcf86cd799439011 --email new@example.com --contact "+1234567890"

  # Update role
  incloud user update 507f1f77bcf86cd799439011 --role-id 5f1e5605fe20f674c2d14d45

  # Update labels
  incloud user update 507f1f77bcf86cd799439011 --label dept=engineering --label level=senior`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})

			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("email") {
				body["email"] = opts.Email
			}
			if cmd.Flags().Changed("contact") {
				body["contact"] = opts.Contact
			}
			if cmd.Flags().Changed("role-id") {
				body["roleId"] = opts.RoleID
			}
			if cmd.Flags().Changed("locale") {
				body["locale"] = opts.Locale
			}
			if cmd.Flags().Changed("label") {
				labels, err := parseLabels(opts.Labels)
				if err != nil {
					return err
				}
				body["labels"] = labels
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --email, --contact, --role-id, --locale, or --label")
			}

			// GET the user to obtain currentOid
			getBody, err := client.Get("/api/v1/users/"+id, nil)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			var getResp struct {
				Result struct {
					OID string `json:"oid"`
				} `json:"result"`
			}
			if err := json.Unmarshal(getBody, &getResp); err != nil {
				return fmt.Errorf("failed to parse user response: %w", err)
			}
			body["currentOid"] = getResp.Result.OID

			respBody, err := client.Put("/api/v1/users/"+id, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			respID, name := resultIDName(respBody)
			if name == "" {
				name = id
			}
			fmt.Fprintf(f.IO.ErrOut, "User %q (%s) updated.\n", name, respID)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Display name")
	cmd.Flags().StringVar(&opts.Email, "email", "", "User email")
	cmd.Flags().StringVar(&opts.Contact, "contact", "", "Contact information")
	cmd.Flags().StringVar(&opts.RoleID, "role-id", "", "Role ID to assign (use 'incloud role list' to find IDs)")
	cmd.Flags().StringVar(&opts.Locale, "locale", "", "Locale (e.g. zh_CN, en_US)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")

	return cmd
}
