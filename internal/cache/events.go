package cache

import (
	"context"
	"database/sql"
	"time"

	"github.com/breinzhang/tokusage/internal/domain"
)

func ReplaceFileEvents(ctx context.Context, db *sql.DB, agent string, pathNorm string, events []domain.UsageEvent) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	fileID, err := upsertFile(ctx, tx, agent, pathNorm)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE agent = ? AND file_id = ?`, agent, fileID); err != nil {
		return err
	}
	for _, event := range events {
		if err := insertProject(ctx, tx, event); err != nil {
			return err
		}
		if err := insertEvent(ctx, tx, fileID, event); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func upsertFile(ctx context.Context, tx *sql.Tx, agent string, pathNorm string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO files (agent, path_raw, path_norm, size, mtime_ns, content_hash, parsed_at, status, error)
VALUES (?, ?, ?, 0, 0, NULL, ?, 'parsed', '')
ON CONFLICT(agent, path_norm) DO UPDATE SET
    path_raw = excluded.path_raw,
    size = excluded.size,
    mtime_ns = excluded.mtime_ns,
    content_hash = excluded.content_hash,
    parsed_at = excluded.parsed_at,
    status = excluded.status,
    error = excluded.error
`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	if _, err := stmt.ExecContext(ctx, agent, pathNorm, pathNorm, now); err != nil {
		return 0, err
	}

	var fileID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM files WHERE agent = ? AND path_norm = ?`, agent, pathNorm).Scan(&fileID)
	return fileID, err
}

func insertProject(ctx context.Context, tx *sql.Tx, event domain.UsageEvent) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO projects (
    project_id,
    agent,
    project_name,
    project_path_norm,
    project_path_display,
    created_at,
    updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id) DO UPDATE SET
    agent = excluded.agent,
    project_name = excluded.project_name,
    project_path_norm = excluded.project_path_norm,
    project_path_display = excluded.project_path_display,
    updated_at = excluded.updated_at
`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx,
		event.ProjectID,
		event.Agent,
		event.ProjectName,
		event.ProjectPath,
		event.ProjectPath,
		now,
		now,
	)
	return err
}

func insertEvent(ctx context.Context, tx *sql.Tx, fileID int64, event domain.UsageEvent) error {
	stmt, err := tx.PrepareContext(ctx, `
INSERT OR IGNORE INTO events (
    agent,
    file_id,
    raw_line_no,
    event_hash,
    timestamp_utc,
    local_date,
    session_id,
    message_id,
    agent_id,
    parent_agent_id,
    is_subagent,
    project_id,
    model,
    standard_input_tokens,
    output_tokens,
    cache_write_5m_tokens,
    cache_write_1h_tokens,
    cache_read_tokens
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	timestampUTC := event.Timestamp.UTC()
	_, err = stmt.ExecContext(ctx,
		event.Agent,
		fileID,
		event.RawLineNo,
		event.EventHash,
		timestampUTC.Format(time.RFC3339Nano),
		timestampUTC.Format(time.DateOnly),
		event.SessionID,
		event.MessageID,
		event.AgentID,
		event.ParentAgentID,
		boolToInt(event.IsSubagent),
		event.ProjectID,
		event.Model,
		event.Tokens.StandardInputTokens,
		event.Tokens.OutputTokens,
		event.Tokens.CacheWrite5mTokens,
		event.Tokens.CacheWrite1hTokens,
		event.Tokens.CacheReadTokens,
	)
	return err
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
