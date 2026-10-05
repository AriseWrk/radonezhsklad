package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncJob — persistent-запись о задаче синхронизации.
// Поля совпадают с прежним in-memory SyncJob, чтобы фронт не пришлось менять.
type SyncJob struct {
	ID           uuid.UUID  `json:"id"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	LastLine     *string    `json:"last_line,omitempty"`
	Error        *string    `json:"error,omitempty"`
	Fetched      int        `json:"fetched"`
	Total        *int       `json:"total,omitempty"`
	CursorBefore *string    `json:"cursor_before,omitempty"`
	CursorAfter  *string    `json:"cursor_after,omitempty"`
	TriggeredBy  string     `json:"triggered_by"`
}

type SyncJobRepo struct {
	db *pgxpool.Pool
}

func NewSyncJobRepo(db *pgxpool.Pool) *SyncJobRepo { return &SyncJobRepo{db: db} }

const syncJobSelect = `
SELECT id, kind, status, started_at, ended_at, last_line, error,
       fetched, total, cursor_before, cursor_after, triggered_by
FROM sync_jobs`

func scanSyncJob(row pgx.Row) (*SyncJob, error) {
	var j SyncJob
	err := row.Scan(
		&j.ID, &j.Kind, &j.Status, &j.StartedAt, &j.EndedAt,
		&j.LastLine, &j.Error, &j.Fetched, &j.Total,
		&j.CursorBefore, &j.CursorAfter, &j.TriggeredBy,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// Create — заводит новую задачу в статусе queued.
func (r *SyncJobRepo) Create(ctx context.Context, kind, triggeredBy string) (*SyncJob, error) {
	if triggeredBy == "" {
		triggeredBy = "manual"
	}

	// Курсор читаем отдельно — иначе PG ругается на $2 в двух местах.
	var cursorBefore *string
	err := r.db.QueryRow(ctx,
		`SELECT last_cursor FROM v_sync_cursor WHERE kind = $1`, kind).Scan(&cursorBefore)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	id := uuid.New()
	_, err = r.db.Exec(ctx, `
INSERT INTO sync_jobs (id, kind, status, triggered_by, cursor_before)
VALUES ($1, $2, 'queued', $3, $4)`,
		id, kind, triggeredBy, cursorBefore)
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

// Get — читает задачу по id.
func (r *SyncJobRepo) Get(ctx context.Context, id uuid.UUID) (*SyncJob, error) {
	row := r.db.QueryRow(ctx, syncJobSelect+" WHERE id = $1", id)
	j, err := scanSyncJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return j, err
}

// GetActive — возвращает текущую queued/running задачу, если есть.
func (r *SyncJobRepo) GetActive(ctx context.Context) (*SyncJob, error) {
	row := r.db.QueryRow(ctx,
		syncJobSelect+" WHERE status IN ('queued','running') ORDER BY started_at DESC LIMIT 1")
	j, err := scanSyncJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return j, err
}

// List — история задач, свежие первыми.
func (r *SyncJobRepo) List(ctx context.Context, limit int) ([]*SyncJob, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(ctx, syncJobSelect+" ORDER BY started_at DESC LIMIT $1", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*SyncJob, 0, limit)
	for rows.Next() {
		j, err := scanSyncJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// MarkRunning — queued -> running.
func (r *SyncJobRepo) MarkRunning(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sync_jobs SET status='running' WHERE id=$1`, id)
	return err
}

// SetLastLine — обновляет только last_line (для прогресс-потока).
func (r *SyncJobRepo) SetLastLine(ctx context.Context, id uuid.UUID, line string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sync_jobs SET last_line=$2 WHERE id=$1`, id, line)
	return err
}

// SetProgress — обновляет fetched/total.
func (r *SyncJobRepo) SetProgress(ctx context.Context, id uuid.UUID, fetched, total int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sync_jobs SET fetched=$2, total=$3 WHERE id=$1`, id, fetched, total)
	return err
}

// Finish — переводит задачу в терминальное состояние.
// status: done | error | cancelled
func (r *SyncJobRepo) Finish(ctx context.Context, id uuid.UUID, status, errMsg string) error {
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	_, err := r.db.Exec(ctx, `
UPDATE sync_jobs
SET status=$2, error=$3, ended_at=NOW()
WHERE id=$1`, id, status, errPtr)
	return err
}

// SetCursorAfter — записывает cursor_after (для v_sync_cursor).
func (r *SyncJobRepo) SetCursorAfter(ctx context.Context, id uuid.UUID, cursor string) error {
	var c *string
	if cursor != "" {
		c = &cursor
	}
	_, err := r.db.Exec(ctx,
		`UPDATE sync_jobs SET cursor_after=$2 WHERE id=$1`, id, c)
	return err
}

// GetLastCursor — последний успешный курсор по kind.
func (r *SyncJobRepo) GetLastCursor(ctx context.Context, kind string) (string, error) {
	var c *string
	err := r.db.QueryRow(ctx,
		`SELECT last_cursor FROM v_sync_cursor WHERE kind=$1`, kind).Scan(&c)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if c == nil {
		return "", nil
	}
	return *c, nil
}

// MarkStaleAsError — на старте api.exe помечаем «зависшие» задачи как упавшие.
// Возвращает количество затронутых строк.
func (r *SyncJobRepo) MarkStaleAsError(ctx context.Context) (int64, error) {
	ct, err := r.db.Exec(ctx, `
UPDATE sync_jobs
SET status='error',
    error='прервано рестартом api.exe',
    ended_at=NOW()
WHERE status IN ('queued','running')`)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
