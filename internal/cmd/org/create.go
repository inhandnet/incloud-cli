package org

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type CreateOptions struct {
	Name        string
	Email       string
	Password    string
	Parent      string
	Phone       string
	Description string
	CountryCode string
	BizCategory string
	Locale      string
	Logo        string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization",
		Long: `Create a new organization on the InCloud platform.

For root organizations (no --parent), --email and --password are required
to create the admin user. For sub-organizations, only --name is required.`,
		Example: `  # Create a root organization with admin user
  incloud org create --name "Acme Corp" --email admin@acme.com --password P@ssw0rd

  # Create a sub-organization
  incloud org create --name "Acme Branch" --parent 61259f8f4be3e571fcfa4d75

  # With additional details
  incloud org create --name "Acme Corp" --email admin@acme.com --password P@ssw0rd \
    --country-code US --biz-category TRANSPORTATION --phone "+1234567890"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"name": opts.Name,
			}
			if opts.Email != "" {
				body["email"] = opts.Email
			}
			if opts.Password != "" {
				body["password"] = opts.Password
			}
			if opts.Parent != "" {
				body["parent"] = opts.Parent
			}
			if opts.Phone != "" {
				body["phone"] = opts.Phone
			}
			if opts.Description != "" {
				body["description"] = opts.Description
			}
			if opts.CountryCode != "" {
				body["countryCode"] = opts.CountryCode
			}
			if opts.BizCategory != "" {
				body["bizCategory"] = opts.BizCategory
			}
			if opts.Locale != "" {
				body["locale"] = opts.Locale
			}
			if opts.Logo != "" {
				body["logo"] = opts.Logo
			}

			respBody, err := client.Post("/api/v1/orgs", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			id, name := resultIDName(respBody)
			fmt.Fprintf(f.IO.ErrOut, "Organization %q created. (id: %s)\n", name, id)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Organization name (required, 2-64 chars)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Admin email (required for root orgs)")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Admin password (required for root orgs, 6-64 chars)")
	cmd.Flags().StringVar(&opts.Parent, "parent", "", "Parent organization ID (use 'incloud org list' to find IDs)")
	cmd.Flags().StringVar(&opts.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Organization description (max 256 chars)")
	cmd.Flags().StringVar(&opts.CountryCode, "country-code", "", "Country code (e.g. US, CN)")
	cmd.Flags().StringVar(&opts.BizCategory, "biz-category", "", "Business category (e.g. TRANSPORTATION, INFORMATION_TECHNOLOGY, HEALTHCARE_MEDICINE)")
	cmd.Flags().StringVar(&opts.Locale, "locale", "", "Locale (e.g. en_US, zh_CN)")
	cmd.Flags().StringVar(&opts.Logo, "logo", "", "Logo URL")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
