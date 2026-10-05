package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MsSyncError — запись журнала ошибок.
type MsSyncError struct {
	ID           uuid.UUID  `json:"id"`
	Entity       string     `json:"entity"`
	LocalID      *uuid.UUID `json:"local_id,omitempty"`
	MsExternalID *uuid.UUID `json:"ms_external_id,omitempty"`
	Op           string     `json:"op"`
	Attempt      int        `json:"attempt"`
	Error        string     `json:"error"`
	CreatedAt    time.Time  `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type MsSyncErrorRepo struct {
	db *pgxpool.Pool
}

func NewMsSyncErrorRepo(db *pgxpool.Pool) *MsSyncErrorRepo {
	return &MsSyncErrorRepo{db: db}
}

// Append — записать ошибку.
func (r *MsSyncErrorRepo) Append(ctx context.Context, e MsSyncError) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO ms_sync_errors (entity, local_id, ms_external_id, op, attempt, error)
VALUES ($1, $2, $3, $4, $5, $6)`,
		e.Entity, e.LocalID, e.MsExternalID, e.Op, e.Attempt, e.Error)
	return err
}

// ResolveOpen — пометить все открытые ошибки local_id как решённые.
func (r *MsSyncErrorRepo) ResolveOpen(ctx context.Context, entity string, localID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE ms_sync_errors
SET resolved_at = NOW()
WHERE entity=$1 AND local_id=$2 AND resolved_at IS NULL`,
		entity, localID)
	return err
}

// ListOpen — открытые ошибки, свежие первыми.
func (r *MsSyncErrorRepo) ListOpen(ctx context.Context, limit int) ([]*MsSyncError, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
SELECT id, entity, local_id, ms_external_id, op, attempt, error, created_at, resolved_at
FROM ms_sync_errors
WHERE resolved_at IS NULL
ORDER BY created_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*MsSyncError, 0, limit)
	for rows.Next() {
		var e MsSyncError
		if err := rows.Scan(&e.ID, &e.Entity, &e.LocalID, &e.MsExternalID, &e.Op,
			&e.Attempt, &e.Error, &e.CreatedAt, &e.ResolvedAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// CountOpen — сколько открытых.
func (r *MsSyncErrorRepo) CountOpen(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM ms_sync_errors WHERE resolved_at IS NULL`).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return n, err
}

// ListPendingDocuments — документы, которые не улетели в МС, но должны.
//
//	source='manual', status='posted', external_id IS NULL, ms_sync_error IS NOT NULL
//
// updated_at за последние 24 часа (защита от бесконечного retry).
func (r *MsSyncErrorRepo) ListPendingDocuments(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
SELECT id FROM documents
WHERE source = 'manual'
  AND status = 'posted'
  AND external_id IS NULL
  AND ms_sync_error IS NOT NULL
  AND updated_at > NOW() - INTERVAL '24 hours'
ORDER BY updated_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ListPendingOrders — внутренние заказы, которые не улетели в МС.
func (r *MsSyncErrorRepo) ListPendingOrders(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, `
SELECT id FROM internal_orders
WHERE source = 'manual'
  AND status IN ('draft','posted')
  AND external_id IS NULL
  AND ms_sync_error IS NOT NULL
  AND updated_at > NOW() - INTERVAL '24 hours'
ORDER BY updated_at DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
