package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClientUpdate(f *factory.Factory) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <client-id>",
		Short: "Update client name",
		Long:  "Update the display name of a connected client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := args[0]

			apiClient, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]string{"name": name}
			respBody, err := apiClient.Put("/api/v1/network/clients/"+clientID, body)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					ID   string `json:"_id"`
					Name string `json:"name"`
				} `json:"result"`
			}
			if err := json.Unmarshal(respBody, &resp); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Client %q (%s) updated.\n", resp.Result.Name, resp.Result.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New client name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
