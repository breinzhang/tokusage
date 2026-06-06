package output

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/breinzhang/tokusage/internal/humanfmt"
)

const (
	tableLessColor = "38;5;95"
	tableMoreColor = "38;5;208"
)

func RenderTable(report Report) string {
	var builder strings.Builder
	labelHeader := report.LabelHeader
	if labelHeader == "" {
		labelHeader = "Date"
	}
	labelWidth := labelColumnWidth(labelHeader, report.Buckets)
	builder.WriteString(fmt.Sprintf("%-*s %12s %12s %12s %12s %12s %14s\n", labelWidth, labelHeader, "Input", "Cache Write", "Cache Read", "Output", "Total", "Est. Cost"))
	rowColors := tableRowColors(report.Buckets, report.ColorEnabled)
	for i, bucket := range report.Buckets {
		cacheWrite := bucket.Tokens.CacheWrite5mTokens + bucket.Tokens.CacheWrite1hTokens
		row := fmt.Sprintf(
			"%-*s %12s %12s %12s %12s %12s %14s",
			labelWidth,
			bucket.Label,
			humanfmt.Tokens(bucket.Tokens.StandardInputTokens),
			humanfmt.Tokens(cacheWrite),
			humanfmt.Tokens(bucket.Tokens.CacheReadTokens),
			humanfmt.Tokens(bucket.Tokens.OutputTokens),
			humanfmt.Tokens(bucket.Tokens.TotalTokens()),
			bucket.EstimatedCost,
		)
		builder.WriteString(colorizeTableRow(row, rowColors[i]))
		builder.WriteString("\n")
	}
	return builder.String()
}

func labelColumnWidth(labelHeader string, buckets []Bucket) int {
	width := len(labelHeader)
	if width < 12 {
		width = 12
	}
	for _, bucket := range buckets {
		if len(bucket.Label) > width {
			width = len(bucket.Label)
		}
	}
	return width
}

func tableRowColors(buckets []Bucket, enabled bool) []string {
	colors := make([]string, len(buckets))
	if !enabled || len(buckets) < 2 {
		return colors
	}

	costs := make([]float64, len(buckets))
	comparable := make([]bool, len(buckets))
	comparableCount := 0
	var minCost float64
	var maxCost float64
	for i, bucket := range buckets {
		cost, ok := tableCostValue(bucket.EstimatedCost)
		if !ok {
			continue
		}
		costs[i] = cost
		comparable[i] = true
		if comparableCount == 0 {
			minCost = cost
			maxCost = cost
		}
		if cost < minCost {
			minCost = cost
		}
		if cost > maxCost {
			maxCost = cost
		}
		comparableCount++
	}
	if comparableCount < 2 || minCost == maxCost {
		return colors
	}

	for i, cost := range costs {
		if !comparable[i] {
			continue
		}
		switch cost {
		case minCost:
			colors[i] = tableLessColor
		case maxCost:
			colors[i] = tableMoreColor
		}
	}
	return colors
}

func tableCostValue(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "$") {
		return 0, false
	}
	fields := strings.Fields(strings.TrimPrefix(value, "$"))
	if len(fields) == 0 {
		return 0, false
	}
	cost, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}
	return cost, true
}

func colorizeTableRow(row string, color string) string {
	if color == "" {
		return row
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", color, row)
}
