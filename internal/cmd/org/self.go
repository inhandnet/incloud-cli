package org

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdSelf(f *factory.Factory) *cobra.Command {
	var fields []string

	cmd := &cobra.Command{
		Use:   "self",
		Short: "Show current organization",
		Example: `  # Show current organization
  incloud org self

  # Table output
  incloud org self -o table

  # Only specific fields
  incloud org self -f name -f email -f userCount -f deviceCount

  # YAML output
  incloud org self -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/orgs/self", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fs := fields
			if len(fs) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fs = defaultListFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fs)
		},
	}

	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
