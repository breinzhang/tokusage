package chart

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestRenderNonSplitTokenChart(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "day",
		SplitBy: SplitNone,
		From:    "2026-05-01",
		To:      "2026-05-30",
		Buckets: []Bucket{
			{Label: "2026-05-09", Value: decimal.NewFromInt(90_400)},
			{Label: "2026-05-10", Value: decimal.NewFromInt(4_300_000)},
		},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone, Width: 10, Color: ColorNever})

	for _, want := range []string{"Metric: tokens | Group: day | 2026-05-01..2026-05-30", "2026-05-09", "90.4K", "2026-05-10", "4.3M", "█"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderNonSplitChartWithOpenDateRange(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "day",
		SplitBy: SplitNone,
		Buckets: []Bucket{{Label: "2026-05-10", Value: decimal.NewFromInt(1)}},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone, Width: 10, Color: ColorNever})

	if !strings.Contains(out, "Metric: tokens | Group: day | all dates") {
		t.Fatalf("open range header missing all dates:\n%s", out)
	}
	if strings.Contains(out, " | ..") {
		t.Fatalf("open range header should not contain empty date range:\n%s", out)
	}
}

func TestRenderASCIIFallback(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "day",
		SplitBy: SplitNone,
		Buckets: []Bucket{{Label: "2026-05-10", Value: decimal.NewFromInt(1)}},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone, Width: 10, ASCII: true, Color: ColorNever})

	if !strings.Contains(out, "#") {
		t.Fatalf("ASCII output missing #:\n%s", out)
	}
	if strings.Contains(out, "█") {
		t.Fatalf("ASCII output should not contain unicode bar:\n%s", out)
	}
}

func TestRenderUsesClaudeCodeColorPalette(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "day",
		SplitBy: SplitNone,
		Buckets: []Bucket{{Label: "2026-05-10", Value: decimal.NewFromInt(1)}},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone, Width: 10, Color: ColorAlways})

	if !strings.Contains(out, "\x1b[38;5;208m") {
		t.Fatalf("output missing Claude Code orange bar color:\n%q", out)
	}
	if strings.Contains(out, "\x1b[31m") || strings.Contains(out, "\x1b[32m") {
		t.Fatalf("output should not use basic rainbow ANSI colors:\n%q", out)
	}
}

func TestRenderUsesRefinedSplitColorPalette(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "month",
		SplitBy: SplitProject,
		Buckets: []Bucket{{
			Label: "2026-05",
			Value: decimal.NewFromInt(3),
			Segments: []Segment{
				{Label: "workspace_infra", Value: decimal.NewFromInt(1)},
				{Label: "environment", Value: decimal.NewFromInt(1)},
				{Label: "info_site", Value: decimal.NewFromInt(1)},
			},
		}},
		Legend: []LegendItem{
			{Key: "1", Label: "workspace_infra", Value: decimal.NewFromInt(1), Share: decimal.RequireFromString("33.3")},
			{Key: "2", Label: "environment", Value: decimal.NewFromInt(1), Share: decimal.RequireFromString("33.3")},
			{Key: "3", Label: "info_site", Value: decimal.NewFromInt(1), Share: decimal.RequireFromString("33.3")},
		},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "month", SplitBy: SplitProject, Top: 3, Width: 9, Color: ColorAlways})

	for _, want := range []string{"\x1b[38;5;208m", "\x1b[38;5;180m", "\x1b[38;5;109m"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing refined split color %q:\n%q", want, out)
		}
	}
	if strings.Contains(out, "\x1b[38;5;215m") || strings.Contains(out, "\x1b[38;5;222m") {
		t.Fatalf("output should not use previous bright split colors:\n%q", out)
	}
}

