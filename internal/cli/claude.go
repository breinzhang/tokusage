package cli

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/breinzhang/tokusage/internal/app"
	"github.com/breinzhang/tokusage/internal/platform"
	"github.com/breinzhang/tokusage/internal/pricing"
	"github.com/spf13/cobra"
)

func newClaudeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Analyze Claude Code usage",
	}
	cmd.AddCommand(newClaudeReportCommand())
	cmd.AddCommand(newClaudeModelsCommand())
	cmd.AddCommand(newClaudeProjectsCommand())
	cmd.AddCommand(newClaudeChartCommand())
	cmd.AddCommand(newClaudeHeatmapCommand())
	return cmd
}

func newClaudeReportCommand() *cobra.Command {
	return newClaudeReportLikeCommand("report", "Show Claude Code usage over time", "day", app.RunClaudeReport)
}

func newClaudeModelsCommand() *cobra.Command {
	return newClaudeReportLikeCommand("models", "Show Claude Code usage by model", "range", app.RunClaudeModels)
}

func newClaudeProjectsCommand() *cobra.Command {
	return newClaudeReportLikeCommand("projects", "Show Claude Code usage by project", "range", app.RunClaudeProjects)
}

func newClaudeChartCommand() *cobra.Command {
	var from, to, timezoneName, cachePath, pricingPath string
	var groupBy, metric, splitBy, colorMode string
	var paths []string
	var excludeSubagents, ascii bool
	var top, width int

	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Show Claude Code usage chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			common, err := claudeOptionsWithPricingAndGroupByFlag(paths, cachePath, pricingPath, from, to, timezoneName, "table", groupBy, cmd.Flags().Changed("group-by"), excludeSubagents)
			if err != nil {
				return err
			}
			if common, err = resolvePricingConfig(common, pricingPath); err != nil {
				return err
			}
			out, err := app.RunClaudeChart(cmd.Context(), app.ChartOptions{
				Options:      common,
				Metric:       metric,
				SplitBy:      splitBy,
				Top:          top,
				Width:        width,
				ASCII:        ascii,
				Color:        colorMode,
				ColorEnabled: resolveColorEnabled(colorMode, cmd.OutOrStdout()),
			})
			if err != nil {
				return err
			}
			cmd.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start date in YYYY-MM-DD")
	cmd.Flags().StringVar(&to, "to", "", "end date in YYYY-MM-DD")
	cmd.Flags().StringVar(&timezoneName, "timezone", "", "IANA timezone")
	cmd.Flags().StringVar(&groupBy, "group-by", "day", "grouping: day, week, month, year")
	cmd.Flags().StringVar(&metric, "metric", "tokens", "chart metric: tokens or cost")
	cmd.Flags().StringVar(&splitBy, "split-by", "model", "split chart by: none, model, project")
	cmd.Flags().IntVar(&top, "top", 8, "number of split labels to show before Other")
	cmd.Flags().IntVar(&width, "width", 40, "maximum chart bar width")
	cmd.Flags().BoolVar(&ascii, "ascii", false, "use ASCII bars")
	cmd.Flags().StringVar(&colorMode, "color", "auto", "color mode: auto, never, always")
	cmd.Flags().StringArrayVar(&paths, "path", nil, "Claude projects directory")
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	cmd.Flags().StringVar(&pricingPath, "pricing", "", "pricing TOML path")
	cmd.Flags().BoolVar(&excludeSubagents, "exclude-subagents", false, "exclude subagent usage")
	return cmd
}

