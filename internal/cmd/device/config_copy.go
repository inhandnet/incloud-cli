package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigCopy(f *factory.Factory) *cobra.Command {
	var (
		module      string
		source      string
		sourceGroup string
		to          []string
		toGroup     []string
	)

	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy configuration to other devices or groups",
		Long:  "Copy a device or group configuration to one or more target devices or groups.",
		Example: `  # Copy from device to device
  incloud device config copy --source DEV1 --to DEV2

  # Copy from device to multiple targets
  incloud device config copy --source DEV1 --to DEV2 --to DEV3

  # Copy from device to groups
  incloud device config copy --source DEV1 --to-group GRP1 --to-group GRP2

  # Copy from group to groups
  incloud device config copy --source-group GRP1 --to-group GRP2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate: exactly one source
			if source == "" && sourceGroup == "" {
				return fmt.Errorf("either --source or --source-group is required")
			}
			if source != "" && sourceGroup != "" {
				return fmt.Errorf("--source and --source-group are mutually exclusive")
			}

			// Validate: at least one target
			if len(to) == 0 && len(toGroup) == 0 {
				return fmt.Errorf("at least one --to or --to-group is required")
			}

			body := map[string]interface{}{}
			if source != "" {
				body["sourceDeviceId"] = source
			}
			if sourceGroup != "" {
				body["sourceGroupId"] = sourceGroup
			}
			if len(to) > 0 {
				body["targetDeviceIds"] = to
			}
			if len(toGroup) > 0 {
				body["targetGroupIds"] = toGroup
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
	cmd.Flags().StringArrayVar(&to, "to", nil, "Target device ID (can be repeated)")
	cmd.Flags().StringArrayVar(&toGroup, "to-group", nil, "Target group ID (can be repeated)")
	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")

	return cmd
}
