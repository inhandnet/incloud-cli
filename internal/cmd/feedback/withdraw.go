package feedback

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdFeedbackWithdraw(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw <feedback-id>",
		Short: "Withdraw a feedback entry you submitted",
		Long: `Withdraw a feedback entry you submitted.

Only the submitter can withdraw, and only while the entry is still new (not yet
handled). A withdrawn entry is marked ignored and carries a withdrawnAt time.`,
		Example: `  # Withdraw your own feedback
  incloud feedback withdraw 69c3e7bb828ddd389e530a57`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Post(fmt.Sprintf("/api/v1/feedbacks/%s/withdraw", url.PathEscape(args[0])), nil)
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
			if result.Result.ID == "" {
				return fmt.Errorf("unexpected response: missing feedback ID")
			}

			fmt.Fprintf(f.IO.ErrOut, "Feedback %s withdrawn.\n", result.Result.ID)
			return nil
		},
	}

	return cmd
}
