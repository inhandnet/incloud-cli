package device

import (
	"context"
	"fmt"
	"io"
	"net/http"

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

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			reqURL := actx.Host + "/api/v1/diagnosis/" + diagID + "/cancel"
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqURL, http.NoBody)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				fmt.Fprintln(f.IO.ErrOut, string(respBody))
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}

			fmt.Fprintf(f.IO.ErrOut, "Diagnosis %s canceled.\n", diagID)
			return nil
		},
	}

	return cmd
}
