package cache

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/breinzhang/tokusage/internal/domain"
)

func TestStoreEventsAndRebuildRollups(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	event := testEvent("hash-1", 100)

	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{event}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}

	rollups, err := QueryDailyRollups(ctx, db, RollupQuery{
		FromDate: "2026-05-10",
		ToDate:   "2026-05-10",
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rollups) != 1 {
		t.Fatalf("len(rollups) = %d, want 1", len(rollups))
	}
	if rollups[0].Tokens.TotalTokens() != 175 {
		t.Fatalf("TotalTokens = %d, want 175", rollups[0].Tokens.TotalTokens())
	}
}

func TestQueryDailyRollupsAllowsOpenDateRange(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	event := testEvent("hash-open-range", 100)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{event}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}

	rollups, err := QueryDailyRollups(ctx, db, RollupQuery{Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rollups) != 1 {
		t.Fatalf("len(rollups) = %d, want 1", len(rollups))
	}
	if rollups[0].Date != "2026-05-10" {
		t.Fatalf("Date = %q, want 2026-05-10", rollups[0].Date)
	}
}

func TestProjectUpsertPreservesCreatedAt(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	first := testEvent("hash-1", 100)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session-1.jsonl", []domain.UsageEvent{first}); err != nil {
		t.Fatal(err)
	}

	var firstCreatedAt string
	if err := db.QueryRowContext(ctx, `SELECT created_at FROM projects WHERE project_id = ?`, first.ProjectID).Scan(&firstCreatedAt); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)
	second := testEvent("hash-2", 200)
	second.ProjectName = "repo-renamed"
	second.ProjectPath = "/Users/example/work/repo-renamed"
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session-2.jsonl", []domain.UsageEvent{second}); err != nil {
		t.Fatal(err)
	}

	var createdAt, updatedAt, projectName, projectPathNorm, projectPathDisplay string
	if err := db.QueryRowContext(ctx, `
SELECT created_at, updated_at, project_name, project_path_norm, project_path_display
FROM projects
WHERE project_id = ?
`, first.ProjectID).Scan(&createdAt, &updatedAt, &projectName, &projectPathNorm, &projectPathDisplay); err != nil {
		t.Fatal(err)
	}
	if createdAt != firstCreatedAt {
		t.Fatalf("created_at = %q, want preserved %q", createdAt, firstCreatedAt)
	}
	if updatedAt < createdAt {
		t.Fatalf("updated_at = %q, want >= created_at %q", updatedAt, createdAt)
	}
	if projectName != second.ProjectName {
		t.Fatalf("project_name = %q, want %q", projectName, second.ProjectName)
	}
	if projectPathNorm != second.ProjectPath {
		t.Fatalf("project_path_norm = %q, want %q", projectPathNorm, second.ProjectPath)
	}
	if projectPathDisplay != second.ProjectPath {
		t.Fatalf("project_path_display = %q, want %q", projectPathDisplay, second.ProjectPath)
	}
}

func TestOpenSetsConnectionPragmas(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	var foreignKeys int
	if err := db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}

	var busyTimeout int
	if err := db.QueryRowContext(ctx, `PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
		t.Fatal(err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

func TestReplaceFileEventsRemovesOldEventsForFile(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	oldEvent := testEvent("old-hash", 100)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{oldEvent}); err != nil {
		t.Fatal(err)
	}

	newEvent := testEvent("new-hash", 200)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{newEvent}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}

	rollups := queryRollups(t, ctx, db, "UTC", "2026-05-10", false)
	if len(rollups) != 1 {
		t.Fatalf("len(rollups) = %d, want 1", len(rollups))
	}
	if rollups[0].Tokens.StandardInputTokens != 200 {
		t.Fatalf("StandardInputTokens = %d, want 200", rollups[0].Tokens.StandardInputTokens)
	}

	var oldEventCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE event_hash = ?`, oldEvent.EventHash).Scan(&oldEventCount); err != nil {
		t.Fatal(err)
	}
	if oldEventCount != 0 {
		t.Fatalf("old event count = %d, want 0", oldEventCount)
	}
}

func TestDuplicateEventHashDoesNotDoubleCount(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	first := testEvent("same-hash", 100)
	second := testEvent("same-hash", 300)
	second.MessageID = "message-2"
	second.RawLineNo = 2
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{first, second}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}

	rollups := queryRollups(t, ctx, db, "UTC", "2026-05-10", false)
	if len(rollups) != 1 {
		t.Fatalf("len(rollups) = %d, want 1", len(rollups))
	}
	if rollups[0].EventCount != 1 {
		t.Fatalf("EventCount = %d, want 1", rollups[0].EventCount)
	}
	if rollups[0].Tokens.StandardInputTokens != 100 {
		t.Fatalf("StandardInputTokens = %d, want 100", rollups[0].Tokens.StandardInputTokens)
	}
}

