package cli

import (
	"bytes"
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
