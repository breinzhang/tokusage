package chart

import "testing"

func TestValidateOptionsAcceptsSupportedValues(t *testing.T) {
	for _, groupBy := range []string{"day", "week", "month", "year"} {
		opts := Options{
			GroupBy: groupBy,
			Metric:  MetricTokens,
			SplitBy: SplitModel,
			Top:     8,
			Width:   40,
			Color:   ColorNever,
		}

		if err := opts.Validate(); err != nil {
			t.Fatalf("Validate(%q) error = %v", groupBy, err)
		}
	}
}

func TestValidateOptionsRejectsUnsupportedMetric(t *testing.T) {
	opts := Options{GroupBy: "day", Metric: Metric("requests"), SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}

	err := opts.Validate()
	if err == nil || err.Error() != `unsupported metric "requests"` {
		t.Fatalf("Validate() error = %v, want unsupported metric", err)
	}
}

func TestValidateOptionsRejectsUnsupportedGroupBy(t *testing.T) {
	opts := Options{GroupBy: "range", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorNever}

	err := opts.Validate()
	if err == nil || err.Error() != `unsupported group-by "range"` {
		t.Fatalf("Validate() error = %v, want unsupported group-by", err)
	}
}

func TestValidateOptionsRejectsUnsupportedSplitBy(t *testing.T) {
	opts := Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitBy("agent"), Top: 8, Width: 40, Color: ColorNever}

	err := opts.Validate()
	if err == nil || err.Error() != `unsupported split-by "agent"` {
		t.Fatalf("Validate() error = %v, want unsupported split-by", err)
	}
}

func TestValidateOptionsRejectsSmallTopAndWidth(t *testing.T) {
	opts := Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitNone, Top: 0, Width: 40, Color: ColorNever}
	if err := opts.Validate(); err == nil || err.Error() != "--top must be at least 1" {
		t.Fatalf("top validation error = %v", err)
	}

	opts = Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 9, Color: ColorNever}
	if err := opts.Validate(); err == nil || err.Error() != "--width must be at least 10" {
		t.Fatalf("width validation error = %v", err)
	}
}

func TestValidateOptionsRejectsUnsupportedColor(t *testing.T) {
	opts := Options{GroupBy: "day", Metric: MetricTokens, SplitBy: SplitNone, Top: 8, Width: 40, Color: ColorMode("sometimes")}

	err := opts.Validate()
	if err == nil || err.Error() != `unsupported color mode "sometimes"` {
		t.Fatalf("Validate() error = %v, want unsupported color mode", err)
	}
}
