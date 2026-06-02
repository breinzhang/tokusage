package aggregate

import (
	"testing"

	"github.com/breinzhang/tokusage/internal/cache"
	"github.com/breinzhang/tokusage/internal/domain"
)

func TestAggregateByModel(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "claude-sonnet-4.5", ProjectID: "p1", Tokens: domain.TokenSummary{StandardInputTokens: 10}},
		{Date: "2026-05-10", Model: "claude-sonnet-4.5", ProjectID: "p2", Tokens: domain.TokenSummary{OutputTokens: 5}},
		{Date: "2026-05-10", Model: "glm-5.1", ProjectID: "p1", Tokens: domain.TokenSummary{CacheReadTokens: 7}},
	}

	result := ByModel(rollups)

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result["claude-sonnet-4.5"].Tokens.TotalTokens() != 15 {
		t.Fatalf("sonnet total = %d", result["claude-sonnet-4.5"].Tokens.TotalTokens())
	}
	if result["glm-5.1"].Tokens.TotalTokens() != 7 {
		t.Fatalf("glm total = %d", result["glm-5.1"].Tokens.TotalTokens())
	}
}

func TestAggregateByDate(t *testing.T) {
	rollups := []cache.DailyRollup{
		{Date: "2026-05-10", Model: "a", Tokens: domain.TokenSummary{StandardInputTokens: 1}},
		{Date: "2026-05-10", Model: "b", Tokens: domain.TokenSummary{StandardInputTokens: 2}},
		{Date: "2026-05-11", Model: "a", Tokens: domain.TokenSummary{StandardInputTokens: 3}},
	}

	result := ByDate(rollups)

	if result["2026-05-10"].Tokens.TotalTokens() != 3 {
		t.Fatalf("2026-05-10 total = %d", result["2026-05-10"].Tokens.TotalTokens())
	}
	if result["2026-05-11"].Tokens.TotalTokens() != 3 {
		t.Fatalf("2026-05-11 total = %d", result["2026-05-11"].Tokens.TotalTokens())
	}
}
