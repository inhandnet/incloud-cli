package alert

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdRuleTypes(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types [type]",
		Short: "List alert rule types and their parameters",
		Long: `List all supported alert rule types and their parameters.

Without arguments, displays a summary table of all types.
With a type name argument, displays detailed parameter information for that type.`,
		Example: `  # List all alert rule types
  incloud alert rule types

  # Show details for a specific type
  incloud alert rule types disconnected

  # JSON output of all types
  incloud alert rule types -o json

  # JSON output of a specific type
  incloud alert rule types disconnected -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")

			if len(args) == 1 {
				return runRuleTypeDetail(f, args[0], output)
			}
			return runRuleTypeList(f, output)
		},
	}

	return cmd
}

func runRuleTypeList(f *factory.Factory, output string) error {
	allTypes := AllRuleTypes()

	if output == "json" {
		data, err := json.MarshalIndent(allTypes, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(f.IO.Out, string(data))
		return err
	}

	tp := iostreams.NewTablePrinter(f.IO.Out, f.IO.IsStdoutTTY())
	tp.AddRow("TYPE", "CATEGORY", "PARAMS")
	for _, t := range allTypes {
		tp.AddRow(t.Type, t.Category, formatParamsSummary(t.Params))
	}
	return tp.Render()
}

func runRuleTypeDetail(f *factory.Factory, typeName, output string) error {
	def, ok := LookupRuleType(typeName)
	if !ok {
		return fmt.Errorf("unknown alert rule type %q (use 'incloud alert rule types' to list all types)", typeName)
	}

	if output == "json" {
		data, err := json.MarshalIndent(def, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(f.IO.Out, string(data))
		return err
	}

	// Print type header info
	fmt.Fprintf(f.IO.Out, "Type: %s\n", def.Type)
	fmt.Fprintf(f.IO.Out, "Category: %s\n", def.Category)
	fmt.Fprintf(f.IO.Out, "Description: %s\n", def.Description)

	if len(def.Params) == 0 {
		fmt.Fprintln(f.IO.Out, "\nThis type has no parameters.")
		return nil
	}

	fmt.Fprintln(f.IO.Out)
	tp := iostreams.NewTablePrinter(f.IO.Out, f.IO.IsStdoutTTY())
	tp.AddRow("PARAM", "TYPE", "UNIT", "DESCRIPTION")
	for _, p := range def.Params {
		unit := p.Unit
		if unit == "" {
			unit = "-"
		}
		tp.AddRow(p.Name, p.Type, unit, p.Description)
	}
	return tp.Render()
}

// formatParamsSummary returns a compact summary of params for the list view.
// e.g. "retention (seconds), threshold (percent)" or "-" if no params.
func formatParamsSummary(params []RuleParam) string {
	if len(params) == 0 {
		return "-"
	}
	parts := make([]string, len(params))
	for i, p := range params {
		if p.Unit != "" {
			parts[i] = fmt.Sprintf("%s (%s)", p.Name, p.Unit)
		} else {
			parts[i] = p.Name
		}
	}
	return strings.Join(parts, ", ")
}
