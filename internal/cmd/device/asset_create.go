package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdAssetCreate(f *factory.Factory) *cobra.Command {
	var (
		name               string
		mac                string
		category           string
		status             string
		number             string
		warrantyExpiration string
		notes              string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a network asset",
		Long:  "Create a new network asset with the given properties.",
		Example: `  # Create a router asset
  incloud device asset create --name "Office Router" --mac "00:18:05:AB:CD:EF" --category router --status in_use

  # Create with all fields
  incloud device asset create --name "Printer" --mac "AA:BB:CC:DD:EE:FF" \
    --category printer --status in_stock --number "AST-001" \
    --warranty-expiration "2027-12-31" --notes "2nd floor"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"name":     name,
				"mac":      mac,
				"category": category,
				"status":   status,
			}
			if number != "" {
				body["number"] = number
			}
			if warrantyExpiration != "" {
				body["warrantyExpiration"] = warrantyExpiration
			}
			if notes != "" {
				body["notes"] = notes
			}

			respBody, err := client.Post("/api/v1/network/assets", body)
			if err != nil {
				return err
			}

			id, name := resultAssetIDName(respBody)
			fmt.Fprintf(f.IO.ErrOut, "Asset %q (%s) created.\n", name, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Asset name (required)")
	cmd.Flags().StringVar(&mac, "mac", "", "MAC address (required, e.g. 00:18:05:AB:CD:EF)")
	cmd.Flags().StringVar(&category, "category", "", "Asset category (required: "+assetCategories+")")
	cmd.Flags().StringVar(&status, "status", "", "Asset status (required: "+assetStatuses+")")
	cmd.Flags().StringVar(&number, "number", "", "Asset number")
	cmd.Flags().StringVar(&warrantyExpiration, "warranty-expiration", "", "Warranty expiration date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&notes, "notes", "", "Additional notes")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("mac")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("status")

	return cmd
}
