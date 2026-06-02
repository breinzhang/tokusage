package pricing

import (
	"github.com/breinzhang/tokusage/internal/domain"
	"github.com/shopspring/decimal"
)

type CostSummary struct {
	StandardInput decimal.Decimal `json:"standard_input"`
	CacheWrite5m  decimal.Decimal `json:"cache_write_5m"`
	CacheWrite1h  decimal.Decimal `json:"cache_write_1h"`
	CacheRead     decimal.Decimal `json:"cache_read"`
	Output        decimal.Decimal `json:"output"`
	Total         decimal.Decimal `json:"total"`
	Partial       bool            `json:"partial"`
}

type Calculator struct {
	Provider Provider
}

func (c Calculator) Calculate(model string, tokens domain.TokenSummary) CostSummary {
	price, ok := c.Provider.PriceFor(model)
	if !ok {
		return CostSummary{Partial: true}
	}
	million := decimal.NewFromInt(1_000_000)
	cost := CostSummary{
		StandardInput: decimal.NewFromInt(tokens.StandardInputTokens).Mul(price.StandardInputPerMTok).Div(million),
		CacheWrite5m:  decimal.NewFromInt(tokens.CacheWrite5mTokens).Mul(price.CacheWrite5mPerMTok).Div(million),
		CacheWrite1h:  decimal.NewFromInt(tokens.CacheWrite1hTokens).Mul(price.CacheWrite1hPerMTok).Div(million),
		CacheRead:     decimal.NewFromInt(tokens.CacheReadTokens).Mul(price.CacheReadPerMTok).Div(million),
		Output:        decimal.NewFromInt(tokens.OutputTokens).Mul(price.OutputPerMTok).Div(million),
	}
	cost.Total = cost.StandardInput.Add(cost.CacheWrite5m).Add(cost.CacheWrite1h).Add(cost.CacheRead).Add(cost.Output)
	return cost
}

func mustDecimal(value string) decimal.Decimal {
	result, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return result
}
