package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type createRequest struct {
	App         string   `json:"app"`
	Type        string   `json:"type"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments,omitempty"`
}

func NewCmdFeedbackCreate(f *factory.Factory) *cobra.Command {
	var (
		content     string
		feedbackTyp string
		files       []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a feedback entry",
		Long: `Submit feedback about the InCloud platform or CLI.

The --content flag accepts a string or a file reference with @ prefix (e.g. @feedback.md).`,
		Example: `  # Submit feedback with inline content
  incloud feedback create --content "Signal rating is inaccurate"

  # Submit feedback from a file
  incloud feedback create --content @feedback.md

  # Submit with attachments
  incloud feedback create --content "See screenshot" --file screenshot.png --file log.txt

  # Submit a suggestion
  incloud feedback create --type suggestion --content "Add batch export feature"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Resolve content: @file or literal string.
			body := content
			if strings.HasPrefix(content, "@") {
				path := content[1:]
				data, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("reading content file %s: %w", path, err)
				}
				body = string(data)
			}

			if body == "" {
				return fmt.Errorf("--content is required and must not be empty")
			}

			// Upload attachments.
			var objectNames []string
			for _, filePath := range files {
				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("opening attachment %s: %w", filePath, err)
				}

				fmt.Fprintf(f.IO.ErrOut, "Uploading %s...\n", filepath.Base(filePath))
				resp, err := client.UploadPut(
					"/api/v1/feedbacks/attachments/upload",
					"file",
					filepath.Base(filePath),
					file,
				)
				_ = file.Close()
				if err != nil {
					return fmt.Errorf("uploading attachment %s: %w", filePath, err)
				}

				var result struct {
					Result struct {
						ObjectName string `json:"objectName"`
					} `json:"result"`
				}
				if err := json.Unmarshal(resp, &result); err != nil {
					return fmt.Errorf("parsing upload response: %w", err)
				}
				objectNames = append(objectNames, result.Result.ObjectName)
			}

			// Create feedback.
			req := createRequest{
				App:         "star",
				Type:        strings.ToUpper(feedbackTyp),
				Content:     body,
				Attachments: objectNames,
			}

			resp, err := client.Post("/api/v1/feedbacks", req)
			if err != nil {
				return err
			}

			var result struct {
				Result struct {
					ID string `json:"_id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "Feedback created. (id: %s)\n", result.Result.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&content, "content", "c", "", "Feedback content (use @file to read from file)")
	cmd.Flags().StringVarP(&feedbackTyp, "type", "t", "issue", "Feedback type: issue, question, comment, suggestion")
	cmd.Flags().StringArrayVar(&files, "file", nil, "Attachment file paths (repeatable)")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}
