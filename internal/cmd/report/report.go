package report

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

const (
	policiesPath = "/api/v1/incloud/report/policies"
	reportsPath  = "/api/v1/incloud/reports"
)

func NewCmdReport(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Manage reports and report policies",
		Long: `List, download and regenerate reports, and manage the policies that produce them.

A report policy either generates a report once (one-off) or on a schedule (daily,
weekly or monthly). Each run produces a report that can be listed and downloaded
with 'incloud report list' / 'incloud report download'.`,
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdDownload(f))
	cmd.AddCommand(NewCmdRecreate(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdPolicy(f))

	return cmd
}
