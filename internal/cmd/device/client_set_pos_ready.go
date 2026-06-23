package device

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

var posReadyLevels = []string{"priority", "default", "bypass"}

func newCmdClientSetPosReady(f *factory.Factory) *cobra.Command {
	var level string

	cmd := &cobra.Command{
		Use:   "set-pos-ready <client-id>",
		Short: "Set POS priority level for a client",
		Long: "Set the POS traffic priority level for a client.\n\n" +
			"Levels:\n" +
			"  priority  prioritize this client's POS traffic\n" +
			"  default   no special handling (equivalent to unmarked)\n" +
			"  bypass    exclude this client from POS handling",
		Args: cobra.ExactArgs(1),
		Example: `  # Prioritize a client's POS traffic
  incloud device client set-pos-ready 69b8c537e7f8d2c1e5fffdbc --level priority

  # Reset to default
  incloud device client set-pos-ready 69b8c537e7f8d2c1e5fffdbc --level default

  # Bypass a client
  incloud dev client set-pos-ready 69b8c537e7f8d2c1e5fffdbc --level bypass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized := strings.ToLower(strings.TrimSpace(level))
			if !isValidPosReadyLevel(normalized) {
				return fmt.Errorf("invalid --level %q (expect one of: %s)", level, strings.Join(posReadyLevels, ", "))
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			clientID := args[0]
			body := map[string]any{"level": normalized}
			_, err = client.Post("/api/v1/network/clients/"+clientID+"/pos-ready", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "POS priority set to %s for client %s.\n", normalized, clientID)
			return nil
		},
	}

	cmd.Flags().StringVar(&level, "level", "", "POS priority level: priority, default, or bypass (required)")
	_ = cmd.MarkFlagRequired("level")

	return cmd
}

func isValidPosReadyLevel(level string) bool {
	return slices.Contains(posReadyLevels, level)
}
