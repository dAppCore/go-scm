// SPDX-Licence-Identifier: EUPL-1.2

package handlers

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFakeSSHCommand(t *testing.T, outputPath string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "ssh")
	scriptContent := "#!/bin/sh\n" +
		"OUT=" + strconv.Quote(outputPath) + "\n" +
		"printf '%s\n' \"$@\" >> \"$OUT\"\n" +
		"cat >> \"${OUT}.stdin\"\n"
	require.NoError(t, os.WriteFile(script, []byte(scriptContent), 0o755))
	return dir
}

// newTestSpinner creates a Spinner with the given agents for testing.
func newTestSpinner(agents map[string]agentci.AgentConfig) *agentci.Spinner {
	return agentci.NewSpinner(agentci.ClothoConfig{Strategy: "direct"}, agents)
}

// --- Match tests ---

func TestDispatch_Match_Good_NeedsCoding(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
	}
	assert.True(t, h.Match(sig))
}

func TestDispatch_Match_Good_MultipleAgents(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
		"local-codex":  {Host: "localhost", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "local-codex",
	}
	assert.True(t, h.Match(sig))
}

func TestDispatch_Match_Bad_HasPR(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: false,
		PRNumber:    7,
		Assignee:    "darbs-claude",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_UnknownAgent(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "unknown-user",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_NotAssigned(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_EmptyAgentMap(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{})
	h := NewDispatchHandler(nil, "", "", spinner)
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
	}
	assert.False(t, h.Match(sig))
}

// --- Name test ---

func TestDispatch_Name_Good(t *testing.T) {
	spinner := newTestSpinner(nil)
	h := NewDispatchHandler(nil, "", "", spinner)
	assert.Equal(t, "dispatch", h.Name())
}

// --- Execute tests ---

func TestDispatch_Execute_Bad_UnknownAgent(t *testing.T) {
	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue", Active: true},
	})
	h := NewDispatchHandler(client, srv.URL, "test-token", spinner)

	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "nonexistent-agent",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		ChildNumber: 1,
	}

	_, err := h.Execute(context.Background(), sig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
}

func TestDispatch_Execute_Bad_InvalidQueueDir(t *testing.T) {
	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {
			Host:     "localhost",
			QueueDir: "/tmp/queue; touch /tmp/pwned",
			Active:   true,
		},
	})
	h := NewDispatchHandler(nil, "", "", spinner)

	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		ChildNumber: 1,
	}

	_, err := h.Execute(context.Background(), sig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent queue dir")
}

func TestDispatch_TicketJSON_Good(t *testing.T) {
	ticket := DispatchTicket{
		ID:           "host-uk-core-5-1234567890",
		RepoOwner:    "host-uk",
		RepoName:     "core",
		IssueNumber:  5,
		IssueTitle:   "Fix the thing",
		IssueBody:    "Please fix this bug",
		TargetBranch: "new",
		EpicNumber:   3,
		ForgeURL:     "https://forge.lthn.ai",
		ForgeUser:    "darbs-claude",
		Model:        "sonnet",
		Runner:       "claude",
		DualRun:      false,
		CreatedAt:    "2026-02-09T12:00:00Z",
	}

	data, err := json.MarshalIndent(ticket, "", "  ")
	require.NoError(t, err)

	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "host-uk-core-5-1234567890", decoded["id"])
	assert.Equal(t, "host-uk", decoded["repo_owner"])
	assert.Equal(t, "core", decoded["repo_name"])
	assert.Equal(t, float64(5), decoded["issue_number"])
	assert.Equal(t, "Fix the thing", decoded["issue_title"])
	assert.Equal(t, "Please fix this bug", decoded["issue_body"])
	assert.Equal(t, "new", decoded["target_branch"])
	assert.Equal(t, float64(3), decoded["epic_number"])
	assert.Equal(t, "https://forge.lthn.ai", decoded["forge_url"])
	assert.Equal(t, "darbs-claude", decoded["forgejo_user"])
	assert.Equal(t, "sonnet", decoded["model"])
	assert.Equal(t, "claude", decoded["runner"])
	// Token should NOT be present in the ticket.
	_, hasToken := decoded["forge_token"]
	assert.False(t, hasToken, "forge_token must not be in ticket JSON")
}

func TestDispatch_TicketJSON_Good_DualRun(t *testing.T) {
	ticket := DispatchTicket{
		ID:          "test-dual",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		IssueNumber: 1,
		ForgeURL:    "https://forge.lthn.ai",
		Model:       "gemini-2.0-flash",
		VerifyModel: "gemini-1.5-pro",
		DualRun:     true,
	}

	data, err := json.Marshal(ticket)
	require.NoError(t, err)

	var roundtrip DispatchTicket
	err = json.Unmarshal(data, &roundtrip)
	require.NoError(t, err)
	assert.True(t, roundtrip.DualRun)
	assert.Equal(t, "gemini-1.5-pro", roundtrip.VerifyModel)
}

