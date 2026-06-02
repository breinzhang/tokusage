package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/breinzhang/tokusage/internal/humanfmt"
	"github.com/shopspring/decimal"
)

var unicodeBarGlyphs = []string{"█", "▓", "▒", "░", "■", "▪"}
var asciiBarGlyphs = []string{"#", "=", "-", "+", "*", "."}
var barColors = []string{"38;5;208", "38;5;180", "38;5;109", "38;5;139", "38;5;73", "38;5;137", "38;5;174", "38;5;102"}

func Render(view Chart, opts Options) string {
	if len(view.Buckets) == 0 {
		return "No usage data for the selected range.\n"
	}

	var out strings.Builder
	out.WriteString(renderHeader(view, opts))
	out.WriteString("\n\n")

	maxValue := maxBucketValue(view.Buckets)
	bucketLabelWidth := maxBucketLabelWidth(view.Buckets)
	for i, bucket := range view.Buckets {
		if i > 0 {
			out.WriteString("\n")
		}
		width := scaledWidth(bucket.Value, maxValue, opts.Width)
		if view.SplitBy == SplitNone {
			fmt.Fprintf(&out, "%-*s  %8s  %s\n", bucketLabelWidth, bucket.Label, formatValue(view.Metric, bucket.Value), renderBar(width, barGlyphs(opts)[0], 0, opts))
			continue
		}
		fmt.Fprintf(&out, "%-*s  %8s  %s\n", bucketLabelWidth, bucket.Label, formatValue(view.Metric, bucket.Value), renderStackedBar(bucket, width, opts))
	}

	if view.SplitBy != SplitNone && len(view.Legend) > 0 {
		out.WriteString("\nLegend\n")
		legendLabelWidth := maxLegendLabelWidth(view.Legend)
		for i, item := range view.Legend {
			label := fmt.Sprintf("%s  %-*s", item.Key, legendLabelWidth, item.Label)
			fmt.Fprintf(&out, "%s %8s  %s\n", colorize(label, i, opts), formatValue(view.Metric, item.Value), humanfmt.Percent(item.Share))
		}
	}

	return out.String()
}

func renderHeader(view Chart, opts Options) string {
	if view.SplitBy == SplitNone {
		return fmt.Sprintf("Metric: %s | Group: %s | %s", view.Metric, view.GroupBy, dateRangeLabel(view.From, view.To))
	}
	return fmt.Sprintf("Metric: %s | Group: %s | Split: %s | Top: %d", view.Metric, view.GroupBy, view.SplitBy, opts.Top)
}

func dateRangeLabel(from string, to string) string {
	if from == "" && to == "" {
		return "all dates"
	}
	return from + ".." + to
}

func maxBucketValue(buckets []Bucket) decimal.Decimal {
	maxValue := decimal.Zero
	for _, bucket := range buckets {
		if bucket.Value.GreaterThan(maxValue) {
			maxValue = bucket.Value
		}
	}
	return maxValue
}

func maxBucketLabelWidth(buckets []Bucket) int {
	width := 0
	for _, bucket := range buckets {
		width = max(width, stringWidth(bucket.Label))
	}
	return width
}

func maxLegendLabelWidth(items []LegendItem) int {
	width := 0
	for _, item := range items {
		width = max(width, stringWidth(item.Label))
	}
	return width
}

func stringWidth(value string) int {
	return len([]rune(value))
}

func scaledWidth(value decimal.Decimal, maxValue decimal.Decimal, width int) int {
	if value.IsZero() || maxValue.IsZero() || width <= 0 {
		return 0
	}
	scaled, _ := value.Div(maxValue).Mul(decimal.NewFromInt(int64(width))).Float64()
	cells := int(math.Round(scaled))
	if cells < 1 {
		return 1
	}
	if cells > width {
		return width
	}
	return cells
}

func renderStackedBar(bucket Bucket, width int, opts Options) string {
	if width == 0 {
		return ""
	}
	glyphs := barGlyphs(opts)
	cells := segmentWidths(bucket.Segments, bucket.Value, width)

	var bar strings.Builder
	for i, cellCount := range cells {
		if cellCount == 0 {
			continue
		}
		glyph := splitBarGlyph(glyphs, i, opts)
		bar.WriteString(renderBar(cellCount, glyph, i, opts))
	}
	return bar.String()
}

func splitBarGlyph(glyphs []string, index int, opts Options) string {
	if shouldColor(opts) && !opts.ASCII {
		return unicodeBarGlyphs[0]
	}
	return glyphs[index%len(glyphs)]
}

func segmentWidths(segments []Segment, total decimal.Decimal, width int) []int {
	cells := make([]int, len(segments))
	if len(segments) == 0 || total.IsZero() || width <= 0 {
		return cells
	}

	nonZero := nonZeroSegmentIndexes(segments)
	if len(nonZero) > width {
		for _, i := range nonZero[:width] {
			cells[i] = 1
		}
		return cells
	}

	used := 0
	for i, segment := range segments {
		if segment.Value.IsZero() {
			continue
		}
		scaled, _ := segment.Value.Div(total).Mul(decimal.NewFromInt(int64(width))).Float64()
		cellCount := int(math.Round(scaled))
		if cellCount < 1 {
			cellCount = 1
		}
		cells[i] = cellCount
		used += cellCount
	}

	for used > width {
		i := widestSegment(cells)
		if i < 0 || cells[i] <= 1 {
			break
		}
		cells[i]--
		used--
	}
	for used < width {
		i := widestValueSegment(segments, cells)
		if i < 0 {
			break
		}
		cells[i]++
		used++
	}

	return cells
}

func nonZeroSegmentIndexes(segments []Segment) []int {
	indexes := make([]int, 0, len(segments))
	for i, segment := range segments {
		if !segment.Value.IsZero() {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func widestSegment(cells []int) int {
	index := -1
	for i, cellCount := range cells {
		if index == -1 || cellCount > cells[index] {
			index = i
		}
	}
	return index
}

func widestValueSegment(segments []Segment, cells []int) int {
	index := -1
	for i, segment := range segments {
		if cells[i] == 0 {
			continue
		}
		if index == -1 || segment.Value.GreaterThan(segments[index].Value) {
			index = i
		}
	}
	return index
}

func renderBar(width int, glyph string, colorIndex int, opts Options) string {
	bar := strings.Repeat(glyph, width)
	return colorize(bar, colorIndex, opts)
}

func colorize(text string, colorIndex int, opts Options) string {
	if !shouldColor(opts) || text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", barColors[colorIndex%len(barColors)], text)
}

func shouldColor(opts Options) bool {
	if opts.Color == ColorNever {
		return false
	}
	return opts.Color == ColorAlways || opts.ColorEnabled
}

func barGlyphs(opts Options) []string {
	if opts.ASCII {
		return asciiBarGlyphs
	}
	return unicodeBarGlyphs
}

func formatValue(metric Metric, value decimal.Decimal) string {
	if metric == MetricCost {
		return humanfmt.Cost(value)
	}
	return humanfmt.Tokens(value.IntPart())
}
