package model

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdDeploy(f *factory.Factory) *cobra.Command {
	var (
		jobsJSON    string
		targetType  string
		scheduledAt string
		filter      bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy AI models to devices or groups",
		Long: `Create batch deployment jobs for AI models to edge devices or device groups.

The --jobs flag accepts a JSON array of job definitions, each with "targets" (device/group IDs) and "model" (model ID).
When using --scheduled-at, the time must be at least 30 minutes from now and no more than 31 days in the future.`,
		Example: `  # Deploy a model to devices
  incloud model deploy --jobs '[{"targets":["device-id-1"],"model":"model-id"}]' --target-type DEVICE

  # Deploy to a group with scheduling
  incloud model deploy --jobs '[{"targets":["group-id-1"],"model":"model-id"}]' --target-type GROUP --scheduled-at 2024-06-01T10:00:00Z

  # Deploy with license filter
  incloud model deploy --jobs '[{"targets":["device-id-1"],"model":"model-id"}]' --target-type DEVICE --filter`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var jobs []interface{}
			if err := json.Unmarshal([]byte(jobsJSON), &jobs); err != nil {
				return fmt.Errorf("invalid --jobs JSON: %w", err)
			}

			reqBody := map[string]interface{}{
				"jobs":       jobs,
				"targetType": targetType,
				"filter":     filter,
			}
			if scheduledAt != "" {
				reqBody["scheduledAt"] = scheduledAt
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/models/batch/jobs", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Model deployment job(s) created.\n")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&jobsJSON, "jobs", "", `Job definitions as JSON array (e.g. '[{"targets":["id"],"model":"model-id"}]') (required)`)
	cmd.Flags().StringVar(&targetType, "target-type", "", "Target type: DEVICE or GROUP (required)")
	cmd.Flags().StringVar(&scheduledAt, "scheduled-at", "", "Scheduled deployment time (ISO 8601, must be 30min-31days from now)")
	cmd.Flags().BoolVar(&filter, "filter", false, "Enable license filtering")
	_ = cmd.MarkFlagRequired("jobs")
	_ = cmd.MarkFlagRequired("target-type")

	return cmd
}