func TestDispatch_TicketJSON_Good_OmitsEmptyModelRunner(t *testing.T) {
	ticket := DispatchTicket{
		ID:           "test-1",
		RepoOwner:    "host-uk",
		RepoName:     "core",
		IssueNumber:  1,
		TargetBranch: "new",
		ForgeURL:     "https://forge.lthn.ai",
	}

	data, err := json.MarshalIndent(ticket, "", "  ")
	require.NoError(t, err)

	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	_, hasModel := decoded["model"]
	_, hasRunner := decoded["runner"]
	assert.False(t, hasModel, "model should be omitted when empty")
	assert.False(t, hasRunner, "runner should be omitted when empty")
}

func TestDispatch_runRemote_Good_EscapesPath(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "ssh-output.txt")
	toolPath := writeFakeSSHCommand(t, outputPath)
	t.Setenv("PATH", toolPath+":"+os.Getenv("PATH"))

	h := NewDispatchHandler(nil, "", "", newTestSpinner(nil))
	dangerousPath := "/tmp/queue with spaces; touch /tmp/pwned"
	err := h.runRemote(
		context.Background(),
		agentci.AgentConfig{Host: "localhost"},
		"rm",
		"-f",
		dangerousPath,
	)
	require.NoError(t, err)

	output, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(output), "rm '-f' '"+dangerousPath+"'\n")
}

func TestDispatch_secureTransfer_Good_EscapesPath(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "ssh-output.txt")
	toolPath := writeFakeSSHCommand(t, outputPath)
	t.Setenv("PATH", toolPath+":"+os.Getenv("PATH"))

	h := NewDispatchHandler(nil, "", "", newTestSpinner(nil))
	dangerousPath := "/tmp/queue with spaces; touch /tmp/pwned"
	err := h.secureTransfer(
		context.Background(),
		agentci.AgentConfig{Host: "localhost"},
		dangerousPath,
		[]byte("hello"),
		0644,
	)
	require.NoError(t, err)

	output, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(output), "cat > '"+dangerousPath+"' && chmod 644 '"+dangerousPath+"'")

	inputPath := outputPath + ".stdin"
	input, err := os.ReadFile(inputPath)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(input))
}

func TestDispatch_TicketJSON_Good_ModelRunnerVariants(t *testing.T) {
	tests := []struct {
		name   string
		model  string
		runner string
	}{
		{"claude-sonnet", "sonnet", "claude"},
		{"claude-opus", "opus", "claude"},
		{"codex-default", "", "codex"},
		{"gemini-default", "", "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := DispatchTicket{
				ID:           "test-" + tt.name,
				RepoOwner:    "host-uk",
				RepoName:     "core",
				IssueNumber:  1,
				TargetBranch: "new",
				ForgeURL:     "https://forge.lthn.ai",
				Model:        tt.model,
				Runner:       tt.runner,
			}

			data, err := json.Marshal(ticket)
			require.NoError(t, err)

			var roundtrip DispatchTicket
			err = json.Unmarshal(data, &roundtrip)
			require.NoError(t, err)
			assert.Equal(t, tt.model, roundtrip.Model)
			assert.Equal(t, tt.runner, roundtrip.Runner)
		})
	}
}

func TestDispatch_Execute_Good_PostsComment(t *testing.T) {
	var commentPosted bool
	var commentBody string

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/host-uk/core/labels":
			json.NewEncoder(w).Encode([]any{})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/host-uk/core/labels":
			json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "in-progress", "color": "#1d76db"})
			return

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/host-uk/core/issues/5":
			json.NewEncoder(w).Encode(map[string]any{"id": 5, "number": 5, "labels": []any{}, "title": "Test"})
			return

		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/repos/host-uk/core/issues/5":
			json.NewEncoder(w).Encode(map[string]any{"id": 5, "number": 5})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/host-uk/core/issues/5/labels":
			json.NewEncoder(w).Encode([]any{map[string]any{"id": 1, "name": "in-progress"}})
			return

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/host-uk/core/issues/5/comments":
			commentPosted = true
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			commentBody = body["body"]
			json.NewEncoder(w).Encode(map[string]any{"id": 1, "body": body["body"]})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{})
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	spinner := newTestSpinner(map[string]agentci.AgentConfig{
		"darbs-claude": {Host: "localhost", QueueDir: "/tmp/nonexistent-queue", Active: true},
	})
	h := NewDispatchHandler(client, srv.URL, "test-token", spinner)

	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		ChildNumber: 5,
		EpicNumber:  3,
		IssueTitle:  "Test issue",
		IssueBody:   "Test body",
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.Equal(t, "dispatch", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core", result.RepoName)
	assert.Equal(t, 3, result.EpicNumber)
	assert.Equal(t, 5, result.ChildNumber)

	if result.Success {
		assert.True(t, commentPosted)
		assert.Contains(t, commentBody, "darbs-claude")
	}
}
