package chart

import "fmt"

func (o Options) Validate() error {
	switch o.GroupBy {
	case "day", "week", "month", "year":
	default:
		return fmt.Errorf("unsupported group-by %q", o.GroupBy)
	}

	switch o.Metric {
	case MetricTokens, MetricCost:
	default:
		return fmt.Errorf("unsupported metric %q", o.Metric)
	}

	switch o.SplitBy {
	case SplitNone, SplitModel, SplitProject:
	default:
		return fmt.Errorf("unsupported split-by %q", o.SplitBy)
	}

	if o.Top < 1 {
		return fmt.Errorf("--top must be at least 1")
	}
	if o.Width < 10 {
		return fmt.Errorf("--width must be at least 10")
	}

	switch o.Color {
	case ColorAuto, ColorNever, ColorAlways:
	default:
		return fmt.Errorf("unsupported color mode %q", o.Color)
	}

	return nil
}
