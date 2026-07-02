package chart

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"github.com/breinzhang/tokusage/internal/cache"
)

const (
	heatmapLabelWidth     = 4
	heatmapCellGlyph      = "■"
	heatmapEmptyCellGlyph = "□"
	heatmapCellSeparator  = "\u2009"
)

var heatmapColors = []string{
	"38;2;50;52;58",
	"38;2;119;65;44",
	"38;2;162;84;49",
	"38;2;203;106;57",
	"38;2;232;126;67",
}

type HeatmapOptions struct {
	From         string
	To           string
	Color        ColorMode
	ColorEnabled bool
}

type Heatmap struct {
	From            string
	To              string
	DefaultLastYear bool
	ActiveDays      int
	TotalTokens     int64
	MaxValue        int64
	Days            []HeatmapDay
}

type HeatmapDay struct {
	Date  string
	Value int64
}

func BuildHeatmap(rollups []cache.DailyRollup, opts HeatmapOptions) (Heatmap, error) {
	if err := opts.Validate(); err != nil {
		return Heatmap{}, err
	}

	values, latest, err := heatmapDailyValues(rollups)
	if err != nil {
		return Heatmap{}, err
	}
	start, end, ok, err := heatmapRange(latest, opts)
	if err != nil {
		return Heatmap{}, err
	}
	if !ok {
		return Heatmap{}, nil
	}

	view := Heatmap{
		From:            start.Format(time.DateOnly),
		To:              end.Format(time.DateOnly),
		DefaultLastYear: opts.From == "" && opts.To == "",
	}
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateText := date.Format(time.DateOnly)
		value := values[dateText]
		if value > 0 {
			view.ActiveDays++
			view.TotalTokens += value
			if value > view.MaxValue {
				view.MaxValue = value
			}
		}
		view.Days = append(view.Days, HeatmapDay{Date: dateText, Value: value})
	}
	return view, nil
}

func (o HeatmapOptions) Validate() error {
	switch o.Color {
	case "", ColorAuto, ColorNever, ColorAlways:
	default:
		return fmt.Errorf("unsupported color mode %q", o.Color)
	}
	if o.From != "" {
		if _, err := parseHeatmapDate(o.From); err != nil {
			return err
		}
	}
	if o.To != "" {
		if _, err := parseHeatmapDate(o.To); err != nil {
			return err
		}
	}
	return nil
}

func RenderHeatmap(view Heatmap, opts HeatmapOptions) string {
	if len(view.Days) == 0 {
		return "No usage data for the selected range.\n"
	}

	var out strings.Builder
	out.WriteString(heatmapTitle(view))
	out.WriteString("\n\n")
	out.WriteString(renderHeatmapMonthLabels(view))
	out.WriteString("\n")

	values := heatmapValueByDate(view)
	weeks := heatmapWeekStarts(view)
	for weekday := time.Sunday; weekday <= time.Saturday; weekday++ {
		fmt.Fprintf(&out, "%-3s ", heatmapWeekdayLabel(weekday))
		for i, week := range weeks {
			date := week.AddDate(0, 0, int(weekday))
			level := heatmapLevel(values[date.Format(time.DateOnly)], view.MaxValue)
			out.WriteString(renderHeatmapCell(level, opts))
			if i < len(weeks)-1 {
				out.WriteString(heatmapCellSeparator)
			}
		}
		out.WriteString("\n")
	}

	out.WriteString("\n")
	out.WriteString(renderHeatmapLegend(opts))
	out.WriteString("\n")
	return out.String()
}

func heatmapDailyValues(rollups []cache.DailyRollup) (map[string]int64, time.Time, error) {
	values := map[string]int64{}
	var latest time.Time
	for _, rollup := range rollups {
		date, err := parseHeatmapDate(rollup.Date)
		if err != nil {
			return nil, time.Time{}, err
		}
		if latest.IsZero() || date.After(latest) {
			latest = date
		}
		values[rollup.Date] += rollup.Tokens.TotalTokens()
	}
	return values, latest, nil
}

func heatmapRange(latest time.Time, opts HeatmapOptions) (time.Time, time.Time, bool, error) {
	var start time.Time
	var end time.Time
	var err error

	if opts.From != "" {
		start, err = parseHeatmapDate(opts.From)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}
	}
	if opts.To != "" {
		end, err = parseHeatmapDate(opts.To)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}
	}
	if end.IsZero() && !latest.IsZero() {
		end = latest
	}
	if start.IsZero() && !end.IsZero() {
		start = end.AddDate(0, 0, -364)
	}
	if end.IsZero() && !start.IsZero() {
		end = start.AddDate(0, 0, 364)
	}
	if start.IsZero() || end.IsZero() {
		return time.Time{}, time.Time{}, false, nil
	}
	if start.After(end) {
		return time.Time{}, time.Time{}, false, fmt.Errorf("--from must be on or before --to")
	}
	return start, end, true, nil
}

