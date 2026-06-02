package humanfmt

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestTokensFormatsUnits(t *testing.T) {
	tests := map[int64]string{
		0:             "0",
		999:           "999",
		1_000:         "1K",
		1_234:         "1.2K",
		1_234_567:     "1.2M",
		1_234_567_890: "1.2B",
	}

	for input, want := range tests {
		if got := Tokens(input); got != want {
			t.Fatalf("Tokens(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestTokensFormatsNegativeValues(t *testing.T) {
	if got := Tokens(-1_234); got != "-1.2K" {
		t.Fatalf("Tokens(-1234) = %q, want -1.2K", got)
	}
}

func TestCostFormatsUSD(t *testing.T) {
	value := decimal.RequireFromString("3.2")

	if got := Cost(value); got != "$3.20" {
		t.Fatalf("Cost(3.2) = %q, want $3.20", got)
	}
}

func TestPercentFormatsOneDecimal(t *testing.T) {
	value := decimal.RequireFromString("12.345")

	if got := Percent(value); got != "12.3%" {
		t.Fatalf("Percent(12.345) = %q, want 12.3%%", got)
	}
}
