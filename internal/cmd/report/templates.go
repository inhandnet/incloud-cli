package report

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// TemplateDef describes a report template (one section of a report).
type TemplateDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	XLSXOnly    bool   `json:"xlsxOnly"`
}

// templateRegistry mirrors incloud-manager report/ReportTemplateName.java.
// XLSXOnly mirrors ReportTemplateName.RESTRICTED_TEMPLATES.
var templateRegistry = []TemplateDef{
	{Name: "TOTAL_DEVICES", Description: "Total number of devices"},
	{Name: "DEVICE_DISTRIBUTION", Description: "Device distribution by model"},
	{Name: "FIRMWARE_STATUS", Description: "Firmware version status"},
	{Name: "CONFIGURATION_STATUS", Description: "Configuration sync status"},
	{Name: "NETWORKING_METHOD", Description: "Uplink networking methods"},
	{Name: "USAGE_SUMMARY", Description: "Data usage summary"},
	{Name: "OFFLINE_TIMES_TOP_10", Description: "Top 10 devices by offline count"},
	{Name: "OFFLINE_TIMES", Description: "Offline count per device"},
	{Name: "TOTAL_USAGE_TOP_10", Description: "Top 10 devices by total data usage"},
	{Name: "USAGE_TOP_10_BY_CELLULAR", Description: "Top 10 devices by cellular data usage"},
	{Name: "USAGE_TOP_10_BY_WIRED", Description: "Top 10 devices by wired data usage"},
	{Name: "USAGE_TOP_10_BY_WIRELESS", Description: "Top 10 devices by wireless data usage"},
	{Name: "DEVICE_DATA_USAGE_DETAILS", Description: "Data usage details per device", XLSXOnly: true},
	{Name: "ALERTS_TOP_10_BY_DEVICE", Description: "Top 10 devices by alert count"},
	{Name: "ALERT_TOP_10_BY_TYPE", Description: "Top 10 alert types"},
	{Name: "DEVICE_ONLINE_RATE", Description: "Online rate per device", XLSXOnly: true},
	{Name: "CLIENT_DATA_USAGE_DETAILS", Description: "Data usage details per client", XLSXOnly: true},
}

func lookupTemplate(name string) (TemplateDef, bool) {
	for _, t := range templateRegistry {
		if t.Name == name {
			return t, true
		}
	}
	return TemplateDef{}, false
}

func NewCmdPolicyTemplates(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List report templates",
		Long: `List the templates (report sections) accepted by --template.

Templates marked xlsxOnly can only be used with --format xlsx.`,
		Example: `  # List templates
  incloud report policy templates -o table`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := json.Marshal(map[string]any{"result": templateRegistry})
			if err != nil {
				return err
			}
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, iostreams.WithColumns("name", "xlsxOnly", "description"))
		},
	}
	return cmd
}
