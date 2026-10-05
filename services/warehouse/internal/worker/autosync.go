// Package worker — фоновые задачи склада.
//
// AutoSync раз в interval пробегает retry-pending (см. mspush.RetryPending)
// и пишет результат в лог. Pull справочников и документов — вручную через
// /sync/pull/:kind (чтобы не заливать МС при случайном флаге в .env).
package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/radonezhsklad/warehouse/internal/mspush"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type AutoSync struct {
	jobs          *repository.SyncJobRepo
	push          *mspush.Pusher
	internalToken string
	interval      time.Duration
	enabled       bool

	mu       sync.Mutex
	lastTick time.Time
	lastErr  string
	tickN    int64
}

func NewAutoSync(jobs *repository.SyncJobRepo, push *mspush.Pusher, internalToken string, interval time.Duration, enabled bool) *AutoSync {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return &AutoSync{
		jobs:          jobs,
		push:          push,
		internalToken: internalToken,
		interval:      interval,
		enabled:       enabled,
	}
}

// Start — запускает горутину. Возвращается сразу.
func (a *AutoSync) Start(ctx context.Context) {
	if !a.enabled {
		slog.Info("autosync: disabled (MS_AUTOSYNC_ENABLED != true)")
		return
	}
	if a.push == nil || !a.push.Enabled() {
		slog.Warn("autosync: enabled, но mspush выключен — автосинк не работает")
		return
	}
	slog.Info("autosync: starting", "interval", a.interval.String())
	go a.loop(ctx)
}

func (a *AutoSync) loop(ctx context.Context) {
	t := time.NewTicker(a.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("autosync: stopped")
			return
		case <-t.C:
			a.tick(ctx)
		}
	}
}

func (a *AutoSync) tick(ctx context.Context) {
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Если что-то уже бежит — не мешаем.
	if active, err := a.jobs.GetActive(runCtx); err == nil && active != nil {
		slog.Info("autosync: skip, active job", "kind", active.Kind, "id", active.ID)
		return
	}

	a.mu.Lock()
	a.lastTick = time.Now()
	a.tickN++
	a.mu.Unlock()

	attempted, succeeded, err := a.push.RetryPending(runCtx, a.internalToken)
	if err != nil {
		a.mu.Lock()
		a.lastErr = err.Error()
		a.mu.Unlock()
		slog.Warn("autosync: retry failed", "error", err)
		return
	}

	a.mu.Lock()
	a.lastErr = ""
	a.mu.Unlock()
	slog.Info("autosync: tick done", "attempted", attempted, "succeeded", succeeded)
}

// Status — для эндпоинта /sync/autosync.
func (a *AutoSync) Status() map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()
	var lastTick *time.Time
	if !a.lastTick.IsZero() {
		lt := a.lastTick
		lastTick = &lt
	}
	return map[string]any{
		"enabled":      a.enabled,
		"interval_s":   int(a.interval.Seconds()),
		"last_tick":    lastTick,
		"last_error":   a.lastErr,
		"tick_count":   a.tickN,
		"push_enabled": a.push != nil && a.push.Enabled(),
	}
}
