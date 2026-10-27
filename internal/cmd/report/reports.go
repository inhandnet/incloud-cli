package report

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

type listOptions struct {
	cmdutil.ListFlags
	Name string
	From string
	To   string
}

// applyTo is not shown but must be fetched: the API fails when it is projected away.
var defaultReportListFields = []string{"_id", "name", "applyTo", "format", "period", "timezone", "completed", "reason", "createdAt", "completedAt"}

var reportListColumns = []string{"_id", "name", "format", "period", "timezone", "completed", "reason", "createdAt", "completedAt"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List generated reports",
		Long:    "List generated reports with pagination. The platform keeps reports from the last 90 days.",
		Aliases: []string{"ls"},
		Example: `  # List reports
  incloud report list

  # Filter by name and creation time
  incloud report list --name weekly --from 2026-10-01 --to 2026-10-27`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultReportListFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.From != "" {
				q.Set("from", cmdutil.ParseTimeFlag(opts.From))
			}
			if opts.To != "" {
				q.Set("to", cmdutil.ParseTimeFlag(opts.To))
			}

			body, err := client.Get(reportsPath, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, iostreams.WithColumns(reportListColumns...))
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by report name (fuzzy match)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Created at or after (e.g. 2026-10-01 or RFC 3339)")
	cmd.Flags().StringVar(&opts.To, "to", "", "Created before (e.g. 2026-10-27 or RFC 3339)")

	return cmd
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <report-id>",
		Short: "Get report details",
		Example: `  # Get report details
  incloud report get 507f1f77bcf86cd799439011`,
		Args: cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "report id", "incloud report list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get(reportsPath+"/"+args[0], nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}
	return cmd
}

func NewCmdDownload(f *factory.Factory) *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "download <report-id>",
		Short: "Download a generated report",
		Long: `Download the file (PDF or XLSX) of a generated report.

With --output-file the report is saved to that path; without it the temporary
download URL is printed instead.`,
		Example: `  # Save a report to a file
  incloud report download 507f1f77bcf86cd799439011 --output-file weekly.pdf

  # Print the temporary download URL
  incloud report download 507f1f77bcf86cd799439011`,
		Args: cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "report id", "incloud report list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get(reportsPath+"/"+args[0]+"/download", nil)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					DownloadURL string `json:"downloadUrl"`
				} `json:"result"`
			}
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing download URL: %w", err)
			}
			if resp.Result.DownloadURL == "" {
				return fmt.Errorf("no download URL for report %s (the report may not be completed yet)", args[0])
			}

			if outputFile == "" {
				_, err := fmt.Fprintln(f.IO.Out, resp.Result.DownloadURL)
				return err
			}

			if err := client.Download(resp.Result.DownloadURL, outputFile); err != nil {
				return fmt.Errorf("downloading report: %w", err)
			}
			fmt.Fprintf(f.IO.ErrOut, "Saved to %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFile, "output-file", "", "Save the report to this file (default: print the download URL)")

	return cmd
}

func NewCmdRecreate(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recreate <report-id>",
		Short: "Regenerate a report",
		Long:  "Regenerate a report with the same settings and time range, e.g. after a failed run.",
		Example: `  # Regenerate a report
  incloud report recreate 507f1f77bcf86cd799439011`,
		Args: cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "report id", "incloud report list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if _, err := client.Post(reportsPath+"/"+args[0]+"/recreate", nil); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.ErrOut, "Report %s is being regenerated.\n", args[0])
			return nil
		},
	}
	return cmd
}

func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <report-id> [<report-id>...]",
		Short:   "Delete generated reports",
		Aliases: []string{"rm"},
		Example: `  # Delete a report (will prompt for confirmation)
  incloud report delete 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud report delete 507f1f77bcf86cd799439011 --yes`,
		Args: cmdutil.ObjectIDArgsAll(cobra.MinimumNArgs(1), "report id", "incloud report list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteByIDs(f, args, yes, reportsPath, "report")
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// deleteByIDs deletes resources one by one (the API has no bulk delete).
func deleteByIDs(f *factory.Factory, ids []string, yes bool, basePath, noun string) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	if !yes {
		msg := fmt.Sprintf("Delete %s %s?", noun, ids[0])
		if len(ids) > 1 {
			msg = fmt.Sprintf("Delete %d %ss?", len(ids), noun)
		}
		confirmed, err := ui.Confirm(f, msg)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	for _, id := range ids {
		if _, err := client.Delete(basePath + "/" + id); err != nil {
			return fmt.Errorf("deleting %s %s: %w", noun, id, err)
		}
		fmt.Fprintf(f.IO.ErrOut, "Deleted %s %s.\n", noun, id)
	}
	return nil
}