func TestRebuildRollupsUsesRequestedTimezoneDate(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	event := testEvent("hash-shanghai", 100)
	event.Timestamp = time.Date(2026, 5, 9, 16, 30, 0, 0, time.UTC)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{event}); err != nil {
		t.Fatal(err)
	}
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "Asia/Shanghai", shanghai); err != nil {
		t.Fatal(err)
	}

	rollups := queryRollups(t, ctx, db, "Asia/Shanghai", "2026-05-10", false)
	if len(rollups) != 1 {
		t.Fatalf("len(rollups) = %d, want 1", len(rollups))
	}
	if rollups[0].Date != "2026-05-10" {
		t.Fatalf("Date = %q, want 2026-05-10", rollups[0].Date)
	}
}

func TestRebuildRollupsIsIdempotent(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	event := testEvent("hash-1", 100)
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{event}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}
	first := queryRollups(t, ctx, db, "UTC", "2026-05-10", false)
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}
	second := queryRollups(t, ctx, db, "UTC", "2026-05-10", false)

	if len(first) != len(second) {
		t.Fatalf("second len = %d, want %d", len(second), len(first))
	}
	if len(second) != 1 {
		t.Fatalf("len(second) = %d, want 1", len(second))
	}
	if second[0].EventCount != first[0].EventCount {
		t.Fatalf("second EventCount = %d, want %d", second[0].EventCount, first[0].EventCount)
	}
	if second[0].Tokens.TotalTokens() != first[0].Tokens.TotalTokens() {
		t.Fatalf("second TotalTokens = %d, want %d", second[0].Tokens.TotalTokens(), first[0].Tokens.TotalTokens())
	}
}

func TestQueryDailyRollupsExcludeSubagents(t *testing.T) {
	ctx, db := openTestDB(t)
	defer db.Close()

	regular := testEvent("regular-hash", 100)
	subagent := testEvent("subagent-hash", 200)
	subagent.IsSubagent = true
	subagent.MessageID = "message-subagent"
	subagent.RawLineNo = 2
	if err := ReplaceFileEvents(ctx, db, "claude-code", "/tmp/session.jsonl", []domain.UsageEvent{regular, subagent}); err != nil {
		t.Fatal(err)
	}
	if err := RebuildRollups(ctx, db, "UTC", time.UTC); err != nil {
		t.Fatal(err)
	}

	allRollups := queryRollups(t, ctx, db, "UTC", "2026-05-10", false)
	if len(allRollups) != 2 {
		t.Fatalf("len(allRollups) = %d, want 2", len(allRollups))
	}

	filtered := queryRollups(t, ctx, db, "UTC", "2026-05-10", true)
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].IsSubagent {
		t.Fatal("filtered rollup includes subagent")
	}
	if filtered[0].Tokens.StandardInputTokens != 100 {
		t.Fatalf("StandardInputTokens = %d, want 100", filtered[0].Tokens.StandardInputTokens)
	}
}

func openTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()

	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "tokusage.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(ctx, db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return ctx, db
}

func testEvent(hash string, inputTokens int64) domain.UsageEvent {
	return domain.UsageEvent{
		Agent:       "claude-code",
		ProjectID:   "project-1",
		ProjectName: "repo",
		ProjectPath: "/Users/example/work/repo",
		SessionID:   "session-1",
		MessageID:   "message-1",
		Model:       "claude-sonnet-4.5",
		Timestamp:   time.Date(2026, 5, 10, 1, 2, 3, 0, time.UTC),
		Tokens: domain.TokenSummary{
			StandardInputTokens: inputTokens,
			OutputTokens:        20,
			CacheWrite5mTokens:  10,
			CacheWrite1hTokens:  5,
			CacheReadTokens:     40,
		},
		SourceFile: "/tmp/session.jsonl",
		RawLineNo:  1,
		EventHash:  hash,
	}
}

func queryRollups(t *testing.T, ctx context.Context, db *sql.DB, timezone string, date string, excludeSubagents bool) []DailyRollup {
	t.Helper()

	rollups, err := QueryDailyRollups(ctx, db, RollupQuery{
		FromDate:         date,
		ToDate:           date,
		Timezone:         timezone,
		ExcludeSubagents: excludeSubagents,
	})
	if err != nil {
		t.Fatal(err)
	}
	return rollups
}
