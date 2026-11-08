package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncSettings struct {
	Enabled         bool       `json:"enabled"`
	IntervalMinutes int        `json:"interval_minutes"`
	LastCycleAt     *time.Time `json:"last_cycle_at"`
	LastSlowCycleAt *time.Time `json:"last_slow_cycle_at"`
	LastStatus      string     `json:"last_status"`
	LastError       string     `json:"last_error"`
}

type SyncSettingsRepo struct {
	pool *pgxpool.Pool
}

func NewSyncSettingsRepo(pool *pgxpool.Pool) *SyncSettingsRepo {
	return &SyncSettingsRepo{pool: pool}
}

func (r *SyncSettingsRepo) Get(ctx context.Context) (*SyncSettings, error) {
	var s SyncSettings
	var lastCycleAt *time.Time
	var lastErr *string
	var lastSlowCycleAt *time.Time
	err := r.pool.QueryRow(ctx, `
SELECT enabled, interval_minutes, last_cycle_at, last_slow_cycle_at, last_status, last_error
FROM sync_settings WHERE id = 1
`).Scan(&s.Enabled, &s.IntervalMinutes, &lastCycleAt, &lastSlowCycleAt, &s.LastStatus, &lastErr)
	if err != nil {
		return nil, err
	}
	s.LastCycleAt = lastCycleAt
	if lastErr != nil {
		s.LastError = *lastErr
	}
	return &s, nil
}

func (r *SyncSettingsRepo) SetEnabled(ctx context.Context, enabled bool) error {
	_, err := r.pool.Exec(ctx, `
UPDATE sync_settings SET enabled = $1, updated_at = now() WHERE id = 1
`, enabled)
	return err
}

func (r *SyncSettingsRepo) SetLastCycle(ctx context.Context, status, errMsg string, slowDone bool) error {
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	var err error
	if slowDone {
		_, err = r.pool.Exec(ctx, `
UPDATE sync_settings
SET last_cycle_at = now(), last_slow_cycle_at = now(), last_status = $1, last_error = $2, updated_at = now()
WHERE id = 1
`, status, errPtr)
	} else {
		_, err = r.pool.Exec(ctx, `
UPDATE sync_settings
SET last_cycle_at = now(), last_status = $1, last_error = $2, updated_at = now()
WHERE id = 1
`, status, errPtr)
	}
	return err
}
