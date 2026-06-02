package chart

import (
	"testing"

	"github.com/breinzhang/tokusage/internal/cache"
	"github.com/breinzhang/tokusage/internal/domain"
	"github.com/breinzhang/tokusage/internal/pricing"
	"github.com/shopspring/decimal"
)

func TestBuildAggregatesTokenMetricByDay(t *testing.T) {
	rollups := []cache.DailyRollup{
		{
			Date:  "2026-05-10",
			Model: "claude-sonnet-4.5",
			Tokens: domain.TokenSummary{
				StandardInputTokens: 100,
				CacheWrite5mTokens:  10,
				CacheWrite1hTokens:  20,
				CacheReadTokens:     40,
				OutputTokens:        30,
			},
		},
		{
			Date:  "2026-05-10",
			Model: "glm-5.1",
			Tokens: domain.TokenSummary{
				StandardInputTokens: 200,
			},
		},
	}

	got, err := Build(rollups, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(got.Buckets))
	}
	if got.Buckets[0].Label != "2026-05-10" {
		t.Fatalf("label = %q, want 2026-05-10", got.Buckets[0].Label)
	}
	if got.Buckets[0].Value.String() != "400" {
		t.Fatalf("value = %s, want 400", got.Buckets[0].Value)
	}
}

func TestBuildAggregatesCostMetricByMonth(t *testing.T) {
	rollups := []cache.DailyRollup{
		{
			Date:  "2026-05-10",
			Model: "claude-sonnet-4.5",
			Tokens: domain.TokenSummary{
				StandardInputTokens: 1_000_000,
				OutputTokens:        1_000_000,
			},
		},
		{
			Date:  "2026-05-11",
			Model: "claude-sonnet-4.5",
			Tokens: domain.TokenSummary{
				CacheReadTokens: 1_000_000,
			},
		},
	}

	got, err := Build(rollups, Options{GroupBy: "month", Metric: MetricCost, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Buckets) != 1 {
		t.Fatalf("buckets = %d, want 1", len(got.Buckets))
	}
	if got.Buckets[0].Label != "2026-05" {
		t.Fatalf("label = %q, want 2026-05", got.Buckets[0].Label)
	}
	if got.Buckets[0].Value.StringFixed(2) != "18.30" {
		t.Fatalf("value = %s, want 18.30", got.Buckets[0].Value.StringFixed(2))
	}
}

func TestBuildAggregatesTokenMetricByMonthWeek(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "claude-sonnet-4.5", Tokens: domain.TokenSummary{StandardInputTokens: 1}},
		{Date: "2026-05-11", Model: "claude-sonnet-4.5", Tokens: domain.TokenSummary{StandardInputTokens: 2}},
	}

	got, err := Build(rollups, Options{GroupBy: "week", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Buckets) != 2 {
		t.Fatalf("buckets = %d, want 2 for two month weeks", len(got.Buckets))
	}
	if got.Buckets[0].Label != "2026-05 W2" || got.Buckets[0].Value.String() != "1" {
		t.Fatalf("first bucket = %+v, want 2026-05 W2 value 1", got.Buckets[0])
	}
	if got.Buckets[1].Label != "2026-05 W3" || got.Buckets[1].Value.String() != "2" {
		t.Fatalf("second bucket = %+v, want 2026-05 W3 value 2", got.Buckets[1])
	}
}

func TestBuildAggregatesTokenMetricByYear(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2025-12-31", Model: "claude-sonnet-4.5", Tokens: domain.TokenSummary{StandardInputTokens: 1}},
		{Date: "2026-05-11", Model: "claude-sonnet-4.5", Tokens: domain.TokenSummary{StandardInputTokens: 2}},
	}

	got, err := Build(rollups, Options{GroupBy: "year", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Buckets) != 2 {
		t.Fatalf("buckets = %d, want 2 years", len(got.Buckets))
	}
	if got.Buckets[0].Label != "2025" || got.Buckets[0].Value.String() != "1" {
		t.Fatalf("first bucket = %+v, want 2025 value 1", got.Buckets[0])
	}
	if got.Buckets[1].Label != "2026" || got.Buckets[1].Value.String() != "2" {
		t.Fatalf("second bucket = %+v, want 2026 value 2", got.Buckets[1])
	}
}

func TestBuildEmptyDataReturnsEmptyChart(t *testing.T) {
	got, err := Build(nil, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Buckets) != 0 {
		t.Fatalf("buckets = %d, want 0", len(got.Buckets))
	}
}

func TestBuildRejectsInvalidDatesForTimeBuckets(t *testing.T) {
	tests := []struct {
		name    string
		groupBy string
		date    string
	}{
		{name: "day", groupBy: "day", date: "not-a-date"},
		{name: "month", groupBy: "month", date: "2026-13-99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollups := []cache.DailyRollup{
				{Date: tt.date, Model: "claude-sonnet-4.5", Tokens: domain.TokenSummary{StandardInputTokens: 1}},
			}

			_, err := Build(rollups, Options{GroupBy: tt.groupBy, Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
			if err == nil {
				t.Fatal("err = nil, want invalid date error")
			}
		})
	}
}

