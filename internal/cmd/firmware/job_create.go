package firmware

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type JobCreateOptions struct {
	TargetType      string
	Targets         []string
	Versions        []string
	ScheduledAt     string
	Filter          bool
	UpgradableStart string
	UpgradableEnd   string
}

var defaultJobFields = []string{"_id", "type", "status", "createdAt", "ignored.unlicensed", "ignored.filtered"}

func (o *JobCreateOptions) toRequestBody(firmwareID string) map[string]any {
	job := map[string]any{
		"firmware": firmwareID,
		"targets":  o.Targets,
	}
	if len(o.Versions) > 0 {
		job["versions"] = o.Versions
	}

	body := map[string]any{
		"targetType": o.TargetType,
		"filter":     o.Filter,
		"jobs":       []any{job},
	}
	if o.ScheduledAt != "" {
		body["scheduledAt"] = o.ScheduledAt
	}
	if o.UpgradableStart != "" || o.UpgradableEnd != "" {
		period := map[string]any{}
		if o.UpgradableStart != "" {
			period["startTime"] = o.UpgradableStart
		}
		if o.UpgradableEnd != "" {
			period["endTime"] = o.UpgradableEnd
		}
		body["upgradablePeriod"] = period
	}
	return body
}

func NewCmdJobCreate(f *factory.Factory) *cobra.Command {
	opts := &JobCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create <firmwareId>",
		Short: "Create an OTA firmware upgrade job",
		Long: `Create an OTA firmware upgrade job for specified target devices or groups.

The job will push the specified firmware to target devices. Devices without
a valid license are automatically excluded (reported as "unlicensed" in ignored).

Target types:
  DEVICE  - target specific devices by ID
  GROUP   - target device groups by ID`,
		Example: `  # Upgrade a device to a specific firmware
  incloud firmware job create 6989afd5eeb72121455dc104 \
    --target-type DEVICE --target 6989ad34a7455f3f0bf9dce2

  # Upgrade multiple devices
  incloud firmware job create 6989afd5eeb72121455dc104 \
    --target-type DEVICE \
    --target 6989ad34a7455f3f0bf9dce2 \
    --target 691bde8c96946b3e64095380

  # Upgrade a device group with version filter
  incloud firmware job create 6989afd5eeb72121455dc104 \
    --target-type GROUP --target 507f1f77bcf86cd799439011 \
    --version V2.0.22

  # Schedule for later with smart filtering
  incloud firmware job create 6989afd5eeb72121455dc104 \
    --target-type DEVICE --target 6989ad34a7455f3f0bf9dce2 \
    --scheduled-at 2026-04-01T10:00:00Z --filter

  # Set upgradable time window
  incloud firmware job create 6989afd5eeb72121455dc104 \
    --target-type DEVICE --target 6989ad34a7455f3f0bf9dce2 \
    --upgradable-start 02:00 --upgradable-end 05:00`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			firmwareID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			respBody, err := client.Post("/api/v1/firmwares/batch/jobs", opts.toRequestBody(firmwareID))
			if err != nil {
				return err
			}

			// Parse response for stderr confirmation
			var resp struct {
				Result []struct {
					ID      string `json:"_id"`
					Ignored struct {
						Unlicensed int `json:"unlicensed"`
						Filtered   int `json:"filtered"`
					} `json:"ignored"`
				} `json:"result"`
			}
			if err := json.Unmarshal(respBody, &resp); err == nil && len(resp.Result) > 0 {
				r := resp.Result[0]
				if r.ID != "" {
					fmt.Fprintf(f.IO.ErrOut, "OTA job %q created for firmware %s. (targets excluded: unlicensed=%d, filtered=%d)\n",
						r.ID, firmwareID, r.Ignored.Unlicensed, r.Ignored.Filtered)
				} else {
					fmt.Fprintf(f.IO.ErrOut, "No job created: all target devices were excluded. (unlicensed=%d, filtered=%d)\n",
						r.Ignored.Unlicensed, r.Ignored.Filtered)
				}
			}

			output, _ := cmd.Flags().GetString("output")
			var fields []string
			if output == "table" {
				fields = defaultJobFields
			}
			return iostreams.FormatOutput(respBody, f.IO, output, fields)
		},
	}

	cmd.Flags().StringVar(&opts.TargetType, "target-type", "", "Target type: DEVICE or GROUP (required)")
	cmd.Flags().StringArrayVar(&opts.Targets, "target", nil, "Target device or group ID (required, repeatable)")
	cmd.Flags().StringArrayVar(&opts.Versions, "version", nil, "Source firmware version to match (repeatable)")
	cmd.Flags().StringVar(&opts.ScheduledAt, "scheduled-at", "", "Scheduled execution time (ISO 8601, must be 30min-31days from now)")
	cmd.Flags().BoolVar(&opts.Filter, "filter", false, "Enable smart filtering (skip devices already at target version)")
	cmd.Flags().StringVar(&opts.UpgradableStart, "upgradable-start", "", "Upgradable period start time (HH:mm)")
	cmd.Flags().StringVar(&opts.UpgradableEnd, "upgradable-end", "", "Upgradable period end time (HH:mm)")

	_ = cmd.MarkFlagRequired("target-type")
	_ = cmd.MarkFlagRequired("target")

	return cmd
}