func newClaudeHeatmapCommand() *cobra.Command {
	var from, to, timezoneName, cachePath, colorMode string
	var paths []string
	var excludeSubagents bool

	cmd := &cobra.Command{
		Use:   "heatmap",
		Short: "Show Claude Code usage contribution heatmap",
		RunE: func(cmd *cobra.Command, args []string) error {
			common, err := claudeOptionsWithGroupByFlag(paths, cachePath, from, to, timezoneName, "table", "day", false, excludeSubagents)
			if err != nil {
				return err
			}
			common.RecentDataDays = 0
			common.RecentDataWeeks = 0
			out, err := app.RunClaudeHeatmap(cmd.Context(), app.HeatmapOptions{
				Options:      common,
				Color:        colorMode,
				ColorEnabled: resolveColorEnabled(colorMode, cmd.OutOrStdout()),
			})
			if err != nil {
				return err
			}
			cmd.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start date in YYYY-MM-DD")
	cmd.Flags().StringVar(&to, "to", "", "end date in YYYY-MM-DD")
	cmd.Flags().StringVar(&timezoneName, "timezone", "", "IANA timezone")
	cmd.Flags().StringVar(&colorMode, "color", "auto", "color mode: auto, never, always")
	cmd.Flags().StringArrayVar(&paths, "path", nil, "Claude projects directory")
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	cmd.Flags().BoolVar(&excludeSubagents, "exclude-subagents", false, "exclude subagent usage")
	return cmd
}

type claudeRunner func(context.Context, app.Options) (string, error)

func newClaudeReportLikeCommand(use string, short string, defaultGroupBy string, runner claudeRunner) *cobra.Command {
	var from, to, timezoneName, format, groupBy, cachePath, pricingPath string
	var paths []string
	var excludeSubagents bool

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := claudeOptionsWithPricingAndGroupByFlag(paths, cachePath, pricingPath, from, to, timezoneName, format, groupBy, cmd.Flags().Changed("group-by"), excludeSubagents)
			if err != nil {
				return err
			}
			if opts, err = resolvePricingConfig(opts, pricingPath); err != nil {
				return err
			}
			opts.ColorEnabled = resolveColorEnabled("auto", cmd.OutOrStdout())
			out, err := runner(cmd.Context(), opts)
			if err != nil {
				return err
			}
			cmd.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "start date in YYYY-MM-DD")
	cmd.Flags().StringVar(&to, "to", "", "end date in YYYY-MM-DD")
	cmd.Flags().StringVar(&timezoneName, "timezone", "", "IANA timezone")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table or json")
	cmd.Flags().StringVar(&groupBy, "group-by", defaultGroupBy, "grouping: day, week, month, year, range")
	cmd.Flags().StringArrayVar(&paths, "path", nil, "Claude projects directory")
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	cmd.Flags().StringVar(&pricingPath, "pricing", "", "pricing TOML path")
	cmd.Flags().BoolVar(&excludeSubagents, "exclude-subagents", false, "exclude subagent usage")
	return cmd
}

func claudeOptions(paths []string, cachePath string, from string, to string, timezoneName string, format string, groupBy string, excludeSubagents bool) (app.Options, error) {
	return claudeOptionsWithGroupByFlag(paths, cachePath, from, to, timezoneName, format, groupBy, false, excludeSubagents)
}

func claudeOptionsWithGroupByFlag(paths []string, cachePath string, from string, to string, timezoneName string, format string, groupBy string, groupByFlagChanged bool, excludeSubagents bool) (app.Options, error) {
	return claudeOptionsWithPricingAndGroupByFlag(paths, cachePath, "", from, to, timezoneName, format, groupBy, groupByFlagChanged, excludeSubagents)
}

func claudeOptionsWithPricingAndGroupByFlag(paths []string, cachePath string, pricingPath string, from string, to string, timezoneName string, format string, groupBy string, groupByFlagChanged bool, excludeSubagents bool) (app.Options, error) {
	if len(paths) == 0 {
		paths = platform.DefaultClaudeProjectsDirs()
	}
	if cachePath == "" {
		defaultCachePath, err := platform.DefaultCacheDBPath()
		if err != nil {
			return app.Options{}, err
		}
		cachePath = defaultCachePath
	}
	if timezoneName == "" {
		timezoneName = time.Local.String()
	}
	if _, err := time.LoadLocation(timezoneName); err != nil {
		return app.Options{}, err
	}
	recentDataDays := 0
	recentDataWeeks := 0
	if from == "" && to == "" {
		recentDataDays = 7
		if groupByFlagChanged && strings.EqualFold(groupBy, "week") {
			recentDataDays = 0
			recentDataWeeks = 3
		}
	}
	return app.Options{
		Paths:            paths,
		CachePath:        cachePath,
		From:             from,
		To:               to,
		Format:           format,
		GroupBy:          groupBy,
		Timezone:         timezoneName,
		ExcludeSubagents: excludeSubagents,
		RecentDataDays:   recentDataDays,
		RecentDataWeeks:  recentDataWeeks,
		PricingPath:      pricingPath,
	}, nil
}

func resolvePricingConfig(opts app.Options, explicitPath string) (app.Options, error) {
	if explicitPath != "" {
		opts.PricingPath = explicitPath
		return opts, nil
	}
	path, err := platform.DefaultPricingConfigPath()
	if err != nil {
		return app.Options{}, err
	}
	if err := pricing.EnsureDefaultConfig(path); err != nil {
		return app.Options{}, err
	}
	opts.PricingPath = path
	return opts, nil
}

func resolveColorEnabled(colorMode string, writer io.Writer) bool {
	switch colorMode {
	case "always":
		return true
	case "never":
		return false
	default:
		file, ok := writer.(*os.File)
		if !ok {
			return false
		}
		info, err := file.Stat()
		if err != nil {
			return false
		}
		return info.Mode()&os.ModeCharDevice != 0
	}
}
