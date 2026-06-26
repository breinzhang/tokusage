package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/breinzhang/tokusage/internal/cache"
)

func TestRunReportOnFixtureDirectory(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[],"usage":{"input_tokens":100,"output_tokens":20,"cache_creation_input_tokens":30,"cache_read_input_tokens":40,"cache_creation":{"ephemeral_5m_input_tokens":10,"ephemeral_1h_input_tokens":20}}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeReport(context.Background(), Options{
		Paths:     []string{root},
		CachePath: filepath.Join(t.TempDir(), "tokusage.db"),
		From:      "2026-05-01",
		To:        "2026-05-30",
		Format:    "table",
		GroupBy:   "day",
		Timezone:  "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2026-05-10") {
		t.Fatalf("output missing date:\n%s", out)
	}
	if !strings.Contains(out, "190") {
		t.Fatalf("output missing total token count 190:\n%s", out)
	}
}

func TestRunCacheStatusAndClear(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "tokusage.db")

	status, err := RunCacheStatus(context.Background(), cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "DB: "+cachePath) || !strings.Contains(status, "Files: 0") {
		t.Fatalf("unexpected status:\n%s", status)
	}

	if err := RunCacheClear(cachePath); err != nil {
		t.Fatal(err)
	}
	status, err = RunCacheStatus(context.Background(), cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "Events: 0") || !strings.Contains(status, "Rollups: 0") {
		t.Fatalf("unexpected empty status:\n%s", status)
	}
}

