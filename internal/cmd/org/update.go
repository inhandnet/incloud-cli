package org

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type UpdateOptions struct {
	Name         string
	Description  string
	Email        string
	ContactEmail string
	Contactor    string
	Phone        string
	CountryCode  string
	BizCategory  string
	Labels       []string
}

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an organization",
		Long:  "Update an existing organization on the InCloud platform.",
		Example: `  # Update name
  incloud org update 61259f8f4be3e571fcfa4d75 --name "New Name"

  # Update multiple fields
  incloud org update 61259f8f4be3e571fcfa4d75 --name "New Name" --description "Updated" --country-code CN

  # Set labels
  incloud org update 61259f8f4be3e571fcfa4d75 --label region=east --label tier=premium`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			output, _ := cmd.Flags().GetString("output")

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
			if cmd.Flags().Changed("contact-email") {
				body["contactEmail"] = opts.ContactEmail
			}
			if cmd.Flags().Changed("contactor") {
				body["contactor"] = opts.Contactor
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
			if cmd.Flags().Changed("label") {
				labels, err := parseLabels(opts.Labels)
				if err != nil {
					return err
				}
				body["labels"] = labels
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --description, --email, --contact-email, --contactor, --phone, --country-code, --biz-category, or --label")
			}

			respBody, err := client.Put("/api/v1/orgs/"+id, body)
			if err != nil {
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			respID, name := api.ResultIDName(respBody)
			if name == "" {
				name = id
			}
			fmt.Fprintf(f.IO.ErrOut, "Organization %q (%s) updated.\n", name, respID)

			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Organization name")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Description (max 256 chars)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Organization email")
	cmd.Flags().StringVar(&opts.ContactEmail, "contact-email", "", "Contact email")
	cmd.Flags().StringVar(&opts.Contactor, "contactor", "", "Contact person name")
	cmd.Flags().StringVar(&opts.Phone, "phone", "", "Phone number")
	cmd.Flags().StringVar(&opts.CountryCode, "country-code", "", "Country code (e.g. US, CN)")
	cmd.Flags().StringVar(&opts.BizCategory, "biz-category", "", "Business category (e.g. TRANSPORTATION, INFORMATION_TECHNOLOGY, HEALTHCARE_MEDICINE)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")

	return cmd
}
