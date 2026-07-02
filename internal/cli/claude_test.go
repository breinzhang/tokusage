package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeChartHelpIncludesChartFlags(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"claude", "chart", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"--metric", "--split-by", `--split-by string`, `default "model"`, "--top", "--width", "--ascii", "--color"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help output missing %q:\n%s", want, text)
		}
	}
}

func TestClaudeHeatmapHelpIncludesHeatmapFlags(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"claude", "heatmap", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{"--from", "--to", "--timezone", "--color", "--exclude-subagents"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help output missing %q:\n%s", want, text)
		}
	}
}

func TestClaudeHeatmapCommandRendersContributionGrid(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(
		`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-04T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[],"usage":{"input_tokens":100,"output_tokens":20}}}` + "\n" +
			`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-08T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-2","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[],"usage":{"input_tokens":400,"output_tokens":100}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"claude", "heatmap",
		"--path", root,
		"--cache", filepath.Join(t.TempDir(), "tokusage.db"),
		"--from", "2026-05-01",
		"--to", "2026-05-14",
		"--timezone", "UTC",
		"--color", "always",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	for _, want := range []string{
		"2 active days from 2026-05-01 to 2026-05-14",
		"May",
		"Mon",
		"Wed",
		"Fri",
		"Less",
		"More",
		"\x1b[38;2;50;52;58m",
		"\x1b[38;2;232;126;67m",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("heatmap output missing %q:\n%s", want, text)
		}
	}
}

func TestClaudeReportHelpIncludesPricingFlag(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"claude", "report", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	text := out.String()
	if !strings.Contains(text, "--pricing") {
		t.Fatalf("help output missing pricing flag:\n%s", text)
	}
}

func TestClaudeModelsCreatesAndUsesDefaultPricingConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"claude", "models",
		"--path", root,
		"--cache", filepath.Join(t.TempDir(), "tokusage.db"),
		"--from", "2026-05-01",
		"--to", "2026-05-30",
		"--timezone", "UTC",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	defaultPricingPath := filepath.Join(home, ".tokusage", "pricing.toml")
	if _, err := os.Stat(defaultPricingPath); err != nil {
		t.Fatalf("default pricing config was not created: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "$5.80") {
		t.Fatalf("models output should use default GLM pricing:\n%s", text)
	}
	if strings.Contains(text, "$18.00") {
		t.Fatalf("models output used built-in Sonnet fallback pricing:\n%s", text)
	}
}

func TestClaudeOptionsDefaultsToRecentSevenDays(t *testing.T) {
	opts, err := claudeOptions([]string{"/tmp/claude-projects"}, "/tmp/tokusage.db", "", "", "UTC", "table", "day", false)
	if err != nil {
		t.Fatal(err)
	}

	if opts.From != "" || opts.To != "" {
		t.Fatalf("default date range = %q..%q, want open scan range", opts.From, opts.To)
	}
	if opts.RecentDataDays != 7 {
		t.Fatalf("RecentDataDays = %d, want 7", opts.RecentDataDays)
	}
}

func TestClaudeOptionsExplicitDayStillDefaultsToRecentSevenDays(t *testing.T) {
	opts, err := claudeOptionsWithGroupByFlag([]string{"/tmp/claude-projects"}, "/tmp/tokusage.db", "", "", "UTC", "table", "day", true, false)
	if err != nil {
		t.Fatal(err)
	}

	if opts.From != "" || opts.To != "" {
		t.Fatalf("explicit day date range = %q..%q, want open scan range", opts.From, opts.To)
	}
	if opts.RecentDataDays != 7 {
		t.Fatalf("RecentDataDays = %d, want 7", opts.RecentDataDays)
	}
}

func TestClaudeOptionsExplicitWeekDefaultsToRecentThreeWeeks(t *testing.T) {
	opts, err := claudeOptionsWithGroupByFlag([]string{"/tmp/claude-projects"}, "/tmp/tokusage.db", "", "", "UTC", "table", "week", true, false)
	if err != nil {
		t.Fatal(err)
	}

	if opts.From != "" || opts.To != "" {
		t.Fatalf("explicit week date range = %q..%q, want open scan range", opts.From, opts.To)
	}
	if opts.RecentDataWeeks != 3 {
		t.Fatalf("RecentDataWeeks = %d, want 3", opts.RecentDataWeeks)
	}
}

func TestClaudeOptionsPreservesExplicitDateRange(t *testing.T) {
	opts, err := claudeOptions([]string{"/tmp/claude-projects"}, "/tmp/tokusage.db", "2026-05-01", "", "UTC", "table", "day", false)
	if err != nil {
		t.Fatal(err)
	}

	if opts.From != "2026-05-01" || opts.To != "" {
		t.Fatalf("explicit date range = %q..%q, want preserved", opts.From, opts.To)
	}
	if opts.RecentDataDays != 0 || opts.RecentDataWeeks != 0 {
		t.Fatalf("recent data defaults = days %d weeks %d, want disabled", opts.RecentDataDays, opts.RecentDataWeeks)
	}
}
