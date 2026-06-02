package chart

import (
	"fmt"
	"sort"

	"github.com/breinzhang/tokusage/internal/cache"
	"github.com/breinzhang/tokusage/internal/pricing"
	"github.com/breinzhang/tokusage/internal/timebucket"
	"github.com/shopspring/decimal"
)

func Build(rollups []cache.DailyRollup, opts Options, provider pricing.Provider) (Chart, error) {
	if err := opts.Validate(); err != nil {
		return Chart{}, err
	}

	builder := chartBuilder{
		opts:     opts,
		provider: provider,
	}
	return builder.build(rollups)
}

type chartBuilder struct {
	opts     Options
	provider pricing.Provider
}

func (b chartBuilder) build(rollups []cache.DailyRollup) (Chart, error) {
	chart := Chart{
		Metric:  b.opts.Metric,
		GroupBy: b.opts.GroupBy,
		SplitBy: b.opts.SplitBy,
		From:    b.opts.From,
		To:      b.opts.To,
	}

	if b.opts.SplitBy == SplitNone {
		bucketValues := map[string]decimal.Decimal{}
		for _, rollup := range rollups {
			bucketLabel, err := timeBucketLabel(b.opts.GroupBy, rollup.Date)
			if err != nil {
				return Chart{}, err
			}
			bucketValues[bucketLabel] = bucketValues[bucketLabel].Add(b.rollupValue(rollup))
		}

		for _, label := range sortedKeys(bucketValues) {
			chart.Buckets = append(chart.Buckets, Bucket{
				Label: label,
				Value: bucketValues[label],
			})
		}
		return chart, nil
	}

	bucketSegments := map[string]map[string]decimal.Decimal{}
	labelTotals := map[string]decimal.Decimal{}
	for _, rollup := range rollups {
		bucketLabel, err := timeBucketLabel(b.opts.GroupBy, rollup.Date)
		if err != nil {
			return Chart{}, err
		}
		segmentLabel := b.segmentLabel(rollup)
		value := b.rollupValue(rollup)
		if bucketSegments[bucketLabel] == nil {
			bucketSegments[bucketLabel] = map[string]decimal.Decimal{}
		}
		bucketSegments[bucketLabel][segmentLabel] = bucketSegments[bucketLabel][segmentLabel].Add(value)
		labelTotals[segmentLabel] = labelTotals[segmentLabel].Add(value)
	}

	topLabels := topLabels(labelTotals, b.opts.Top)
	topSet := map[string]bool{}
	for _, label := range topLabels {
		topSet[label] = true
	}

	rangeTotal := decimal.Zero
	for _, value := range labelTotals {
		rangeTotal = rangeTotal.Add(value)
	}
	for i, label := range topLabels {
		chart.Legend = append(chart.Legend, LegendItem{
			Key:   fmt.Sprintf("%d", i+1),
			Label: label,
			Value: labelTotals[label],
			Share: percentOf(labelTotals[label], rangeTotal),
		})
	}

	otherTotal := decimal.Zero
	for label, value := range labelTotals {
		if !topSet[label] {
			otherTotal = otherTotal.Add(value)
		}
	}
	if !otherTotal.IsZero() {
		chart.Legend = append(chart.Legend, LegendItem{
			Key:   "+",
			Label: "Other",
			Value: otherTotal,
			Share: percentOf(otherTotal, rangeTotal),
		})
	}

	for _, bucketLabel := range sortedNestedKeys(bucketSegments) {
		bucket := Bucket{Label: bucketLabel}
		for _, label := range topLabels {
			value := bucketSegments[bucketLabel][label]
			bucket.Value = bucket.Value.Add(value)
			bucket.Segments = append(bucket.Segments, Segment{Label: label, Value: value})
		}

		other := decimal.Zero
		for label, value := range bucketSegments[bucketLabel] {
			if !topSet[label] {
				other = other.Add(value)
			}
		}
		if !otherTotal.IsZero() {
			bucket.Value = bucket.Value.Add(other)
			bucket.Segments = append(bucket.Segments, Segment{Label: "Other", Value: other})
		}
		chart.Buckets = append(chart.Buckets, bucket)
	}

	return chart, nil
}

func (b chartBuilder) segmentLabel(rollup cache.DailyRollup) string {
	switch b.opts.SplitBy {
	case SplitModel:
		return rollup.Model
	case SplitProject:
		if rollup.ProjectName != "" {
			return rollup.ProjectName
		}
		if rollup.ProjectPathDisplay != "" {
			return rollup.ProjectPathDisplay
		}
		return rollup.ProjectID
	default:
		return ""
	}
}

func (b chartBuilder) rollupValue(rollup cache.DailyRollup) decimal.Decimal {
	if b.opts.Metric == MetricCost {
		calc := pricing.Calculator{Provider: b.provider}
		return calc.Calculate(rollup.Model, rollup.Tokens).Total
	}
	return decimal.NewFromInt(rollup.Tokens.TotalTokens())
}

func timeBucketLabel(groupBy string, dateText string) (string, error) {
	return timebucket.Label(groupBy, dateText)
}

func sortedKeys(values map[string]decimal.Decimal) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func topLabels(totals map[string]decimal.Decimal, limit int) []string {
	labels := make([]string, 0, len(totals))
	for label := range totals {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		left := totals[labels[i]]
		right := totals[labels[j]]
		if !left.Equal(right) {
			return left.GreaterThan(right)
		}
		return labels[i] < labels[j]
	})
	if len(labels) > limit {
		return labels[:limit]
	}
	return labels
}

func sortedNestedKeys(values map[string]map[string]decimal.Decimal) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func percentOf(value decimal.Decimal, total decimal.Decimal) decimal.Decimal {
	if total.IsZero() {
		return decimal.Zero
	}
	return value.Mul(decimal.NewFromInt(100)).Div(total)
}
