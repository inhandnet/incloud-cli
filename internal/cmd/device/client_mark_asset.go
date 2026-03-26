package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClientMarkAsset(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-asset <client-id> [client-id...]",
		Short: "Mark clients as assets",
		Long:  "Convert one or more connected clients into tracked network assets.",
		Args:  cobra.MinimumNArgs(1),
		Example: `  # Mark a single client as asset
  incloud device client mark-asset 507f1f77bcf86cd799439011

  # Mark multiple clients
  incloud device client mark-asset id1 id2 id3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"ids": args,
			}
			_, err = client.Put("/api/v1/network/clients/mark-assets", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "%d client(s) marked as assets.\n", len(args))
			return nil
		},
	}

	return cmd
}
