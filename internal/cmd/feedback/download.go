package feedback

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdFeedbackDownload(f *factory.Factory) *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "download <feedback-id>",
		Short: "Download attachments from a feedback entry",
		Long: `Download all attachments from a feedback entry to the current directory (or --dir).

Use 'incloud feedback list' to find feedback IDs and see which entries have attachments.

Note: this command searches the latest 100 feedback entries to locate the given ID.
Very old entries may not be found; use 'incloud feedback list --page N' to verify.`,
		Example: `  # Download attachments from a feedback entry
  incloud feedback download 69c3e7bb828ddd389e530a57

  # Download to a specific directory
  incloud feedback download 69c3e7bb828ddd389e530a57 --dir /tmp/attachments`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			feedbackID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// List feedbacks and find the matching entry by ID.
			// The API has no single-resource GET endpoint, so we search the list.
			q := url.Values{}
			q.Set("page", "0")
			q.Set("limit", "100")
			body, err := client.Get("/api/v1/feedbacks", q)
			if err != nil {
				return fmt.Errorf("listing feedbacks: %w", err)
			}

			var resp struct {
				Result []struct {
					ID          string   `json:"_id"`
					Attachments []string `json:"attachments"`
				} `json:"result"`
			}
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing feedback list: %w", err)
			}

			var attachments []string
			found := false
			for _, fb := range resp.Result {
				if fb.ID == feedbackID {
					found = true
					for _, a := range fb.Attachments {
						if a != "" {
							attachments = append(attachments, a)
						}
					}
					break
				}
			}
			if !found {
				return fmt.Errorf("feedback %s not found (searched latest 100 entries)", feedbackID)
			}

			if len(attachments) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "No attachments found for this feedback entry.")
				return nil
			}

			// Ensure output directory exists.
			if outputDir != "" {
				if err := os.MkdirAll(outputDir, 0o750); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}
			}

			for _, objectName := range attachments {
				// Extract filename from objectName (e.g. "2026-03-25/abc123/file.png" -> "file.png")
				filename := filepath.Base(objectName)
				destPath := filename
				if outputDir != "" {
					destPath = filepath.Join(outputDir, filename)
				}

				// Step 1: GET presigned URL from API.
				// objectName contains slashes (e.g. "2026-03-25/abc123/file.png") which map to
				// the backend's {date}/{random}/{filename} path parameters — do NOT escape them.
				apiPath := fmt.Sprintf("/api/v1/feedbacks/%s/attachments/%s",
					url.PathEscape(feedbackID), objectName)
				urlBody, err := client.Get(apiPath, nil)
				if err != nil {
					return fmt.Errorf("getting download URL for %s: %w", filename, err)
				}

				var urlResp struct {
					Result string `json:"result"`
				}
				if err := json.Unmarshal(urlBody, &urlResp); err != nil {
					return fmt.Errorf("parsing download URL for %s: %w", filename, err)
				}
				if urlResp.Result == "" {
					return fmt.Errorf("empty download URL for %s", filename)
				}

				// Step 2: Download from S3 presigned URL.
				fmt.Fprintf(f.IO.ErrOut, "Downloading %s...\n", filename)
				if err := client.Download(urlResp.Result, destPath); err != nil {
					return fmt.Errorf("downloading %s: %w", filename, err)
				}
				fmt.Fprintf(f.IO.ErrOut, "  Saved to %s\n", destPath)
			}

			fmt.Fprintf(f.IO.ErrOut, "%d attachment(s) downloaded.\n", len(attachments))
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "dir", "", "Output directory for downloaded files (default: current directory)")

	return cmd
}
