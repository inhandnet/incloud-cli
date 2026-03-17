package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecCancel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel <diagnosis-id>",
		Short:   "Cancel a running diagnostic task",
		Example: `  incloud device exec cancel 507f1f77bcf86cd799439011`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			diagID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			_, err = client.Put("/api/v1/diagnosis/"+diagID+"/cancel", nil)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Diagnosis %s canceled.\n", diagID)
			return nil
		},
	}

	return cmd
}
