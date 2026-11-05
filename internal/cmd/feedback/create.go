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
	App         string     `json:"app"`
	Type        string     `json:"type"`
	Content     string     `json:"content"`
	Attachments []string   `json:"attachments,omitempty"`
	Source      string     `json:"source,omitempty"`
	AIDiagnosis string     `json:"aiDiagnosis,omitempty"`
	AIContext   *aiContext `json:"aiContext,omitempty"`
}

type aiContext struct {
	SessionID string `json:"sessionId"`
	TraceID   string `json:"traceId,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

// Set by the InCloud Copilot sandbox. Their presence means the feedback is
// being submitted by the AI assistant on the user's behalf, so it is recorded
// as AI_AGENT with a pointer back to the conversation.
const (
	envAgentSessionID = "INCLOUD_AGENT_SESSION_ID"
	envAgentTraceID   = "INCLOUD_AGENT_TRACE_ID"
	envAgentPlatform  = "INCLOUD_AGENT_PLATFORM"
)

// buildCreateRequest fills in source/aiContext from the copilot environment.
// app: explicit --app wins, otherwise the copilot platform, otherwise star.
func buildCreateRequest(app, feedbackTyp, body, aiDiagnosis string, attachments []string, getenv func(string) string) createRequest {
	req := createRequest{
		App:         app,
		Type:        strings.ToUpper(feedbackTyp),
		Content:     body,
		Attachments: attachments,
	}
	platform := strings.ToLower(strings.TrimSpace(getenv(envAgentPlatform)))
	if req.App == "" {
		if platform == "devicelive" {
			req.App = "devicelive"
		} else {
			req.App = "star"
		}
	}
	if sessionID := strings.TrimSpace(getenv(envAgentSessionID)); sessionID != "" {
		req.Source = "AI_AGENT"
		req.AIDiagnosis = aiDiagnosis
		req.AIContext = &aiContext{
			SessionID: sessionID,
			TraceID:   strings.TrimSpace(getenv(envAgentTraceID)),
			Platform:  platform,
		}
	} else if aiDiagnosis != "" {
		// Outside the copilot there is no conversation to point at; keep the
		// diagnosis rather than dropping it silently.
		req.Content = body + "\n\n---\nAI diagnosis:\n" + aiDiagnosis
	}
	return req
}

func readContentArg(v string) (string, error) {
	if !strings.HasPrefix(v, "@") {
		return v, nil
	}
	path := v[1:]
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading content file %s: %w", path, err)
	}
	return string(data), nil
}

func NewCmdFeedbackCreate(f *factory.Factory) *cobra.Command {
	var (
		content     string
		aiDiagnosis string
		app         string
		feedbackTyp string
		files       []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a feedback entry",
		Long: `Submit feedback about the InCloud platform or CLI.

The --content flag accepts a string or a file reference with @ prefix (e.g. @feedback.md).

When submitting on a user's behalf, put only what the user said and provided
in --content, and your own findings/conclusions in --ai-diagnosis.`,
		Example: `  # Submit feedback with inline content
  incloud feedback create --content "Signal rating is inaccurate"

  # Submit feedback from a file
  incloud feedback create --content @feedback.md

  # Submit with attachments
  incloud feedback create --content "See screenshot" --file screenshot.png --file log.txt

  # Submit a suggestion
  incloud feedback create --type suggestion --content "Add batch export feature"

  # Keep the user's report and the AI's diagnosis apart
  incloud feedback create --content @report.md --ai-diagnosis @diagnosis.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Resolve content: @file or literal string.
			body, err := readContentArg(content)
			if err != nil {
				return err
			}
			if strings.TrimSpace(body) == "" {
				return fmt.Errorf("--content is required and must not be empty")
			}
			diagnosis, err := readContentArg(aiDiagnosis)
			if err != nil {
				return err
			}
			// Outside the copilot there is no conversation to attach the diagnosis to.
			if diagnosis != "" && strings.TrimSpace(os.Getenv(envAgentSessionID)) == "" {
				return fmt.Errorf("--ai-diagnosis is only supported when submitted by InCloud Copilot")
			}
			app = strings.ToLower(strings.TrimSpace(app))
			if app != "" && app != "star" && app != "devicelive" {
				return fmt.Errorf("invalid --app %q: must be star or devicelive", app)
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
			req := buildCreateRequest(app, feedbackTyp, body, diagnosis, objectNames, os.Getenv)

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
	cmd.Flags().StringVar(&aiDiagnosis, "ai-diagnosis", "", "AI assistant's own findings, kept apart from the user's report (use @file to read from file)")
	cmd.Flags().StringVar(&app, "app", "", "SaaS app the feedback belongs to: star, devicelive (default: current platform, else star)")
	cmd.Flags().StringVarP(&feedbackTyp, "type", "t", "issue", "Feedback type: issue, question, comment, suggestion")
	cmd.Flags().StringArrayVar(&files, "file", nil, "Attachment file paths (repeatable)")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}
