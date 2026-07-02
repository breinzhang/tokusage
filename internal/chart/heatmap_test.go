package chart

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/breinzhang/tokusage/internal/cache"
	"github.com/breinzhang/tokusage/internal/domain"
)

func TestBuildHeatmapDefaultsToLastYearEndingAtLatestUsageDate(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2025-04-01", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", Tokens: domain.TokenSummary{StandardInputTokens: 200}},
		{Date: "2026-05-09", Tokens: domain.TokenSummary{StandardInputTokens: 300}},
	}

	got, err := BuildHeatmap(rollups, HeatmapOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if got.From != "2025-05-11" || got.To != "2026-05-10" {
		t.Fatalf("range = %s..%s, want 2025-05-11..2026-05-10", got.From, got.To)
	}
	if got.ActiveDays != 2 {
		t.Fatalf("ActiveDays = %d, want 2", got.ActiveDays)
	}
	if len(got.Days) != 365 {
		t.Fatalf("Days = %d, want 365", len(got.Days))
	}
	if got.TotalTokens != 500 {
		t.Fatalf("TotalTokens = %d, want 500", got.TotalTokens)
	}
}

func TestBuildHeatmapAggregatesDailyRollups(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-04", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-04", Tokens: domain.TokenSummary{OutputTokens: 25}},
		{Date: "2026-05-08", Tokens: domain.TokenSummary{StandardInputTokens: 500}},
	}

	got, err := BuildHeatmap(rollups, HeatmapOptions{From: "2026-05-01", To: "2026-05-14"})
	if err != nil {
		t.Fatal(err)
	}

	values := map[string]int64{}
	for _, day := range got.Days {
		if day.Value > 0 {
			values[day.Date] = day.Value
		}
	}
	if values["2026-05-04"] != 125 {
		t.Fatalf("2026-05-04 value = %d, want 125", values["2026-05-04"])
	}
	if values["2026-05-08"] != 500 {
		t.Fatalf("2026-05-08 value = %d, want 500", values["2026-05-08"])
	}
}

