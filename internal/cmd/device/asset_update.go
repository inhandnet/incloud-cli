package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdAssetUpdate(f *factory.Factory) *cobra.Command {
	var (
		name               string
		category           string
		status             string
		number             string
		warrantyExpiration string
		notes              string
	)

	cmd := &cobra.Command{
		Use:   "update <asset-id>",
		Short: "Update a network asset",
		Long:  "Update the properties of an existing network asset. Name, category, and status are required by the API; omitting them will result in a server-side validation error.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Update status only (name and category still required by API)
  incloud device asset update 507f1f77bcf86cd799439011 \
    --name "Office Router" --category router --status decommissioned

  # Update just the notes
  incloud device asset update 507f1f77bcf86cd799439011 --notes "moved to 3rd floor"

  # Clear warranty expiration
  incloud device asset update 507f1f77bcf86cd799439011 --warranty-expiration ""`,
		RunE: func(cmd *cobra.Command, args []string) error {
			assetID := args[0]

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				body["name"] = name
			}
			if cmd.Flags().Changed("category") {
				body["category"] = category
			}
			if cmd.Flags().Changed("status") {
				body["status"] = status
			}
			if cmd.Flags().Changed("number") {
				body["number"] = number
			}
			if cmd.Flags().Changed("warranty-expiration") {
				if warrantyExpiration == "" {
					body["warrantyExpiration"] = nil
				} else {
					body["warrantyExpiration"] = warrantyExpiration
				}
			}
			if cmd.Flags().Changed("notes") {
				body["notes"] = notes
			}

			if len(body) == 0 {
				return fmt.Errorf("at least one field must be specified")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			respBody, err := client.Put("/api/v1/network/assets/"+assetID, body)
			if err != nil {
				return err
			}

			id, name := resultAssetIDName(respBody)
			fmt.Fprintf(f.IO.ErrOut, "Asset %q (%s) updated.\n", name, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Asset name")
	cmd.Flags().StringVar(&category, "category", "", "Asset category ("+assetCategories+")")
	cmd.Flags().StringVar(&status, "status", "", "Asset status ("+assetStatuses+")")
	cmd.Flags().StringVar(&number, "number", "", "Asset number")
	cmd.Flags().StringVar(&warrantyExpiration, "warranty-expiration", "", "Warranty expiration date (YYYY-MM-DD, empty to clear)")
	cmd.Flags().StringVar(&notes, "notes", "", "Additional notes")

	return cmd
}
