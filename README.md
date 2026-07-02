# tokusage

Local CLI for estimating Claude Code token usage from transcript JSONL files.

## Status

Current implementation supports Claude Code CLI reports, terminal charts, and contribution-style heatmaps. TUI, query-result caching, and multi-agent support are not part of this phase.

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
tokusage claude heatmap
tokusage cache status
tokusage cache rebuild
tokusage cache clear
```

Default report scope is the latest 7 days that have usage data. Explicit `--group-by week` uses the latest 3 week buckets that have usage data. Use `--from` and `--to` to override the date range.

## Charts

`tokusage claude chart` renders terminal bar charts and splits by model by default. Use `--metric tokens` for token totals, `--metric cost` for estimated cost, `--split-by none` for plain bars, and `--split-by project` for project stacks. Use `--ascii` or `--color never` for log-friendly output.

`tokusage claude heatmap` renders a GitHub contributions-style daily usage grid. By default it shows the last year ending at the latest usage date; use `--from` and `--to` for an explicit range.

## Pricing

Costs are estimates. On first CLI use, tokusage creates `~/.tokusage/pricing.toml` with editable prices for common Anthropic, Kimi/Moonshot, GLM/Z.AI, DeepSeek, and MiniMax models. Unknown models fall back to Anthropic Sonnet pricing.

Use `--pricing path/to/pricing.toml` to use a different pricing file. Configured prices are checked first, then built-in prices are used as fallback.

```toml
[[models]]
pattern = "glm-5*"
standard_input_per_mtok = "1.00"
cache_write_5m_per_mtok = "1.00"
cache_write_1h_per_mtok = "1.00"
cache_read_per_mtok = "0.10"
output_per_mtok = "2.00"
```

Prices are USD per 1M tokens. `pattern` supports Go filepath-style matching such as `glm-5*`.

## Claude Code transcripts

The tool reads local transcripts from `~/.claude/projects` or `$CLAUDE_CONFIG_DIR/projects`. Use `--path` to point at a custom transcript root.
