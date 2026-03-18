package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdGroupDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete a device group",
		Example: `  # Delete with confirmation prompt
  incloud device group delete 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud device group delete 507f1f77bcf86cd799439011 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Fetch group to show name in confirmation and detect invalid IDs early.
			body, err := client.Get("/api/v1/devicegroups/"+id, nil)
			if err != nil {
				return fmt.Errorf("device group %s not found", id)
			}
			var resp struct {
				Result struct {
					Name string `json:"name"`
				} `json:"result"`
				Error string `json:"error"`
			}
			_ = json.Unmarshal(body, &resp)
			if resp.Error != "" {
				return fmt.Errorf("device group %s not found", id)
			}
			name := resp.Result.Name

			if !yes {
				prompt := fmt.Sprintf("Delete device group %q (%s)?", name, id)
				confirmed, err := ui.Confirm(f, prompt)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Delete("/api/v1/devicegroups/" + id)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Device group %q (%s) deleted.\n", name, id)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