func TestRunClaudeModelsSummarizesByModel(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(
		`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"claude-sonnet-4.5","content":[],"usage":{"input_tokens":100,"output_tokens":20}}}` + "\n" +
			`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:03:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-2","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":200,"output_tokens":50}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeModels(context.Background(), Options{
		Paths:     []string{root},
		CachePath: filepath.Join(t.TempDir(), "tokusage.db"),
		From:      "2026-05-01",
		To:        "2026-05-30",
		Format:    "table",
		Timezone:  "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Model", "claude-sonnet-4.5", "glm-5.1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("models output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Date") {
		t.Fatalf("models output should not use Date header:\n%s", out)
	}
	if strings.Contains(out, "N/A") {
		t.Fatalf("unknown model should use Anthropic fallback pricing:\n%s", out)
	}
}

func TestRunClaudeModelsUsesPricingConfig(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	pricingPath := filepath.Join(t.TempDir(), "pricing.toml")
	pricingConfig := []byte(`
[[models]]
pattern = "glm-5*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.00"
cache_write_1h_per_mtok = "1.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "2.00"
`)
	if err := os.WriteFile(pricingPath, pricingConfig, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeModels(context.Background(), Options{
		Paths:       []string{root},
		CachePath:   filepath.Join(t.TempDir(), "tokusage.db"),
		From:        "2026-05-01",
		To:          "2026-05-30",
		Format:      "table",
		Timezone:    "UTC",
		PricingPath: pricingPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "$3.00") {
		t.Fatalf("models output should use configured GLM pricing:\n%s", out)
	}
	if strings.Contains(out, "$18.00") {
		t.Fatalf("models output used built-in Sonnet fallback pricing:\n%s", out)
	}
}

func TestRunClaudeProjectsSummarizesByProjectName(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeProjects(context.Background(), Options{
		Paths:     []string{root},
		CachePath: filepath.Join(t.TempDir(), "tokusage.db"),
		From:      "2026-05-01",
		To:        "2026-05-30",
		Format:    "table",
		Timezone:  "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Project", "repo", "1M", "$"} {
		if !strings.Contains(out, want) {
			t.Fatalf("projects output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "project_id") || strings.Contains(out, "5501de") {
		t.Fatalf("projects output should not show hashed project IDs:\n%s", out)
	}
	if strings.Contains(out, "Date") {
		t.Fatalf("projects output should not use Date header:\n%s", out)
	}
	if strings.Contains(out, "N/A") {
		t.Fatalf("unknown model should use Anthropic fallback pricing:\n%s", out)
	}
}

func TestReportBucketLabelFormatsMonthWeek(t *testing.T) {
	first := reportBucketLabel(Options{GroupBy: "week"}, cacheRollupDate("2026-05-10"))
	second := reportBucketLabel(Options{GroupBy: "week"}, cacheRollupDate("2026-05-11"))

	if first != "2026-05 W2" {
		t.Fatalf("first week label = %q, want 2026-05 W2", first)
	}
	if second != "2026-05 W3" {
		t.Fatalf("second week label = %q, want 2026-05 W3", second)
	}
}

func TestReportBucketLabelFormatsYear(t *testing.T) {
	got := reportBucketLabel(Options{GroupBy: "year"}, cacheRollupDate("2026-05-10"))

	if got != "2026" {
		t.Fatalf("year label = %q, want 2026", got)
	}
}

func TestApplyRecentDataWindowKeepsLatestSevenDataDays(t *testing.T) {
	rollups := []cache.DailyRollup{
		cacheRollupDate("2026-05-01"),
		cacheRollupDate("2026-05-03"),
		cacheRollupDate("2026-05-08"),
		cacheRollupDate("2026-05-10"),
		cacheRollupDate("2026-05-11"),
		cacheRollupDate("2026-05-13"),
		cacheRollupDate("2026-05-18"),
		cacheRollupDate("2026-06-02"),
	}

	got, opts := applyRecentDataWindow(rollups, Options{RecentDataDays: 7})

	if len(got) != 7 {
		t.Fatalf("len(filtered) = %d, want 7", len(got))
	}
	if got[0].Date != "2026-05-03" {
		t.Fatalf("first filtered date = %q, want 2026-05-03", got[0].Date)
	}
	if got[len(got)-1].Date != "2026-06-02" {
		t.Fatalf("last filtered date = %q, want 2026-06-02", got[len(got)-1].Date)
	}
	if opts.From != "2026-05-03" || opts.To != "2026-06-02" {
		t.Fatalf("effective range = %q..%q, want filtered data range", opts.From, opts.To)
	}
}

func TestApplyRecentDataWindowKeepsLatestThreeWeekBuckets(t *testing.T) {
	rollups := []cache.DailyRollup{
		cacheRollupDate("2026-05-01"),
		cacheRollupDate("2026-05-10"),
		cacheRollupDate("2026-05-11"),
		cacheRollupDate("2026-05-18"),
		cacheRollupDate("2026-06-02"),
	}

	got, opts := applyRecentDataWindow(rollups, Options{RecentDataWeeks: 3})

	if len(got) != 3 {
		t.Fatalf("len(filtered) = %d, want 3", len(got))
	}
	if got[0].Date != "2026-05-11" {
		t.Fatalf("first filtered date = %q, want 2026-05-11", got[0].Date)
	}
	if opts.From != "2026-05-11" || opts.To != "2026-06-02" {
		t.Fatalf("effective range = %q..%q, want filtered week data range", opts.From, opts.To)
	}
}

func TestApplyRecentDataWindowPreservesExplicitDateRange(t *testing.T) {
	rollups := []cache.DailyRollup{
		cacheRollupDate("2026-05-01"),
		cacheRollupDate("2026-05-03"),
	}

	got, opts := applyRecentDataWindow(rollups, Options{From: "2026-05-01", RecentDataDays: 7})

	if len(got) != 2 {
		t.Fatalf("len(filtered) = %d, want unchanged rollups", len(got))
	}
	if opts.From != "2026-05-01" || opts.To != "" {
		t.Fatalf("explicit range = %q..%q, want preserved", opts.From, opts.To)
	}
}

func TestRunClaudeChartOnFixtureDirectory(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(
		`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n" +
			`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-11T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-2","type":"message","role":"assistant","model":"glm-4.7","content":[],"usage":{"input_tokens":1000,"output_tokens":1000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeChart(context.Background(), ChartOptions{
		Options: Options{
			Paths:     []string{root},
			CachePath: filepath.Join(t.TempDir(), "tokusage.db"),
			From:      "2026-05-01",
			To:        "2026-05-30",
			GroupBy:   "day",
			Timezone:  "UTC",
		},
		Metric:  "tokens",
		SplitBy: "none",
		Top:     8,
		Width:   10,
		Color:   "never",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Metric: tokens", "2026-05-10", "2M", "2026-05-11", "2K"} {
		if !strings.Contains(out, want) {
			t.Fatalf("chart output missing %q:\n%s", want, out)
		}
	}
}

func TestRunClaudeChartUsesPricingConfig(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	pricingPath := filepath.Join(t.TempDir(), "pricing.toml")
	pricingConfig := []byte(`
[[models]]
pattern = "glm-5*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.00"
cache_write_1h_per_mtok = "1.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "2.00"
`)
	if err := os.WriteFile(pricingPath, pricingConfig, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeChart(context.Background(), ChartOptions{
		Options: Options{
			Paths:       []string{root},
			CachePath:   filepath.Join(t.TempDir(), "tokusage.db"),
			From:        "2026-05-01",
			To:          "2026-05-30",
			GroupBy:     "day",
			Timezone:    "UTC",
			PricingPath: pricingPath,
		},
		Metric:  "cost",
		SplitBy: "none",
		Top:     8,
		Width:   10,
		Color:   "never",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "$3.00") {
		t.Fatalf("chart output should use configured GLM pricing:\n%s", out)
	}
	if strings.Contains(out, "$18.00") {
		t.Fatalf("chart output used built-in Sonnet fallback pricing:\n%s", out)
	}
}

func TestRunClaudeChartValidatesBeforeScanning(t *testing.T) {
	_, err := RunClaudeChart(context.Background(), ChartOptions{
		Options: Options{
			Paths:     []string{"/definitely/missing"},
			CachePath: filepath.Join(t.TempDir(), "tokusage.db"),
			From:      "2026-05-01",
			To:        "2026-05-30",
			GroupBy:   "day",
			Timezone:  "UTC",
		},
		Metric:  "requests",
		SplitBy: "none",
		Top:     8,
		Width:   10,
		Color:   "never",
	})
	if err == nil || err.Error() != `unsupported metric "requests"` {
		t.Fatalf("error = %v, want unsupported metric before path scan", err)
	}
}

func TestRunClaudeChartExcludesSubagents(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "-Users-example-work-repo")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(
		`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:02:03.000Z","cwd":"/Users/example/work/repo","message":{"id":"msg-1","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000,"output_tokens":1000}}}` + "\n" +
			`{"type":"assistant","sessionId":"session-a","timestamp":"2026-05-10T01:03:03.000Z","cwd":"/Users/example/work/repo","agentId":"agent-1","message":{"id":"msg-2","type":"message","role":"assistant","model":"glm-5.1","content":[],"usage":{"input_tokens":1000000,"output_tokens":1000000}}}` + "\n")
	if err := os.WriteFile(filepath.Join(projectDir, "session-a.jsonl"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := RunClaudeChart(context.Background(), ChartOptions{
		Options: Options{
			Paths:            []string{root},
			CachePath:        filepath.Join(t.TempDir(), "tokusage.db"),
			From:             "2026-05-01",
			To:               "2026-05-30",
			GroupBy:          "day",
			Timezone:         "UTC",
			ExcludeSubagents: true,
		},
		Metric:  "tokens",
		SplitBy: "none",
		Top:     8,
		Width:   10,
		Color:   "never",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2K") {
		t.Fatalf("chart output missing top-level tokens:\n%s", out)
	}
	if strings.Contains(out, "2M") {
		t.Fatalf("chart output includes subagent tokens:\n%s", out)
	}
}

func cacheRollupDate(date string) cache.DailyRollup {
	return cache.DailyRollup{Date: date}
}
