package user

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type CreateOptions struct {
	Email    string
	Password string
	Name     string
	RoleID   string
	Locale   string
	Labels   []string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user",
		Long:  "Create a new user on the InCloud platform.",
		Example: `  # Create a user with required fields
  incloud user create --email user@example.com --password P@ssw0rd --role-id 5f1e5605fe20f674c2d14d45

  # With display name and locale
  incloud user create --email user@example.com --password P@ssw0rd --role-id 5f1e5605fe20f674c2d14d45 --name "John Doe" --locale en_US

  # With labels
  incloud user create --email user@example.com --password P@ssw0rd --role-id 5f1e5605fe20f674c2d14d45 --label dept=engineering --label level=senior`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"email":    opts.Email,
				"password": opts.Password,
				"roleId":   opts.RoleID,
			}
			if opts.Name != "" {
				body["name"] = opts.Name
			}
			if opts.Locale != "" {
				body["locale"] = opts.Locale
			}
			if len(opts.Labels) > 0 {
				labels, err := parseLabels(opts.Labels)
				if err != nil {
					return err
				}
				body["labels"] = labels
			}

			respBody, err := client.Post("/api/v1/users", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			id, name := resultIDName(respBody)
			fmt.Fprintf(f.IO.ErrOut, "User %q created. (id: %s)\n", name, id)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Email, "email", "", "User email (required)")
	cmd.Flags().StringVar(&opts.Password, "password", "", "User password (required, 6-64 chars)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Display name (defaults to email)")
	cmd.Flags().StringVar(&opts.RoleID, "role-id", "", "Role ID to assign (required; use 'incloud role list' to find IDs)")
	cmd.Flags().StringVar(&opts.Locale, "locale", "", "Locale (e.g. zh_CN, en_US)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("role-id")

	return cmd
}

// parseLabels converts ["key=value", ...] into [{"key":"key","value":"value"}, ...]
func parseLabels(pairs []string) ([]map[string]string, error) {
	labels := make([]map[string]string, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label: %s (expected key=value)", pair)
		}
		labels = append(labels, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}
	return labels, nil
}
