package feedback

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newTestFactory(t *testing.T, host string) (*factory.Factory, *bytes.Buffer) {
	t.Helper()

	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {Host: host, Token: "test-token"},
		},
	}
	if _, err := config.Update(cfgPath, func(c *config.Config) error { *c = *cfg; return nil }); err != nil {
		t.Fatal(err)
	}

	errBuf := &bytes.Buffer{}
	f := &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: errBuf,
		},
		ConfigPath: cfgPath,
	}
	return f, errBuf
}

func newFeedbackRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdFeedback(f))
	return root
}

func TestFeedbackWithdraw(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{"_id": "abc123", "resolution": "IGNORED"},
		})
	}))
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newFeedbackRoot(f)
	root.SetArgs([]string{"feedback", "withdraw", "abc123"})
	if err := root.Execute(); err != nil {
		t.Fatalf("feedback withdraw: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/api/v1/feedbacks/abc123/withdraw" {
		t.Errorf("request = %s %s, want POST /api/v1/feedbacks/abc123/withdraw", gotMethod, gotPath)
	}
	if !strings.Contains(errBuf.String(), "Feedback abc123 withdrawn.") {
		t.Errorf("stderr = %q, want withdrawn message", errBuf.String())
	}
}

func TestFeedbackWithdraw_AlreadyHandled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"error":   "feedback abc123 has already been handled and cannot be withdrawn",
			"status":  409,
			"code":    "invalid_state",
			"message": "feedback abc123 has already been handled and cannot be withdrawn",
		})
	}))
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newFeedbackRoot(f)
	root.SetArgs([]string{"feedback", "withdraw", "abc123"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected an error for a handled feedback")
	}
	if strings.Contains(errBuf.String(), "withdrawn.") {
		t.Errorf("stderr = %q, must not report success", errBuf.String())
	}
}

func TestFeedbackWithdraw_RequiresID(t *testing.T) {
	f, _ := newTestFactory(t, "http://127.0.0.1:1")
	root := newFeedbackRoot(f)
	root.SetArgs([]string{"feedback", "withdraw"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected an error without a feedback ID")
	}
}
