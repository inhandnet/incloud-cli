package pos

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdRulesSet(f *factory.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "set <device-id>",
		Short: "Replace a device's POS custom rules",
		Long: "Replace the full set of POS custom rules for a device from a JSON file.\n\n" +
			"The file may contain either a bare array of rule entries or an object with a " +
			"\"rules\" array. Each entry has: type (add|mask), clientType (UPPERCASE), " +
			"vendor, protocol, address, port. Max 100 entries; this replaces all existing rules.",
		Args: cobra.ExactArgs(1),
		Example: `  # Set rules from a file
  incloud pos rules set 507f1f77bcf86cd799439011 --file rules.json

  # Read from stdin
  cat rules.json | incloud pos rules set 507f1f77bcf86cd799439011 --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}

			var data []byte
			var err error
			if file == "-" {
				data, err = io.ReadAll(f.IO.In)
			} else {
				data, err = os.ReadFile(file)
			}
			if err != nil {
				return fmt.Errorf("reading rules: %w", err)
			}

			body, err := buildRulesBody(data)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Post("/api/v1/network/devices/"+args[0]+"/pos/custom-rules", body)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				outcome := "ok"
				var env struct {
					PushOutcome string `json:"pushOutcome"`
					PushError   string `json:"pushError"`
				}
				if json.Unmarshal(resp, &env) == nil && env.PushOutcome != "" {
					outcome = env.PushOutcome
				}
				fmt.Fprintf(f.IO.ErrOut, "POS custom rules updated for device %s (push: %s).\n", args[0], outcome)
				if env.PushError != "" {
					fmt.Fprintf(f.IO.ErrOut, "Push error: %s\n", env.PushError)
				}
				return nil
			}
			return iostreams.FormatOutput(resp, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file with rules (use '-' for stdin)")

	return cmd
}

// buildRulesBody normalizes the input JSON into the {"rules": [...]} request
// envelope. It accepts either a bare array of rule entries or an object that
// already contains a "rules" array.
func buildRulesBody(data []byte) (map[string]json.RawMessage, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		return map[string]json.RawMessage{"rules": mustMarshal(arr)}, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if _, ok := obj["rules"]; !ok {
		return nil, fmt.Errorf("JSON object must contain a \"rules\" array")
	}
	return obj, nil
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
