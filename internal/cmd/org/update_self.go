package org

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type UpdateSelfOptions struct {
	Name        string
	Description string
	Email       string
	Phone       string
	CountryCode string
	BizCategory string
	Contactor   string
}

func NewCmdUpdateSelf(f *factory.Factory) *cobra.Command {
	opts := &UpdateSelfOptions{}

	cmd := &cobra.Command{
		Use:   "update-self",
		Short: "Update current organization",
		Long:  "Update the organization that the current user belongs to.",
		Example: `  # Update organization name
  incloud org update-self --name "New Name"

  # Update multiple fields
  incloud org update-self --name "New Name" --description "Updated desc" --phone "+1234567890"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})

			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("description") {
				body["description"] = opts.Description
			}
			if cmd.Flags().Changed("email") {
				body["email"] = opts.Email
			}
			if cmd.Flags().Changed("phone") {
				body["phone"] = opts.Phone
			}
			if cmd.Flags().Changed("country-code") {
				body["countryCode"] = opts.CountryCode
			}
			if cmd.Flags().Changed("biz-category") {
				body["bizCategory"] = opts.BizCategory
			}
			if cmd.Flags().Changed("contactor") {
				body["contactor"] = opts.Contactor
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --description, --email, --phone, --country-code, --biz-category, or --contactor")
			}

			respBody, err := client.Put("/api/v1/orgs/self", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			id, name := resultIDName(respBody)
			if name == "" {
				name = id
			}
			fmt.Fprintf(f.IO.ErrOut, "Organization %q (%s) updated.\n", name, id)

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Organization name")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Description (max 256 chars)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Organization email")
	cmd.Flags().StringVar(&opts.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&opts.CountryCode, "country-code", "", "Country code (e.g. US, CN)")
	cmd.Flags().StringVar(&opts.BizCategory, "biz-category", "", "Business category (e.g. TRANSPORTATION, INFORMATION_TECHNOLOGY, HEALTHCARE_MEDICINE)")
	cmd.Flags().StringVar(&opts.Contactor, "contactor", "", "Contact person name")

	return cmd
}
