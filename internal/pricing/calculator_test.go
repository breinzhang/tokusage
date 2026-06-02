package pricing

import (
	"testing"

	"github.com/breinzhang/tokusage/internal/domain"
)

func TestCalculateKnownModelCost(t *testing.T) {
	provider := NewStaticProvider([]ModelPrice{{
		Pattern:              "claude-sonnet-4*",
		StandardInputPerMTok: mustDecimal("3.00"),
		CacheWrite5mPerMTok:  mustDecimal("3.75"),
		CacheWrite1hPerMTok:  mustDecimal("6.00"),
		CacheReadPerMTok:     mustDecimal("0.30"),
		OutputPerMTok:        mustDecimal("15.00"),
	}})
	calc := Calculator{Provider: provider}

	cost := calc.Calculate("claude-sonnet-4.5", domain.TokenSummary{
		StandardInputTokens: 1_000_000,
		CacheWrite5mTokens:  1_000_000,
		CacheWrite1hTokens:  1_000_000,
		CacheReadTokens:     1_000_000,
		OutputTokens:        1_000_000,
	})

	if cost.Partial {
		t.Fatal("Partial = true, want false")
	}
	if cost.Total.StringFixed(2) != "28.05" {
		t.Fatalf("Total = %s, want 28.05", cost.Total.StringFixed(2))
	}
}

func TestCalculateUnknownModelIsPartial(t *testing.T) {
	calc := Calculator{Provider: NewStaticProvider(nil)}

	cost := calc.Calculate("glm-5.1", domain.TokenSummary{StandardInputTokens: 1})

	if !cost.Partial {
		t.Fatal("Partial = false, want true")
	}
	if cost.Total.String() != "0" {
		t.Fatalf("Total = %s, want 0", cost.Total.String())
	}
}

func TestBuiltinAnthropicProviderUsesCurrentClaudePricing(t *testing.T) {
	provider := BuiltinAnthropicProvider()

	opus, ok := provider.PriceFor("claude-opus-4.8")
	if !ok {
		t.Fatal("claude-opus-4.8 price not found")
	}
	if opus.StandardInputPerMTok.StringFixed(2) != "5.00" || opus.OutputPerMTok.StringFixed(2) != "25.00" {
		t.Fatalf("opus 4.8 price = input %s output %s", opus.StandardInputPerMTok, opus.OutputPerMTok)
	}

	sonnet, ok := provider.PriceFor("claude-sonnet-4.6")
	if !ok {
		t.Fatal("claude-sonnet-4.6 price not found")
	}
	if sonnet.StandardInputPerMTok.StringFixed(2) != "3.00" || sonnet.CacheReadPerMTok.StringFixed(2) != "0.30" || sonnet.OutputPerMTok.StringFixed(2) != "15.00" {
		t.Fatalf("sonnet 4.6 price = input %s read %s output %s", sonnet.StandardInputPerMTok, sonnet.CacheReadPerMTok, sonnet.OutputPerMTok)
	}
}

func TestBuiltinAnthropicProviderUsesSonnetPricingForUnknownModels(t *testing.T) {
	provider := BuiltinAnthropicProvider()

	price, ok := provider.PriceFor("glm-5.1")
	if !ok {
		t.Fatal("glm-5.1 fallback price not found")
	}
	if price.StandardInputPerMTok.StringFixed(2) != "3.00" || price.OutputPerMTok.StringFixed(2) != "15.00" {
		t.Fatalf("fallback price = input %s output %s, want Sonnet pricing", price.StandardInputPerMTok, price.OutputPerMTok)
	}
}

func TestBuiltinAnthropicProviderMatchesHyphenatedModelIDs(t *testing.T) {
	provider := BuiltinAnthropicProvider()

	opus, ok := provider.PriceFor("claude-opus-4-5-20251101")
	if !ok {
		t.Fatal("claude-opus-4-5-20251101 price not found")
	}
	if opus.StandardInputPerMTok.StringFixed(2) != "5.00" || opus.OutputPerMTok.StringFixed(2) != "25.00" {
		t.Fatalf("opus 4-5 price = input %s output %s", opus.StandardInputPerMTok, opus.OutputPerMTok)
	}

	sonnet, ok := provider.PriceFor("claude-sonnet-4-5-20250929")
	if !ok {
		t.Fatal("claude-sonnet-4-5-20250929 price not found")
	}
	if sonnet.StandardInputPerMTok.StringFixed(2) != "3.00" || sonnet.OutputPerMTok.StringFixed(2) != "15.00" {
		t.Fatalf("sonnet 4-5 price = input %s output %s", sonnet.StandardInputPerMTok, sonnet.OutputPerMTok)
	}
}