func TestBuildSplitsByModelWithTopNAndOther(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "model-a", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", Model: "model-b", Tokens: domain.TokenSummary{StandardInputTokens: 90}},
		{Date: "2026-05-10", Model: "model-c", Tokens: domain.TokenSummary{StandardInputTokens: 10}},
		{Date: "2026-05-11", Model: "model-a", Tokens: domain.TokenSummary{StandardInputTokens: 50}},
		{Date: "2026-05-11", Model: "model-c", Tokens: domain.TokenSummary{StandardInputTokens: 20}},
	}

	got, err := Build(rollups, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitModel, Top: 2, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Legend) != 3 {
		t.Fatalf("legend length = %d, want 3", len(got.Legend))
	}
	if got.Legend[0].Label != "model-a" || got.Legend[1].Label != "model-b" || got.Legend[2].Label != "Other" {
		t.Fatalf("legend = %+v, want model-a, model-b, Other", got.Legend)
	}
	if got.Buckets[0].Segments[0].Label != "model-a" || got.Buckets[0].Segments[0].Value.String() != "100" {
		t.Fatalf("first segment = %+v, want model-a 100", got.Buckets[0].Segments[0])
	}
	if got.Buckets[0].Segments[2].Label != "Other" || got.Buckets[0].Segments[2].Value.String() != "10" {
		t.Fatalf("other segment = %+v, want Other 10", got.Buckets[0].Segments[2])
	}
	for _, bucket := range got.Buckets {
		assertBucketValueEqualsSegments(t, bucket)
	}
}

func TestBuildSplitsEmptyDataReturnsEmptyChartAndLegend(t *testing.T) {
	got, err := Build(nil, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitModel, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Buckets) != 0 {
		t.Fatalf("buckets = %d, want 0", len(got.Buckets))
	}
	if len(got.Legend) != 0 {
		t.Fatalf("legend = %d, want 0", len(got.Legend))
	}
}

func TestBuildSplitTieOrdersByLabel(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "model-b", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", Model: "model-a", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", Model: "model-c", Tokens: domain.TokenSummary{StandardInputTokens: 50}},
	}

	got, err := Build(rollups, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitModel, Top: 2, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Legend) != 3 {
		t.Fatalf("legend length = %d, want 3", len(got.Legend))
	}
	if got.Legend[0].Label != "model-a" || got.Legend[1].Label != "model-b" || got.Legend[2].Label != "Other" {
		t.Fatalf("legend = %+v, want model-a, model-b, Other", got.Legend)
	}
}

func TestBuildSplitOmitsOtherWhenAllLabelsFitTop(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "model-a", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", Model: "model-b", Tokens: domain.TokenSummary{StandardInputTokens: 50}},
	}

	got, err := Build(rollups, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitModel, Top: 2, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Legend) != 2 {
		t.Fatalf("legend length = %d, want 2", len(got.Legend))
	}
	if got.Legend[0].Label != "model-a" || got.Legend[1].Label != "model-b" {
		t.Fatalf("legend = %+v, want model-a, model-b", got.Legend)
	}
	if len(got.Buckets[0].Segments) != 2 {
		t.Fatalf("segments length = %d, want 2", len(got.Buckets[0].Segments))
	}
}

func TestBuildSplitsByProjectName(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", ProjectID: "hash-a", ProjectName: "workspace_infra", Model: "glm-5.1", Tokens: domain.TokenSummary{StandardInputTokens: 100}},
		{Date: "2026-05-10", ProjectID: "hash-b", ProjectPathDisplay: "/Users/example/work/basic_dev", Model: "glm-5.1", Tokens: domain.TokenSummary{StandardInputTokens: 50}},
	}

	got, err := Build(rollups, Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitProject, Top: 8, Width: 40, Color: ColorNever}, pricing.BuiltinAnthropicProvider())
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Legend) != 2 {
		t.Fatalf("legend length = %d, want 2", len(got.Legend))
	}
	if got.Legend[0].Label != "workspace_infra" {
		t.Fatalf("first project label = %q, want workspace_infra", got.Legend[0].Label)
	}
	if got.Legend[1].Label != "/Users/example/work/basic_dev" {
		t.Fatalf("second project label = %q, want display path fallback", got.Legend[1].Label)
	}
}

func assertBucketValueEqualsSegments(t *testing.T, bucket Bucket) {
	t.Helper()

	sum := decimal.Zero
	for _, segment := range bucket.Segments {
		sum = sum.Add(segment.Value)
	}
	if !bucket.Value.Equal(sum) {
		t.Fatalf("bucket %q value = %s, want segment sum %s", bucket.Label, bucket.Value, sum)
	}
}
