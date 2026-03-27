package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecCapture(f *factory.Factory) *cobra.Command {
	var (
		iface         string
		captureTime   int
		source        string
		expertOptions string
		download      string
	)

	cmd := &cobra.Command{
		Use:   "capture <device-id>",
		Short: "Start packet capture (tcpdump) on a device",
		Example: `  # Capture on a specific interface
  incloud device exec capture 507f1f77bcf86cd799439011 --interface eth0

  # With duration and source filter
  incloud device exec capture 507f1f77bcf86cd799439011 --interface eth0 --duration 60 --source 192.168.1.1

  # Capture and download the pcap file
  incloud device exec capture 507f1f77bcf86cd799439011 --interface eth0 --download capture.pcap`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCaptureAndWait(f, cmd, args[0], map[string]interface{}{
				"interface":     iface,
				"captureTime":   captureTime,
				"source":        source,
				"expertOptions": expertOptions,
			}, download)
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "", "Network interface (required)")
	cmd.Flags().IntVar(&captureTime, "duration", 0, "Capture duration in seconds")
	cmd.Flags().StringVar(&source, "source", "", "Source IP filter")
	cmd.Flags().StringVar(&expertOptions, "expert-options", "", "Advanced tcpdump options")
	cmd.Flags().StringVarP(&download, "download", "d", "", "Download pcap file to local path after capture")
	_ = cmd.MarkFlagRequired("interface")

	return cmd
}

// capturePollInterval controls how often we poll for capture status.
// Tests can override this to avoid slow waits.
var capturePollInterval = 2 * time.Second

// runCaptureAndWait starts a capture task, polls until completion, and outputs the result.
// On Ctrl+C, it cancels the running task.
func runCaptureAndWait(f *factory.Factory, cmd *cobra.Command, deviceID string, params map[string]interface{}, download string) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := cleanDiagnosisParams(params)
	respBody, err := client.Post("/api/v1/devices/"+deviceID+"/diagnosis/capture", body)
	if err != nil {
		return err
	}

	taskID := gjson.GetBytes(respBody, "result._id").String()
	status := gjson.GetBytes(respBody, "result.status").String()

	// Already terminal?
	if isCaptureTerminalStatus(status) {
		return captureFinish(f, cmd, client, respBody, download)
	}

	ctx, cancel := setupTaskCancellation(client, taskID)
	defer cancel()

	if f.IO.IsStdoutTTY() {
		fmt.Fprintf(f.IO.ErrOut, "Capturing... (Ctrl+C to cancel)\n")
	}

	// Poll until terminal status
	endpoint := "/api/v1/devices/" + deviceID + "/diagnosis/capture"
	ticker := time.NewTicker(capturePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("canceled")
		case <-ticker.C:
			respBody, err = client.Get(endpoint, nil)
			if err != nil {
				return err
			}
			status = gjson.GetBytes(respBody, "result.status").String()
			if isCaptureTerminalStatus(status) {
				return captureFinish(f, cmd, client, respBody, download)
			}
		}
	}
}

// captureFinish handles the final output after capture completes.
// If download is set and capture succeeded, downloads the pcap file.
// Otherwise prints a download hint.
func captureFinish(f *factory.Factory, cmd *cobra.Command, client *api.APIClient, respBody []byte, download string) error {
	status := strings.ToLower(gjson.GetBytes(respBody, "result.status").String())
	fileURL := gjson.GetBytes(respBody, "result.fileUrl").String()

	if status == "finished" && fileURL != "" {
		if download != "" {
			if err := client.Download(fileURL, download); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.ErrOut, "Capture saved to %s\n", download)
			return nil
		}
		fmt.Fprintf(f.IO.ErrOut, "Capture finished. Download:\n  incloud api %s --output-file capture.pcap\n", fileURL)
	}

	return formatOutput(cmd, f.IO, respBody)
}

func isCaptureTerminalStatus(status string) bool {
	switch strings.ToLower(status) {
	case "finished", "failed", "canceled":
		return true
	}
	return false
}
