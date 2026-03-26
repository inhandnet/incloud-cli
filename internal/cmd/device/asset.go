package device

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// Asset category and status values used in flag descriptions.
const (
	assetCategories = "router,gateway,ap,cash_register,barcode_scanner,voip_phone,printer,camera,mobile_phone,pc,pad,others"
	assetStatuses   = "in_stock,in_use,in_repair,decommissioned"
)

// resultAssetIDName extracts _id and name from {"result": {...}} response.
func resultAssetIDName(body []byte) (id, name string) {
	var resp struct {
		Result struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	_ = json.Unmarshal(body, &resp)
	return resp.Result.ID, resp.Result.Name
}

func NewCmdAsset(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Manage network assets",
		Long:  "List, create, update, and delete network assets tracked in your organization.",
	}

	cmd.AddCommand(newCmdAssetList(f))
	cmd.AddCommand(newCmdAssetCreate(f))
	cmd.AddCommand(newCmdAssetUpdate(f))
	cmd.AddCommand(newCmdAssetDelete(f))

	return cmd
}
