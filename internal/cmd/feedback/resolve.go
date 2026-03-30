package feedback

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type resolveRequest struct {
	Resolution string `json:"resolution"`
	Reply      string `json:"reply,omitempty"`
}

func NewCmdFeedbackResolve(f *factory.Factory) *cobra.Command {
	var (
		resolution string
		reply      string
	)

	cmd := &cobra.Command{
		Use:   "resolve <feedback-id>",
		Short: "Update the resolution status of a feedback entry",
		Long: `Update the resolution status of a feedback entry.

Valid resolution values: new, approved, resolved, ignored`,
		Example: `  # Mark as approved
  incloud feedback resolve 69c3e7bb828ddd389e530a57 --resolution approved

  # Mark as resolved with a reply
  incloud feedback resolve 69c3e7bb828ddd389e530a57 --resolution resolved --reply "Fixed in v2.1.0"

  # Mark as ignored
  incloud feedback resolve 69c3e7bb828ddd389e530a57 --resolution ignored --reply "Works as intended"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			valid := map[string]bool{"new": true, "approved": true, "resolved": true, "ignored": true}
			if !valid[strings.ToLower(resolution)] {
				return fmt.Errorf("invalid resolution %q: must be one of new, approved, resolved, ignored", resolution)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := resolveRequest{
				Resolution: strings.ToLower(resolution),
				Reply:      reply,
			}

			resp, err := client.Put(fmt.Sprintf("/api/v1/feedbacks/%s", url.PathEscape(id)), req)
			if err != nil {
				return err
			}

			var result struct {
				Result struct {
					ID         string `json:"_id"`
					Resolution string `json:"resolution"`
					Reply      string `json:"reply"`
				} `json:"result"`
			}
			if err := json.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			if result.Result.ID == "" {
				return fmt.Errorf("unexpected response: missing feedback ID")
			}

			fmt.Fprintf(f.IO.ErrOut, "Feedback %s updated. (resolution: %s)\n", result.Result.ID, result.Result.Resolution)
			return nil
		},
	}

	cmd.Flags().StringVarP(&resolution, "resolution", "r", "", "Resolution status: new, approved, resolved, ignored")
	cmd.Flags().StringVar(&reply, "reply", "", "Resolution reply or comment")
	_ = cmd.MarkFlagRequired("resolution")

	return cmd
}
