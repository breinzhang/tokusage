package domain

type TokenSummary struct {
	StandardInputTokens int64 `json:"standard_input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheWrite5mTokens  int64 `json:"cache_write_5m_tokens"`
	CacheWrite1hTokens  int64 `json:"cache_write_1h_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
}

func (t TokenSummary) TotalTokens() int64 {
	return t.StandardInputTokens + t.OutputTokens + t.CacheWrite5mTokens + t.CacheWrite1hTokens + t.CacheReadTokens
}

func (t TokenSummary) Add(other TokenSummary) TokenSummary {
	return TokenSummary{
		StandardInputTokens: t.StandardInputTokens + other.StandardInputTokens,
		OutputTokens:        t.OutputTokens + other.OutputTokens,
		CacheWrite5mTokens:  t.CacheWrite5mTokens + other.CacheWrite5mTokens,
		CacheWrite1hTokens:  t.CacheWrite1hTokens + other.CacheWrite1hTokens,
		CacheReadTokens:     t.CacheReadTokens + other.CacheReadTokens,
	}
}