func TestRenderAlignsBucketValueAndBarColumns(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "week",
		SplitBy: SplitNone,
		Buckets: []Bucket{
			{Label: "W1", Value: decimal.NewFromInt(1_000)},
			{Label: "2026-W10", Value: decimal.NewFromInt(2_000)},
		},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "week", SplitBy: SplitNone, Width: 10, ASCII: true, Color: ColorNever})

	first := renderedLineWithPrefix(t, out, "W1")
	second := renderedLineWithPrefix(t, out, "2026-W10")
	firstValueCol := strings.Index(first, "1K")
	secondValueCol := strings.Index(second, "2K")
	if firstValueCol != secondValueCol {
		t.Fatalf("value columns differ: %d != %d\n%s", firstValueCol, secondValueCol, out)
	}
	firstBarCol := strings.Index(first, "#")
	secondBarCol := strings.Index(second, "#")
	if firstBarCol != secondBarCol {
		t.Fatalf("bar columns differ: %d != %d\n%s", firstBarCol, secondBarCol, out)
	}
}

func TestRenderAddsSpacingBetweenBucketRowsNotSegments(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "month",
		SplitBy: SplitModel,
		Buckets: []Bucket{
			{
				Label: "2026-05-08",
				Value: decimal.NewFromInt(4),
				Segments: []Segment{
					{Label: "glm-5.1", Value: decimal.NewFromInt(2)},
					{Label: "glm-4.7", Value: decimal.NewFromInt(2)},
				},
			},
			{
				Label: "2026-05-09",
				Value: decimal.NewFromInt(4),
				Segments: []Segment{
					{Label: "glm-5.1", Value: decimal.NewFromInt(2)},
					{Label: "glm-4.7", Value: decimal.NewFromInt(2)},
				},
			},
		},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "month", SplitBy: SplitModel, Top: 2, Width: 4, ASCII: true, Color: ColorNever})

	if !strings.Contains(out, "##==") {
		t.Fatalf("split bar should not include spacing between segments:\n%s", out)
	}
	if strings.Contains(out, "## ==") {
		t.Fatalf("split bar should keep segment glyphs adjacent:\n%s", out)
	}
	if !strings.Contains(out, "\n\n2026-05-09") {
		t.Fatalf("bucket rows should have vertical spacing:\n%s", out)
	}
}

func TestRenderSplitCostChartWithLegend(t *testing.T) {
	view := Chart{
		Metric:  MetricCost,
		GroupBy: "month",
		SplitBy: SplitModel,
		From:    "2026-05-01",
		To:      "2026-05-30",
		Buckets: []Bucket{{
			Label: "2026-05",
			Value: decimal.RequireFromString("8.45"),
			Segments: []Segment{
				{Label: "glm-5.1", Value: decimal.RequireFromString("7.51")},
				{Label: "Other", Value: decimal.RequireFromString("0.94")},
			},
		}},
		Legend: []LegendItem{
			{Key: "1", Label: "glm-5.1", Value: decimal.RequireFromString("7.51"), Share: decimal.RequireFromString("88.9")},
			{Key: "+", Label: "Other", Value: decimal.RequireFromString("0.94"), Share: decimal.RequireFromString("11.1")},
		},
	}

	out := Render(view, Options{Metric: MetricCost, GroupBy: "month", SplitBy: SplitModel, Top: 1, Width: 10, Color: ColorNever})

	for _, want := range []string{"Metric: cost | Group: month | Split: model | Top: 1", "2026-05", "$8.45", "█", "▓", "Legend", "1  glm-5.1", "$7.51", "88.9%", "+  Other"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("ColorNever output contains ANSI escapes:\n%s", out)
	}
}

func TestRenderSplitLegendUsesMatchingBarColors(t *testing.T) {
	view := Chart{
		Metric:  MetricCost,
		GroupBy: "month",
		SplitBy: SplitModel,
		Buckets: []Bucket{{
			Label: "2026-05",
			Value: decimal.RequireFromString("3.00"),
			Segments: []Segment{
				{Label: "glm-5.1", Value: decimal.RequireFromString("2.00")},
				{Label: "glm-4.7", Value: decimal.RequireFromString("1.00")},
			},
		}},
		Legend: []LegendItem{
			{Key: "1", Label: "glm-5.1", Value: decimal.RequireFromString("2.00"), Share: decimal.RequireFromString("66.7")},
			{Key: "2", Label: "glm-4.7", Value: decimal.RequireFromString("1.00"), Share: decimal.RequireFromString("33.3")},
		},
	}

	out := Render(view, Options{Metric: MetricCost, GroupBy: "month", SplitBy: SplitModel, Top: 2, Width: 10, Color: ColorAlways})

	for _, want := range []string{
		"\x1b[38;5;208m█",
		"\x1b[38;5;180m█",
		"\x1b[38;5;208m1  glm-5.1",
		"\x1b[38;5;180m2  glm-4.7",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing matching legend/bar color %q:\n%q", want, out)
		}
	}
	if strings.Contains(out, "▓") {
		t.Fatalf("colored split bars should use solid blocks only:\n%q", out)
	}
}

