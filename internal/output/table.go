package output

import (
	"fmt"
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

	minTokens := buckets[0].Tokens.TotalTokens()
	maxTokens := minTokens
	for _, bucket := range buckets[1:] {
		total := bucket.Tokens.TotalTokens()
		if total < minTokens {
			minTokens = total
		}
		if total > maxTokens {
			maxTokens = total
		}
	}
	if minTokens == maxTokens {
		return colors
	}

	for i, bucket := range buckets {
		switch total := bucket.Tokens.TotalTokens(); total {
		case minTokens:
			colors[i] = tableLessColor
		case maxTokens:
			colors[i] = tableMoreColor
		}
	}
	return colors
}

func colorizeTableRow(row string, color string) string {
	if color == "" {
		return row
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", color, row)
}