func parseHeatmapDate(value string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", value, err)
	}
	return date, nil
}

func heatmapTitle(view Heatmap) string {
	if view.DefaultLastYear {
		return fmt.Sprintf("%d active days in the last year", view.ActiveDays)
	}
	return fmt.Sprintf("%d active days from %s to %s", view.ActiveDays, view.From, view.To)
}

func renderHeatmapMonthLabels(view Heatmap) string {
	weeks := heatmapWeekStarts(view)
	labels := heatmapMonthLabels(weeks, view)
	line := []rune(strings.Repeat(" ", heatmapLabelWidth))
	for i := range weeks {
		line = append(line, ' ')
		if i < len(weeks)-1 {
			line = append(line, []rune(heatmapCellSeparator)...)
		}
	}
	for i, label := range labels {
		if label == "" {
			continue
		}
		offset := heatmapMonthLabelOffset(i)
		for j, char := range label {
			for len(line) <= offset+j {
				line = append(line, ' ')
			}
			line[offset+j] = char
		}
	}
	return strings.TrimRightFunc(string(line), unicode.IsSpace)
}

func heatmapMonthLabelOffset(weekIndex int) int {
	return heatmapLabelWidth + weekIndex*(len([]rune(heatmapCellGlyph))+len([]rune(heatmapCellSeparator)))
}

func heatmapMonthLabels(weeks []time.Time, view Heatmap) []string {
	labels := make([]string, len(weeks))
	from, err := parseHeatmapDate(view.From)
	if err != nil {
		return labels
	}
	to, err := parseHeatmapDate(view.To)
	if err != nil {
		return labels
	}

	for i, week := range weeks {
		for offset := 0; offset < 7; offset++ {
			date := week.AddDate(0, 0, offset)
			if date.Before(from) || date.After(to) {
				continue
			}
			if date.Day() == 1 || date.Equal(from) {
				labels[i] = date.Format("Jan")
				break
			}
		}
	}
	suppressOverlappingMonthLabels(labels)
	return labels
}

func suppressOverlappingMonthLabels(labels []string) {
	lastIndex := -1
	lastEnd := -1
	for i, label := range labels {
		if label == "" {
			continue
		}
		offset := heatmapMonthLabelOffset(i)
		if lastIndex >= 0 && offset < lastEnd {
			labels[lastIndex] = ""
		}
		lastIndex = i
		lastEnd = offset + len([]rune(label))
	}
}

func heatmapWeekStarts(view Heatmap) []time.Time {
	from, err := parseHeatmapDate(view.From)
	if err != nil {
		return nil
	}
	to, err := parseHeatmapDate(view.To)
	if err != nil {
		return nil
	}
	start := from.AddDate(0, 0, -int(from.Weekday()))
	end := to.AddDate(0, 0, int(time.Saturday-to.Weekday()))

	var weeks []time.Time
	for week := start; !week.After(end); week = week.AddDate(0, 0, 7) {
		weeks = append(weeks, week)
	}
	return weeks
}

func heatmapWeekdayLabel(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return "Mon"
	case time.Wednesday:
		return "Wed"
	case time.Friday:
		return "Fri"
	default:
		return ""
	}
}

func heatmapValueByDate(view Heatmap) map[string]int64 {
	values := make(map[string]int64, len(view.Days))
	for _, day := range view.Days {
		values[day.Date] = day.Value
	}
	return values
}

func heatmapLevel(value int64, maxValue int64) int {
	if value <= 0 || maxValue <= 0 {
		return 0
	}
	level := int(math.Ceil(float64(value) / float64(maxValue) * 4))
	if level < 1 {
		return 1
	}
	if level > 4 {
		return 4
	}
	return level
}

func renderHeatmapLegend(opts HeatmapOptions) string {
	var out strings.Builder
	out.WriteString("Less ")
	for level := 0; level <= 4; level++ {
		out.WriteString(renderHeatmapCell(level, opts))
		if level < 4 {
			out.WriteString(heatmapCellSeparator)
		}
	}
	out.WriteString(" More")
	return out.String()
}

func renderHeatmapCell(level int, opts HeatmapOptions) string {
	if level < 0 {
		level = 0
	}
	if level >= len(heatmapColors) {
		level = len(heatmapColors) - 1
	}
	glyph := heatmapCellGlyph
	if !shouldHeatmapColor(opts) && level == 0 {
		glyph = heatmapEmptyCellGlyph
	}
	if !shouldHeatmapColor(opts) {
		return glyph
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", heatmapColors[level], glyph)
}

func shouldHeatmapColor(opts HeatmapOptions) bool {
	if opts.Color == ColorNever {
		return false
	}
	return opts.Color == ColorAlways || opts.ColorEnabled
}
