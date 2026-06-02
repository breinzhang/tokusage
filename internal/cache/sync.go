package cache

import (
	"context"
	"database/sql"
)

type SyncStats struct {
	FilesParsed int
	EventsSaved int
	Warnings    int
}

type StatusResult struct {
	Path        string
	FileCount   int64
	EventCount  int64
	RollupCount int64
}

func Status(ctx context.Context, db *sql.DB, path string) (StatusResult, error) {
	result := StatusResult{Path: path}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM files`).Scan(&result.FileCount); err != nil {
		return result, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events`).Scan(&result.EventCount); err != nil {
		return result, err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM daily_project_model_rollups`).Scan(&result.RollupCount); err != nil {
		return result, err
	}
	return result, nil
}
