package cache

import (
	"context"
	"database/sql"
)

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent TEXT NOT NULL,
    path_raw TEXT NOT NULL,
    path_norm TEXT NOT NULL,
    size INTEGER NOT NULL,
    mtime_ns INTEGER NOT NULL,
    content_hash TEXT,
    parsed_at TEXT,
    status TEXT NOT NULL,
    error TEXT,
    UNIQUE(agent, path_norm)
);

CREATE TABLE IF NOT EXISTS projects (
    project_id TEXT PRIMARY KEY,
    agent TEXT NOT NULL,
    project_name TEXT NOT NULL,
    project_path_norm TEXT NOT NULL,
    project_path_display TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent TEXT NOT NULL,
    file_id INTEGER NOT NULL,
    raw_line_no INTEGER NOT NULL,
    event_hash TEXT NOT NULL,
    timestamp_utc TEXT NOT NULL,
    local_date TEXT NOT NULL,
    session_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    agent_id TEXT,
    parent_agent_id TEXT,
    is_subagent INTEGER NOT NULL DEFAULT 0,
    project_id TEXT NOT NULL,
    model TEXT NOT NULL,
    standard_input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_5m_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_1h_tokens INTEGER NOT NULL DEFAULT 0,
    cache_read_tokens INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(file_id) REFERENCES files(id),
    UNIQUE(agent, event_hash)
);

CREATE INDEX IF NOT EXISTS idx_events_agent_file ON events(agent, file_id);

CREATE TABLE IF NOT EXISTS daily_project_model_rollups (
    date TEXT NOT NULL,
    timezone TEXT NOT NULL,
    agent TEXT NOT NULL,
    project_id TEXT NOT NULL,
    model TEXT NOT NULL,
    is_subagent INTEGER NOT NULL DEFAULT 0,
    standard_input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_5m_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_1h_tokens INTEGER NOT NULL DEFAULT 0,
    cache_read_tokens INTEGER NOT NULL DEFAULT 0,
    event_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(date, timezone, agent, project_id, model, is_subagent)
);
`)
	return err
}
