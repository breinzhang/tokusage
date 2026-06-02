package output

import (
	"strings"
	"testing"

	"github.com/breinzhang/tokusage/internal/domain"
)

func TestRenderTableCompactColumns(t *testing.T) {
	report := Report{
		LabelHeader: "Date",
		Buckets: []Bucket{{
			Label: "2026-05-10",
			Tokens: domain.TokenSummary{
				StandardInputTokens: 10,
				CacheWrite5mTokens:  2,
				CacheWrite1hTokens:  3,
				CacheReadTokens:     4,
				OutputTokens:        5,
			},
			EstimatedCost: "$0.12 partial",
		}},
	}

	table := RenderTable(report)

	for _, want := range []string{"Date", "Input", "Cache Write", "Cache Read", "Output", "Total", "Est. Cost", "2026-05-10", "$0.12 partial"} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing %q:\n%s", want, table)
		}
	}
}

func TestRenderTableUsesCustomLabelHeader(t *testing.T) {
	report := Report{
		LabelHeader: "Model",
		Buckets: []Bucket{{
			Label: "glm-5.1",
		}},
	}

	table := RenderTable(report)

	if !strings.Contains(table, "Model") {
		t.Fatalf("table missing custom label header:\n%s", table)
	}
	if strings.Contains(table, "Date") {
		t.Fatalf("table should not contain Date header for model report:\n%s", table)
	}
}

func TestRenderTableFormatsTokenCountsWithUnits(t *testing.T) {
	report := Report{
		LabelHeader: "Project",
		Buckets: []Bucket{{
			Label: "/Users/example/work/repo",
			Tokens: domain.TokenSummary{
				StandardInputTokens: 1234,
				CacheReadTokens:     1_234_567,
				OutputTokens:        1_234_567_890,
			},
		}},
	}

	table := RenderTable(report)

	for _, want := range []string{"1.2K", "1.2M", "1.2B"} {
		if !strings.Contains(table, want) {
			t.Fatalf("table missing formatted token count %q:\n%s", want, table)
		}
	}
	if strings.Contains(table, "1,234") {
		t.Fatalf("table should not use comma-separated token counts:\n%s", table)
	}
}

func TestRenderTableAlignsRowsWithLongLabels(t *testing.T) {
	report := Report{
		LabelHeader: "Project",
		Buckets: []Bucket{
			{
				Label: "basic_dev",
				Tokens: domain.TokenSummary{
					StandardInputTokens: 808,
				},
			},
			{
				Label: "workspace_infra",
				Tokens: domain.TokenSummary{
					StandardInputTokens: 300_400,
				},
			},
		},
	}

	table := RenderTable(report)
	lines := strings.Split(strings.TrimSpace(table), "\n")
	if len(lines) != 3 {
		t.Fatalf("table lines = %d, want 3:\n%s", len(lines), table)
	}
	headerInputColumn := strings.Index(lines[0], "Input")
	longRowInputColumn := strings.Index(lines[2], "300.4K")
	if headerInputColumn == -1 || longRowInputColumn == -1 {
		t.Fatalf("table missing expected cells:\n%s", table)
	}
	headerInputEnd := headerInputColumn + len("Input")
	longRowInputEnd := longRowInputColumn + len("300.4K")
	if longRowInputEnd != headerInputEnd {
		t.Fatalf("long row input end column = %d, header input end column = %d:\n%s", longRowInputEnd, headerInputEnd, table)
	}
}

func TestRenderTableColorsLeastAndMostUsageRows(t *testing.T) {
	report := Report{
		LabelHeader:  "Model",
		ColorEnabled: true,
		Buckets: []Bucket{
			{
				Label:  "least",
				Tokens: domain.TokenSummary{StandardInputTokens: 10},
			},
			{
				Label:  "middle",
				Tokens: domain.TokenSummary{StandardInputTokens: 20},
			},
			{
				Label:  "most",
				Tokens: domain.TokenSummary{StandardInputTokens: 30},
			},
		},
	}

	table := RenderTable(report)

	if !strings.Contains(table, "\x1b[38;5;95mleast") {
		t.Fatalf("least usage row missing less color:\n%q", table)
	}
	if !strings.Contains(table, "\x1b[38;5;208mmost") {
		t.Fatalf("most usage row missing more color:\n%q", table)
	}
	if strings.Contains(table, "\x1b[38;5;95mmiddle") || strings.Contains(table, "\x1b[38;5;208mmiddle") {
		t.Fatalf("middle row should not be colored:\n%q", table)
	}
}

func TestRenderTableDoesNotColorSingleRow(t *testing.T) {
	report := Report{
		LabelHeader:  "Model",
		ColorEnabled: true,
		Buckets: []Bucket{{
			Label:  "only",
			Tokens: domain.TokenSummary{StandardInputTokens: 10},
		}},
	}

	table := RenderTable(report)

	if strings.Contains(table, "\x1b[") {
		t.Fatalf("single row table should not be colored:\n%q", table)
	}
}