func TestRenderHeatmapIncludesGithubStyleLabelsAndClaudeOrangeGradient(t *testing.T) {
	view, err := BuildHeatmap([]cache.DailyRollup{
		{Date: "2026-05-04", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-08", Tokens: domain.TokenSummary{StandardInputTokens: 500}},
	}, HeatmapOptions{From: "2026-05-01", To: "2026-05-14"})
	if err != nil {
		t.Fatal(err)
	}

	out := RenderHeatmap(view, HeatmapOptions{Color: ColorAlways})

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
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[38;2;156;160;170m") {
		t.Fatalf("active heatmap output should not use gray high-intensity color:\n%s", out)
	}
}

func TestRenderHeatmapUsesCompactWeekColumns(t *testing.T) {
	view, err := BuildHeatmap([]cache.DailyRollup{
		{Date: "2026-05-04", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
	}, HeatmapOptions{From: "2026-05-01", To: "2026-05-14"})
	if err != nil {
		t.Fatal(err)
	}

	out := RenderHeatmap(view, HeatmapOptions{Color: ColorNever})
	mon := heatmapLineWithPrefix(t, out, "Mon ")
	cells := strings.TrimPrefix(mon, "Mon ")

	if strings.Contains(cells, " ") {
		t.Fatalf("heatmap cells should not use regular separator spaces: %q\n%s", cells, out)
	}
	if !strings.Contains(cells, heatmapCellSeparator) {
		t.Fatalf("heatmap cells should use thin separators for small gaps: %q\n%s", cells, out)
	}
	cellCount := strings.Count(cells, heatmapCellGlyph) + strings.Count(cells, heatmapEmptyCellGlyph)
	if got, want := cellCount, len(heatmapWeekStarts(view)); got != want {
		t.Fatalf("cell glyph count = %d, want %d\n%s", got, want, out)
	}
}

func TestRenderHeatmapMonthLabelsAlignToWeekSlots(t *testing.T) {
	view, err := BuildHeatmap(nil, HeatmapOptions{From: "2026-01-04", To: "2026-03-14"})
	if err != nil {
		t.Fatal(err)
	}

	line := renderHeatmapMonthLabels(view)
	weeks := heatmapWeekStarts(view)
	expected := map[string]string{
		"Jan": "2026-01-04",
		"Feb": "2026-02-01",
		"Mar": "2026-03-01",
	}
	for label, date := range expected {
		weekIndex := heatmapWeekIndexForDate(t, weeks, date)
		want := heatmapMonthLabelRuneOffset(weekIndex)
		got := runeIndex(line, label)
		if got != want {
			t.Fatalf("%s starts at rune %d, want %d for week %d\n%s", label, got, want, weekIndex, line)
		}
	}
}

func TestRenderHeatmapMonthLabelsDoNotOverlapAtRangeStart(t *testing.T) {
	view, err := BuildHeatmap(nil, HeatmapOptions{From: "2025-06-25", To: "2025-08-09"})
	if err != nil {
		t.Fatal(err)
	}

	line := renderHeatmapMonthLabels(view)
	if strings.Contains(line, "JuJul") {
		t.Fatalf("month labels should not overlap at range start:\n%s", line)
	}
	if strings.Contains(line, "Jun") {
		t.Fatalf("partial start month should be skipped when it overlaps the next month:\n%s", line)
	}
	if !strings.Contains(line, "Jul") || !strings.Contains(line, "Aug") {
		t.Fatalf("month labels should keep full month labels:\n%s", line)
	}
}

func TestRenderHeatmapCellsUseOriginalSizeSquareGlyph(t *testing.T) {
	colored := renderHeatmapCell(4, HeatmapOptions{Color: ColorAlways})
	for _, forbidden := range []string{"◼", "▪"} {
		if strings.Contains(colored, forbidden) {
			t.Fatalf("colored heatmap cell should not use wrong-size glyph %q: %q", forbidden, colored)
		}
	}
	if !strings.Contains(colored, "■") {
		t.Fatalf("colored heatmap cell should use original-size square glyph: %q", colored)
	}

	active := renderHeatmapCell(4, HeatmapOptions{Color: ColorNever})
	if active != "■" {
		t.Fatalf("active uncolored cell = %q, want original-size square glyph", active)
	}
}

func TestHeatmapUsesGrayInactiveAndSingleHueOrangeActiveGradient(t *testing.T) {
	want := []string{
		"38;2;50;52;58",
		"38;2;119;65;44",
		"38;2;162;84;49",
		"38;2;203;106;57",
		"38;2;232;126;67",
	}
	if !reflect.DeepEqual(heatmapColors, want) {
		t.Fatalf("heatmapColors = %#v, want gray inactive plus single-hue orange active gradient %#v", heatmapColors, want)
	}
}

func heatmapLineWithPrefix(t *testing.T, out string, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("line with prefix %q missing:\n%s", prefix, out)
	return ""
}

func heatmapWeekIndexForDate(t *testing.T, weeks []time.Time, dateText string) int {
	t.Helper()
	date, err := parseHeatmapDate(dateText)
	if err != nil {
		t.Fatal(err)
	}
	for i, week := range weeks {
		if !date.Before(week) && date.Before(week.AddDate(0, 0, 7)) {
			return i
		}
	}
	t.Fatalf("week for date %s missing in %#v", dateText, weeks)
	return -1
}

func heatmapMonthLabelRuneOffset(weekIndex int) int {
	return heatmapLabelWidth + weekIndex*(len([]rune(heatmapCellGlyph))+len([]rune(heatmapCellSeparator)))
}

func runeIndex(value string, needle string) int {
	byteIndex := strings.Index(value, needle)
	if byteIndex < 0 {
		return -1
	}
	return len([]rune(value[:byteIndex]))
}
