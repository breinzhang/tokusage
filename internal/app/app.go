package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/breinzhang/tokusage/internal/cache"
	"github.com/breinzhang/tokusage/internal/chart"
	claudedata "github.com/breinzhang/tokusage/internal/datasource/claude"
	"github.com/breinzhang/tokusage/internal/domain"
	"github.com/breinzhang/tokusage/internal/output"
	"github.com/breinzhang/tokusage/internal/platform"
	"github.com/breinzhang/tokusage/internal/pricing"
	"github.com/breinzhang/tokusage/internal/timebucket"
	"github.com/shopspring/decimal"
)

type Options struct {
	Paths            []string
	CachePath        string
	From             string
	To               string
	Format           string
	GroupBy          string
	Timezone         string
	ExcludeSubagents bool
	ColorEnabled     bool
	RecentDataDays   int
	RecentDataWeeks  int
	PricingPath      string
}

type ChartOptions struct {
	Options
	Metric       string
	SplitBy      string
	Top          int
	Width        int
	ASCII        bool
	Color        string
	ColorEnabled bool
}

func RunClaudeReport(ctx context.Context, opts Options) (string, error) {
	rollups, err := loadClaudeRollups(ctx, opts)
	if err != nil {
		return "", err
	}
	rollups, opts = applyRecentDataWindow(rollups, opts)
	return renderRollupReport(opts, rollups, "Date", func(rollup cache.DailyRollup) string {
		return reportBucketLabel(opts, rollup)
	})
}

func RunClaudeModels(ctx context.Context, opts Options) (string, error) {
	rollups, err := loadClaudeRollups(ctx, opts)
	if err != nil {
		return "", err
	}
	rollups, opts = applyRecentDataWindow(rollups, opts)
	return renderRollupReport(opts, rollups, "Model", func(rollup cache.DailyRollup) string {
		return rollup.Model
	})
}

func RunClaudeProjects(ctx context.Context, opts Options) (string, error) {
	rollups, err := loadClaudeRollups(ctx, opts)
	if err != nil {
		return "", err
	}
	rollups, opts = applyRecentDataWindow(rollups, opts)
	return renderRollupReport(opts, rollups, "Project", func(rollup cache.DailyRollup) string {
		return projectBucketLabel(rollup)
	})
}

func RunClaudeChart(ctx context.Context, opts ChartOptions) (string, error) {
	chartOpts := chart.Options{
		GroupBy:      opts.GroupBy,
		Metric:       chart.Metric(opts.Metric),
		SplitBy:      chart.SplitBy(opts.SplitBy),
		Top:          opts.Top,
		Width:        opts.Width,
		ASCII:        opts.ASCII,
		Color:        chart.ColorMode(opts.Color),
		ColorEnabled: opts.ColorEnabled,
		From:         opts.From,
		To:           opts.To,
	}
	if err := chartOpts.Validate(); err != nil {
		return "", err
	}

	rollups, err := loadClaudeRollups(ctx, opts.Options)
	if err != nil {
		return "", err
	}
	rollups, opts.Options = applyRecentDataWindow(rollups, opts.Options)
	chartOpts.From = opts.From
	chartOpts.To = opts.To
	provider, err := loadPricingProvider(opts.Options)
	if err != nil {
		return "", err
	}
	view, err := chart.Build(rollups, chartOpts, provider)
	if err != nil {
		return "", err
	}
	return chart.Render(view, chartOpts), nil
}

type CacheStatus struct {
	Path        string
	FileCount   int64
	EventCount  int64
	RollupCount int64
}

