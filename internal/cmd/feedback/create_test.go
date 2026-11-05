package feedback

import (
	"encoding/json"
	"strings"
	"testing"
)

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestBuildCreateRequest_HumanUnchanged(t *testing.T) {
	req := buildCreateRequest("", "issue", "slow", "", nil, envOf(nil))
	if req.App != "star" || req.Type != "ISSUE" || req.Source != "" || req.AIContext != nil {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestBuildCreateRequest_HumanJSONOmitsAgentFields(t *testing.T) {
	data, err := json.Marshal(buildCreateRequest("", "issue", "slow", "", nil, envOf(map[string]string{envAgentSessionID: "  "})))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"source", "aiDiagnosis", "aiContext"} {
		if strings.Contains(string(data), `"`+key+`"`) {
			t.Fatalf("unexpected %s in %s", key, data)
		}
	}
}

func TestBuildCreateRequest_AgentJSON(t *testing.T) {
	data, _ := json.Marshal(buildCreateRequest("", "issue", "slow", "", nil, envOf(map[string]string{envAgentSessionID: "s1"})))
	want := `{"app":"star","type":"ISSUE","content":"slow","source":"AI_AGENT","aiContext":{"sessionId":"s1"}}`
	if string(data) != want {
		t.Fatalf("got %s", data)
	}
}

func TestBuildCreateRequest_Agent(t *testing.T) {
	req := buildCreateRequest("", "comment", "slow", "index missing", []string{"a.png"}, envOf(map[string]string{
		envAgentSessionID: "s1",
		envAgentTraceID:   "t1",
		envAgentPlatform:  "DeviceLive",
	}))
	if req.App != "devicelive" || req.Source != "AI_AGENT" || req.AIDiagnosis != "index missing" || req.Content != "slow" {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.AIContext == nil || *req.AIContext != (aiContext{SessionID: "s1", TraceID: "t1", Platform: "devicelive"}) {
		t.Fatalf("unexpected aiContext: %+v", req.AIContext)
	}
}

func TestBuildCreateRequest_ExplicitAppWins(t *testing.T) {
	req := buildCreateRequest("star", "issue", "x", "", nil, envOf(map[string]string{
		envAgentSessionID: "s1",
		envAgentPlatform:  "devicelive",
	}))
	if req.App != "star" {
		t.Fatalf("app = %s", req.App)
	}
}
