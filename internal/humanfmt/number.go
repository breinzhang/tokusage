package humanfmt

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func Tokens(value int64) string {
	if value < 0 {
		return "-" + Tokens(-value)
	}
	units := []struct {
		threshold int64
		suffix    string
	}{
		{threshold: 1_000_000_000, suffix: "B"},
		{threshold: 1_000_000, suffix: "M"},
		{threshold: 1_000, suffix: "K"},
	}
	for _, unit := range units {
		if value >= unit.threshold {
			scaled := float64(value) / float64(unit.threshold)
			text := strings.TrimSuffix(fmt.Sprintf("%.1f", scaled), ".0")
			return text + unit.suffix
		}
	}
	return fmt.Sprintf("%d", value)
}

func Cost(value decimal.Decimal) string {
	return "$" + value.StringFixed(2)
}

func Percent(value decimal.Decimal) string {
	return value.StringFixed(1) + "%"
}
