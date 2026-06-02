package domain

import "testing"

func TestTokenSummaryTotal(t *testing.T) {
	tokens := TokenSummary{
		StandardInputTokens: 10,
		OutputTokens:        20,
		CacheWrite5mTokens:  30,
		CacheWrite1hTokens:  40,
		CacheReadTokens:     50,
	}

	if got := tokens.TotalTokens(); got != 150 {
		t.Fatalf("TotalTokens() = %d, want 150", got)
	}
}

func TestTokenSummaryAdd(t *testing.T) {
	left := TokenSummary{StandardInputTokens: 1, OutputTokens: 2}
	right := TokenSummary{CacheWrite5mTokens: 3, CacheWrite1hTokens: 4, CacheReadTokens: 5}

	got := left.Add(right)

	if got.StandardInputTokens != 1 || got.OutputTokens != 2 || got.CacheWrite5mTokens != 3 || got.CacheWrite1hTokens != 4 || got.CacheReadTokens != 5 {
		t.Fatalf("Add() = %+v", got)
	}
}