func RunCacheStatus(ctx context.Context, cachePath string) (string, error) {
	db, err := cache.Open(cachePath)
	if err != nil {
		return "", err
	}
	defer db.Close()
	if err := cache.EnsureSchema(ctx, db); err != nil {
		return "", err
	}
	status, err := cache.Status(ctx, db, cachePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("DB: %s\nFiles: %d\nEvents: %d\nRollups: %d\n", status.Path, status.FileCount, status.EventCount, status.RollupCount), nil
}

func RunCacheClear(cachePath string) error {
	for _, path := range []string{cachePath, cachePath + "-shm", cachePath + "-wal"} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func RunCacheRebuild(ctx context.Context, opts Options) (string, error) {
	if err := RunCacheClear(opts.CachePath); err != nil {
		return "", err
	}
	if _, err := RunClaudeReport(ctx, opts); err != nil {
		return "", err
	}
	return "cache rebuilt\n", nil
}

func loadClaudeRollups(ctx context.Context, opts Options) ([]cache.DailyRollup, error) {
	db, err := cache.Open(opts.CachePath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := cache.EnsureSchema(ctx, db); err != nil {
		return nil, err
	}

	for _, root := range opts.Paths {
		files, err := claudedata.DiscoverJSONLFiles(root)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			events, _, err := claudedata.ParseJSONLFile(file)
			if err != nil {
				return nil, err
			}
			if err := cache.ReplaceFileEvents(ctx, db, "claude-code", platform.NormalizePathForStorage(file), events); err != nil {
				return nil, err
			}
		}
	}

	loc, err := time.LoadLocation(opts.Timezone)
	if err != nil {
		return nil, err
	}
	if err := cache.RebuildRollups(ctx, db, opts.Timezone, loc); err != nil {
		return nil, err
	}
	return cache.QueryDailyRollups(ctx, db, cache.RollupQuery{
		FromDate:         opts.From,
		ToDate:           opts.To,
		Timezone:         opts.Timezone,
		ExcludeSubagents: opts.ExcludeSubagents,
	})
}

func applyRecentDataWindow(rollups []cache.DailyRollup, opts Options) ([]cache.DailyRollup, Options) {
	if opts.From != "" || opts.To != "" || len(rollups) == 0 {
		return rollups, opts
	}
	if opts.RecentDataWeeks > 0 {
		return filterRecentWeekBuckets(rollups, opts)
	}
	if opts.RecentDataDays > 0 {
		return filterRecentDates(rollups, opts)
	}
	return rollups, opts
}

func filterRecentDates(rollups []cache.DailyRollup, opts Options) ([]cache.DailyRollup, Options) {
	dates := sortedUniqueDates(rollups)
	if len(dates) > opts.RecentDataDays {
		dates = dates[len(dates)-opts.RecentDataDays:]
	}
	keep := setFromStrings(dates)
	filtered := filterRollupsByDateSet(rollups, keep)
	return filtered, optionsWithDateRange(opts, filtered)
}

func filterRecentWeekBuckets(rollups []cache.DailyRollup, opts Options) ([]cache.DailyRollup, Options) {
	labels := make([]string, 0)
	seen := map[string]bool{}
	for _, date := range sortedUniqueDates(rollups) {
		label, err := timebucket.Label("week", date)
		if err != nil || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
	}
	if len(labels) > opts.RecentDataWeeks {
		labels = labels[len(labels)-opts.RecentDataWeeks:]
	}
	keep := setFromStrings(labels)
	filtered := make([]cache.DailyRollup, 0, len(rollups))
	for _, rollup := range rollups {
		label, err := timebucket.Label("week", rollup.Date)
		if err == nil && keep[label] {
			filtered = append(filtered, rollup)
		}
	}
	return filtered, optionsWithDateRange(opts, filtered)
}

func sortedUniqueDates(rollups []cache.DailyRollup) []string {
	seen := map[string]bool{}
	for _, rollup := range rollups {
		seen[rollup.Date] = true
	}
	dates := make([]string, 0, len(seen))
	for date := range seen {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	return dates
}

func setFromStrings(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func filterRollupsByDateSet(rollups []cache.DailyRollup, keep map[string]bool) []cache.DailyRollup {
	filtered := make([]cache.DailyRollup, 0, len(rollups))
	for _, rollup := range rollups {
		if keep[rollup.Date] {
			filtered = append(filtered, rollup)
		}
	}
	return filtered
}

func optionsWithDateRange(opts Options, rollups []cache.DailyRollup) Options {
	if len(rollups) == 0 {
		return opts
	}
	dates := sortedUniqueDates(rollups)
	opts.From = dates[0]
	opts.To = dates[len(dates)-1]
	return opts
}

func renderRollupReport(opts Options, rollups []cache.DailyRollup, labelHeader string, labelFor func(cache.DailyRollup) string) (string, error) {
	provider, err := loadPricingProvider(opts)
	if err != nil {
		return "", err
	}
	calc := pricing.Calculator{Provider: provider}
	report := output.Report{
		Agent:        "claude-code",
		Timezone:     opts.Timezone,
		From:         opts.From,
		To:           opts.To,
		LabelHeader:  labelHeader,
		Estimated:    true,
		PricingHash:  provider.Hash(),
		ColorEnabled: opts.ColorEnabled,
	}

	buckets := map[string]bucketTotals{}
	for _, rollup := range rollups {
		label := labelFor(rollup)
		current := buckets[label]
		current.tokens = current.tokens.Add(rollup.Tokens)
		cost := calc.Calculate(rollup.Model, rollup.Tokens)
		if cost.Partial {
			current.partial = true
		} else {
			current.cost = current.cost.Add(cost.Total)
		}
		buckets[label] = current
	}

	labels := make([]string, 0, len(buckets))
	for label := range buckets {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		current := buckets[label]
		report.Partial = report.Partial || current.partial
		report.Buckets = append(report.Buckets, output.Bucket{
			Label:         label,
			Tokens:        current.tokens,
			EstimatedCost: estimatedCost(current.cost, current.partial),
			Partial:       current.partial,
		})
	}

	if strings.EqualFold(opts.Format, "json") {
		data, err := output.RenderJSON(report)
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	}
	return output.RenderTable(report), nil
}

func loadPricingProvider(opts Options) (pricing.Provider, error) {
	builtin := pricing.BuiltinAnthropicProvider()
	if opts.PricingPath == "" {
		return builtin, nil
	}
	configPrices, err := pricing.LoadConfig(opts.PricingPath)
	if err != nil {
		return nil, fmt.Errorf("load pricing config: %w", err)
	}
	configured := pricing.NewStaticProvider(configPrices)
	return pricing.NewOverlayProvider(configured, builtin), nil
}

type bucketTotals struct {
	tokens  domain.TokenSummary
	cost    decimal.Decimal
	partial bool
}

func reportBucketLabel(opts Options, rollup cache.DailyRollup) string {
	switch strings.ToLower(opts.GroupBy) {
	case "range":
		if opts.From != "" || opts.To != "" {
			return opts.From + ".." + opts.To
		}
		return "range"
	case "month":
		if len(rollup.Date) >= len("2006-01") {
			return rollup.Date[:len("2006-01")]
		}
	case "week":
		label, err := timebucket.Label("week", rollup.Date)
		if err == nil {
			return label
		}
	case "year":
		label, err := timebucket.Label("year", rollup.Date)
		if err == nil {
			return label
		}
	}
	return rollup.Date
}

func projectBucketLabel(rollup cache.DailyRollup) string {
	if rollup.ProjectName != "" {
		return rollup.ProjectName
	}
	if rollup.ProjectPathDisplay != "" {
		return rollup.ProjectPathDisplay
	}
	return rollup.ProjectID
}

func estimatedCost(cost decimal.Decimal, partial bool) string {
	if partial && cost.IsZero() {
		return "N/A"
	}
	value := "$" + cost.StringFixed(2)
	if partial {
		return value + " partial"
	}
	return value
}
