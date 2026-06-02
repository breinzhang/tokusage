package cache

import (
	"context"
	"database/sql"
	"time"

	"github.com/breinzhang/tokusage/internal/domain"
)

type DailyRollup struct {
	Date               string
	Agent              string
	ProjectID          string
	ProjectName        string
	ProjectPathDisplay string
	Model              string
	IsSubagent         bool
	Tokens             domain.TokenSummary
	EventCount         int64
}

type RollupQuery struct {
	FromDate         string
	ToDate           string
	Timezone         string
	ExcludeSubagents bool
	ProjectIDs       []string
	ModelPatterns    []string
}

func RebuildRollups(ctx context.Context, db *sql.DB, timezoneName string, loc *time.Location) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM daily_project_model_rollups WHERE timezone = ?`, timezoneName); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
    timestamp_utc,
    agent,
    project_id,
    model,
    is_subagent,
    standard_input_tokens,
    output_tokens,
    cache_write_5m_tokens,
    cache_write_1h_tokens,
    cache_read_tokens
FROM events
`)
	if err != nil {
		return err
	}
	defer rows.Close()

	rollups := map[rollupKey]DailyRollup{}
	for rows.Next() {
		var (
			timestampUTC string
			rollup       DailyRollup
			isSubagent   int
		)
		if err := rows.Scan(
			&timestampUTC,
			&rollup.Agent,
			&rollup.ProjectID,
			&rollup.Model,
			&isSubagent,
			&rollup.Tokens.StandardInputTokens,
			&rollup.Tokens.OutputTokens,
			&rollup.Tokens.CacheWrite5mTokens,
			&rollup.Tokens.CacheWrite1hTokens,
			&rollup.Tokens.CacheReadTokens,
		); err != nil {
			return err
		}
		timestamp, err := time.Parse(time.RFC3339Nano, timestampUTC)
		if err != nil {
			return err
		}
		rollup.Date = timestamp.In(loc).Format(time.DateOnly)
		rollup.IsSubagent = isSubagent != 0
		rollup.EventCount = 1

		key := rollupKey{
			date:       rollup.Date,
			agent:      rollup.Agent,
			projectID:  rollup.ProjectID,
			model:      rollup.Model,
			isSubagent: rollup.IsSubagent,
		}
		existing := rollups[key]
		if existing.EventCount == 0 {
			rollups[key] = rollup
			continue
		}
		existing.Tokens = existing.Tokens.Add(rollup.Tokens)
		existing.EventCount++
		rollups[key] = existing
	}
	if err := rows.Err(); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO daily_project_model_rollups (
    date,
    timezone,
    agent,
    project_id,
    model,
    is_subagent,
    standard_input_tokens,
    output_tokens,
    cache_write_5m_tokens,
    cache_write_1h_tokens,
    cache_read_tokens,
    event_count
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, rollup := range rollups {
		if _, err := stmt.ExecContext(ctx,
			rollup.Date,
			timezoneName,
			rollup.Agent,
			rollup.ProjectID,
			rollup.Model,
			boolToInt(rollup.IsSubagent),
			rollup.Tokens.StandardInputTokens,
			rollup.Tokens.OutputTokens,
			rollup.Tokens.CacheWrite5mTokens,
			rollup.Tokens.CacheWrite1hTokens,
			rollup.Tokens.CacheReadTokens,
			rollup.EventCount,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func QueryDailyRollups(ctx context.Context, db *sql.DB, query RollupQuery) ([]DailyRollup, error) {
	rows, err := db.QueryContext(ctx, `
SELECT
    r.date,
    r.agent,
    r.project_id,
    COALESCE(p.project_name, ''),
    COALESCE(p.project_path_display, ''),
    r.model,
    r.is_subagent,
    r.standard_input_tokens,
    r.output_tokens,
    r.cache_write_5m_tokens,
    r.cache_write_1h_tokens,
    r.cache_read_tokens,
    r.event_count
FROM daily_project_model_rollups r
LEFT JOIN projects p ON p.agent = r.agent AND p.project_id = r.project_id
WHERE r.timezone = ?
  AND (? = '' OR r.date >= ?)
  AND (? = '' OR r.date <= ?)
ORDER BY r.date, r.project_id, r.model, r.is_subagent
`, query.Timezone, query.FromDate, query.FromDate, query.ToDate, query.ToDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rollups []DailyRollup
	for rows.Next() {
		var (
			rollup     DailyRollup
			isSubagent int
		)
		if err := rows.Scan(
			&rollup.Date,
			&rollup.Agent,
			&rollup.ProjectID,
			&rollup.ProjectName,
			&rollup.ProjectPathDisplay,
			&rollup.Model,
			&isSubagent,
			&rollup.Tokens.StandardInputTokens,
			&rollup.Tokens.OutputTokens,
			&rollup.Tokens.CacheWrite5mTokens,
			&rollup.Tokens.CacheWrite1hTokens,
			&rollup.Tokens.CacheReadTokens,
			&rollup.EventCount,
		); err != nil {
			return nil, err
		}
		rollup.IsSubagent = isSubagent != 0
		if query.ExcludeSubagents && rollup.IsSubagent {
			continue
		}
		rollups = append(rollups, rollup)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rollups, nil
}

type rollupKey struct {
	date       string
	agent      string
	projectID  string
	model      string
	isSubagent bool
}