func TestRenderAlignsLongLegendLabels(t *testing.T) {
	view := Chart{
		Metric:  MetricCost,
		GroupBy: "month",
		SplitBy: SplitModel,
		Buckets: []Bucket{{
			Label: "2026-05",
			Value: decimal.RequireFromString("3.00"),
			Segments: []Segment{
				{Label: "short", Value: decimal.RequireFromString("1.00")},
				{Label: "very-long-model-name", Value: decimal.RequireFromString("2.00")},
			},
		}},
		Legend: []LegendItem{
			{Key: "1", Label: "short", Value: decimal.RequireFromString("1.00"), Share: decimal.RequireFromString("33.3")},
			{Key: "2", Label: "very-long-model-name", Value: decimal.RequireFromString("2.00"), Share: decimal.RequireFromString("66.7")},
		},
	}

	out := Render(view, Options{Metric: MetricCost, GroupBy: "month", SplitBy: SplitModel, Top: 2, Width: 10, ASCII: true, Color: ColorNever})

	first := renderedLineWithPrefix(t, out, "1  short")
	second := renderedLineWithPrefix(t, out, "2  very-long-model-name")
	firstValueCol := strings.Index(first, "$1.00")
	secondValueCol := strings.Index(second, "$2.00")
	if firstValueCol != secondValueCol {
		t.Fatalf("legend value columns differ: %d != %d\n%s", firstValueCol, secondValueCol, out)
	}
	firstPercentCol := strings.Index(first, "33.3%")
	secondPercentCol := strings.Index(second, "66.7%")
	if firstPercentCol != secondPercentCol {
		t.Fatalf("legend percent columns differ: %d != %d\n%s", firstPercentCol, secondPercentCol, out)
	}
}

func TestRenderSplitChartCapsBarWidthWhenSegmentsExceedWidth(t *testing.T) {
	view := Chart{
		Metric:  MetricTokens,
		GroupBy: "month",
		SplitBy: SplitModel,
		Buckets: []Bucket{{
			Label: "2026-05",
			Value: decimal.NewFromInt(5),
			Segments: []Segment{
				{Label: "a", Value: decimal.NewFromInt(1)},
				{Label: "b", Value: decimal.NewFromInt(1)},
				{Label: "c", Value: decimal.NewFromInt(1)},
				{Label: "d", Value: decimal.NewFromInt(1)},
				{Label: "e", Value: decimal.NewFromInt(1)},
			},
		}},
	}

	out := Render(view, Options{Metric: MetricTokens, GroupBy: "month", SplitBy: SplitModel, Top: 5, Width: 3, ASCII: true, Color: ColorNever})

	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "2026-05") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			t.Fatalf("bucket line missing bar field:\n%s", out)
		}
		if got := len([]rune(fields[2])); got > 3 {
			t.Fatalf("bar width = %d, want <= 3:\n%s", got, out)
		}
		return
	}
	t.Fatalf("bucket line missing:\n%s", out)
}

func TestRenderEmptyChart(t *testing.T) {
	out := Render(Chart{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone}, Options{Metric: MetricTokens, GroupBy: "day", SplitBy: SplitNone, Width: 10, Color: ColorNever})

	if strings.TrimSpace(out) != "No usage data for the selected range." {
		t.Fatalf("empty output = %q", out)
	}
}

func renderedLineWithPrefix(t *testing.T, out string, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("line with prefix %q missing:\n%s", prefix, out)
	return ""
}
