package user

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// resultIDName extracts _id and name from a standard {"result": {...}} API response.
func resultIDName(body []byte) (id, name string) {
	var resp struct {
		Result struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	_ = json.Unmarshal(body, &resp)
	return resp.Result.ID, resp.Result.Name
}

func NewCmdUser(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  "List, create, update, delete, and manage users on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdMe(f))
	cmd.AddCommand(NewCmdIdentity(f))
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdUpdate(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdLock(f))
	cmd.AddCommand(NewCmdUnlock(f))

	return cmd
}
