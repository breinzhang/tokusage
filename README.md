# tokusage

Local CLI for estimating Claude Code token usage from transcript JSONL files.

## Status

Current implementation supports Claude Code CLI reports and terminal charts. TUI, query-result caching, and multi-agent support are not part of this phase.

## Usage

```bash
tokusage claude report
tokusage claude models
tokusage claude projects
tokusage claude chart --group-by day --metric tokens
tokusage claude chart --group-by day --metric cost
tokusage claude chart --group-by month --metric cost --split-by model
tokusage claude chart --group-by month --metric cost --split-by project
tokusage claude chart --group-by year --metric cost
tokusage cache status
tokusage cache rebuild
tokusage cache clear
```

Default report scope is the latest 7 days that have usage data. Explicit `--group-by week` uses the latest 3 week buckets that have usage data. Use `--from` and `--to` to override the date range.

## Charts

`tokusage claude chart` renders terminal bar charts and splits by model by default. Use `--metric tokens` for token totals, `--metric cost` for estimated cost, `--split-by none` for plain bars, and `--split-by project` for project stacks. Use `--ascii` or `--color never` for log-friendly output.

## Pricing

Costs are estimates. Built-in Anthropic model prices are included. Unknown or non-Anthropic models fall back to Anthropic Sonnet pricing.

## Claude Code transcripts

The tool reads local transcripts from `~/.claude/projects` or `$CLAUDE_CONFIG_DIR/projects`. Use `--path` to point at a custom transcript root.
