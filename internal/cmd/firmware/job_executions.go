package firmware

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type JobExecutionsOptions struct {
	cmdutil.ListFlags
	Firmware     string
	Device       string
	Module       string
	JobID        string
	Status       string
	SerialNumber string
}

func NewCmdJobExecutions(f *factory.Factory) *cobra.Command {
	opts := &JobExecutionsOptions{}

	cmd := &cobra.Command{
		Use:   "executions",
		Short: "List OTA job executions",
		Long: `List OTA firmware upgrade job executions with optional filtering.

Execution statuses: QUEUED, INPROGRESS, SUCCEEDED, FAILED, CANCELED

Use --firmware to list executions for a specific firmware, or --device to
list completed executions for a specific device.`,
		Aliases: []string{"exec"},
		Example: `  # List all OTA executions
  incloud firmware job executions

  # Filter by status
  incloud firmware job executions --status SUCCEEDED

  # Filter by job ID
  incloud firmware job executions --job 20260318000001

  # Filter by device serial number
  incloud firmware job executions --sn MR805123

  # List executions for a specific firmware
  incloud firmware job executions --firmware 69afb47b2ad10a3f4b20c02f

  # List completed executions for a specific device
  incloud firmware job executions --device 69b24d278760dc6390e28db1

  # Expand related resources
  incloud firmware job executions --expand job

  # Select fields
  incloud firmware job executions -f _id -f status -f device.name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Route to the appropriate endpoint
			if opts.Firmware != "" {
				return runFirmwareExecutions(cmd, f, opts)
			}
			if opts.Device != "" {
				return runDeviceExecutions(cmd, f, opts)
			}
			return runOTAExecutions(cmd, f, opts)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware ID (uses /firmwares/{id}/job/executions)")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID (uses /devices/{id}/ota/jobs/completed)")
	cmd.Flags().StringVar(&opts.Module, "module", "", "Filter by OTA module name (default: \"default\")")
	cmd.Flags().StringVar(&opts.JobID, "job", "", "Filter by job ID")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (QUEUED|INPROGRESS|SUCCEEDED|FAILED|CANCELED)")
	cmd.Flags().StringVar(&opts.SerialNumber, "sn", "", "Filter by device serial number (supports regex)")
	opts.ListFlags.RegisterExpand(cmd, "job")

	cmd.AddCommand(NewCmdExecCancel(f))
	cmd.AddCommand(NewCmdExecRetry(f))

	return cmd
}

func commonQuery(cmd *cobra.Command, opts *JobExecutionsOptions) url.Values {
	q := cmdutil.NewQuery(cmd, nil)
	if opts.JobID != "" {
		q.Set("jobId", opts.JobID)
	}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.SerialNumber != "" {
		q.Set("serialNumber", opts.SerialNumber)
	}
	return q
}

// runOTAExecutions lists all OTA job executions via GET /api/v1/ota/job/executions.
func runOTAExecutions(cmd *cobra.Command, f *factory.Factory, opts *JobExecutionsOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := commonQuery(cmd, opts)
	if opts.Module != "" {
		q.Set("module", opts.Module)
	} else {
		q.Set("module", "default")
	}

	output, _ := cmd.Flags().GetString("output")
	body, err := client.Get("/api/v1/ota/job/executions", q)
	if err != nil {
		return err
	}

	return iostreams.FormatOutput(body, f.IO, output)
}

// runFirmwareExecutions lists executions for a specific firmware via
// GET /api/v1/firmwares/{id}/job/executions.
func runFirmwareExecutions(cmd *cobra.Command, f *factory.Factory, opts *JobExecutionsOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := commonQuery(cmd, opts)
	if opts.Module != "" {
		q.Set("module", opts.Module)
	}

	output, _ := cmd.Flags().GetString("output")
	body, err := client.Get("/api/v1/firmwares/"+url.PathEscape(opts.Firmware)+"/job/executions", q)
	if err != nil {
		return err
	}

	return iostreams.FormatOutput(body, f.IO, output)
}

// runDeviceExecutions lists completed executions for a specific device via
// GET /api/v1/devices/{id}/ota/jobs/completed.
func runDeviceExecutions(cmd *cobra.Command, f *factory.Factory, opts *JobExecutionsOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := commonQuery(cmd, opts)
	if opts.Module != "" {
		q.Set("module", opts.Module)
	}

	output, _ := cmd.Flags().GetString("output")
	body, err := client.Get("/api/v1/devices/"+url.PathEscape(opts.Device)+"/ota/jobs/completed", q)
	if err != nil {
		return err
	}

	return iostreams.FormatOutput(body, f.IO, output)
}
