package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdConfigCopy(f *factory.Factory) *cobra.Command {
	var (
		module      string
		source      string
		sourceGroup string
		target      []string
		targetGroup []string
		yes         bool
	)

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy configuration to other devices or groups",
		Long:  "Copy a device or group configuration to one or more target devices or groups.",
		Example: `  # Copy from device to device
  incloud device config copy --source DEV1 --target DEV2

  # Copy from device to multiple targets
  incloud device config copy --source DEV1 --target DEV2 --target DEV3

  # Copy from device to groups
  incloud device config copy --source DEV1 --target-group GRP1 --target-group GRP2

  # Copy from group to groups
  incloud device config copy --source-group GRP1 --target-group GRP2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate: exactly one source
			if source == "" && sourceGroup == "" {
				return fmt.Errorf("either --source or --source-group is required")
			}
			if source != "" && sourceGroup != "" {
				return fmt.Errorf("--source and --source-group are mutually exclusive")
			}

			// Validate: at least one target
			if len(target) == 0 && len(targetGroup) == 0 {
				return fmt.Errorf("at least one --target or --target-group is required")
			}

			if !yes {
				sourceDesc := source
				if sourceDesc == "" {
					sourceDesc = "group " + sourceGroup
				}
				var parts []string
				if len(target) > 0 {
					parts = append(parts, fmt.Sprintf("%d device(s)", len(target)))
				}
				if len(targetGroup) > 0 {
					parts = append(parts, fmt.Sprintf("%d group(s)", len(targetGroup)))
				}
				msg := fmt.Sprintf("Copy config from %s to %s?", sourceDesc, strings.Join(parts, " and "))
				confirmed, err := ui.Confirm(f, msg)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			body := map[string]any{}
			if source != "" {
				body["sourceDeviceId"] = source
			}
			if sourceGroup != "" {
				body["sourceGroupId"] = sourceGroup
			}
			if len(target) > 0 {
				body["targetDeviceIds"] = target
			}
			if len(targetGroup) > 0 {
				body["targetGroupIds"] = targetGroup
			}
			if module != "" {
				body["module"] = module
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Post("/api/v1/config/layer/bulk-copy", body)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				fmt.Fprintln(f.IO.ErrOut, "Configuration copied successfully.")
				return nil
			}
			return iostreams.FormatOutput(resp, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Source device ID")
	cmd.Flags().StringVar(&sourceGroup, "source-group", "", "Source group ID (mutually exclusive with --source)")
	cmd.Flags().StringArrayVar(&target, "target", nil, "Target device ID (can be repeated)")
	cmd.Flags().StringArrayVar(&targetGroup, "target-group", nil, "Target group ID (can be repeated)")

	// Hidden aliases for backward compatibility
	cmd.Flags().StringArrayVar(&target, "to", nil, "Target device ID (can be repeated)")
	cmd.Flags().StringArrayVar(&targetGroup, "to-group", nil, "Target group ID (can be repeated)")
	_ = cmd.Flags().MarkHidden("to")
	_ = cmd.Flags().MarkHidden("to-group")
	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
