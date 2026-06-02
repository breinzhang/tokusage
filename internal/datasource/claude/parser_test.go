package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJSONLExtractsUsageEvents(t *testing.T) {
	events, warnings, err := ParseJSONLFile(filepath.Join("testdata", "basic.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v", warnings)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}

	first := events[0]
	if first.SessionID != "session-a" || first.MessageID != "msg-1" {
		t.Fatalf("first event IDs = %s %s", first.SessionID, first.MessageID)
	}
	if first.Model != "claude-sonnet-4.5" {
		t.Fatalf("Model = %q", first.Model)
	}
	if first.Tokens.StandardInputTokens != 100 || first.Tokens.OutputTokens != 20 || first.Tokens.CacheWrite5mTokens != 10 || first.Tokens.CacheWrite1hTokens != 20 || first.Tokens.CacheReadTokens != 40 {
		t.Fatalf("tokens = %+v", first.Tokens)
	}

	second := events[1]
	if second.Model != "glm-5.1" {
		t.Fatalf("second model = %q", second.Model)
	}
	if second.Tokens.StandardInputTokens != 200 || second.Tokens.OutputTokens != 50 {
		t.Fatalf("second tokens = %+v", second.Tokens)
	}
}

func TestParseJSONLDeduplicatesMessageID(t *testing.T) {
	events, warnings, err := ParseJSONLFile(filepath.Join("testdata", "duplicate-message.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v", warnings)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
}

func TestParseJSONLMarksSubagent(t *testing.T) {
	events, _, err := ParseJSONLFile(filepath.Join("testdata", "subagent.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if !events[0].IsSubagent {
		t.Fatal("IsSubagent = false, want true")
	}
	if events[0].AgentID != "agent-123" || events[0].ParentAgentID != "parent-456" {
		t.Fatalf("agent IDs = %q %q", events[0].AgentID, events[0].ParentAgentID)
	}
}

func TestFallbackMessageIDIncludesTokenCounts(t *testing.T) {
	base := rawRecord{
		Timestamp: "2026-05-10T01:02:03.000Z",
		Message: rawMessage{
			Model: "claude-sonnet-4.5",
			Usage: &rawUsage{
				InputTokens:              100,
				OutputTokens:             20,
				CacheCreationInputTokens: 30,
				CacheReadInputTokens:     40,
				CacheCreation: &rawCacheCreation{
					Ephemeral5mInputTokens: 10,
					Ephemeral1hInputTokens: 20,
				},
			},
		},
	}
	changedTokens := base
	changedTokens.Message.Usage = &rawUsage{
		InputTokens:              101,
		OutputTokens:             20,
		CacheCreationInputTokens: 30,
		CacheReadInputTokens:     40,
		CacheCreation: &rawCacheCreation{
			Ephemeral5mInputTokens: 10,
			Ephemeral1hInputTokens: 20,
		},
	}

	first := fallbackMessageID("source.jsonl", 7, base)
	second := fallbackMessageID("source.jsonl", 7, changedTokens)
	if first == second {
		t.Fatalf("fallbackMessageID did not change with token counts: %q", first)
	}
	if got := first[:len("fallback-")]; got != "fallback-" {
		t.Fatalf("fallback prefix = %q, want fallback-", got)
	}
}

func TestParseJSONLHandlesLongLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.jsonl")
	longText := strings.Repeat("x", 70*1024)
	line := fmt.Sprintf(`{"type":"assistant","sessionId":"session-long","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-long","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[{"type":"text","text":"%s"}],"usage":{"input_tokens":1,"output_tokens":2}}}`+"\n", longText)
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}

	events, warnings, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v", warnings)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
}

func TestParseJSONLResolvesProjectBeforeSubagentsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "-Users-example-work-repo", "subagents", "agent-a")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "session.jsonl")
	line := `{"type":"assistant","sessionId":"session-sub","timestamp":"2026-05-10T02:00:00.000Z","cwd":"","message":{"id":"msg-sub-path","type":"message","role":"assistant","model":"claude-haiku-4.5","content":[],"usage":{"input_tokens":12,"output_tokens":8}}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}

	events, warnings, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %+v", warnings)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if !events[0].IsSubagent {
		t.Fatal("IsSubagent = false, want true")
	}
	if events[0].ProjectName != "repo" || events[0].ProjectPath != "/Users/example/work/repo" {
		t.Fatalf("project = %q %q, want repo /Users/example/work/repo", events[0].ProjectName, events[0].ProjectPath)
	}
}

func TestParseJSONLWarningsAndBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "warnings.jsonl")
	lines := strings.Join([]string{
		`{`,
		`{"type":"assistant","sessionId":"session-warnings","timestamp":"not-a-time","cwd":"/Users/example/work/repo","message":{"id":"msg-bad-time","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[],"usage":{"input_tokens":1,"output_tokens":2}}}`,
		`{"type":"assistant","sessionId":"session-warnings","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-boundary","type":"message","role":"assistant","model":"","content":[],"usage":{"input_tokens":3,"output_tokens":4,"cache_creation_input_tokens":25,"cache_read_input_tokens":6,"cache_creation":{"ephemeral_5m_input_tokens":10,"ephemeral_1h_input_tokens":5}}}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(lines), 0o600); err != nil {
		t.Fatal(err)
	}

	events, warnings, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if len(warnings) != 3 {
		t.Fatalf("warnings = %+v, want 3 warnings", warnings)
	}
	if warnings[0].Message != "malformed JSONL" {
		t.Fatalf("first warning = %q, want malformed JSONL", warnings[0].Message)
	}
	if warnings[1].Message != "missing or invalid timestamp" {
		t.Fatalf("second warning = %q, want missing or invalid timestamp", warnings[1].Message)
	}
	if warnings[2].Message != "cache creation tokens lacked full tier detail; assigned remainder to 5m" {
		t.Fatalf("third warning = %q, want cache creation remainder warning", warnings[2].Message)
	}
	event := events[0]
	if event.Model != "unknown" {
		t.Fatalf("Model = %q, want unknown", event.Model)
	}
	if event.Tokens.CacheWrite5mTokens != 20 || event.Tokens.CacheWrite1hTokens != 5 {
		t.Fatalf("cache write tokens = %+v, want 5m=20 1h=5", event.Tokens)
	}
}
