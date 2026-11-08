// Package worker — фоновые задачи склада.
//
// AutoSync раз в 60 секунд проверяет тумблер в sync_settings и, если он включён
// и с последнего цикла прошло >= interval_minutes, запускает двустороннюю
// синхронизацию: pull всех справочников и документов из МС + push pending
// документов/заказов обратно в МС.
package worker

import (
"context"
"fmt"
"log/slog"
"sync"
"time"

"github.com/radonezhsklad/warehouse/internal/mspull"
"github.com/radonezhsklad/warehouse/internal/mspush"
"github.com/radonezhsklad/warehouse/internal/repository"
)

var pullKinds = []string{
"counterparties", "warehouses", "organizations", "projects",
"docs-enter", "docs-demand", "docs-loss", "docs-move",
"docs-purchaseorder", "docs-invoicein", "docs-purchasereturn",
"docs-internalorder",
}

type AutoSync struct {
jobs          *repository.SyncJobRepo
push          *mspush.Pusher
pull          *mspull.Runner
settings      *repository.SyncSettingsRepo
internalToken string

mu      sync.Mutex
running bool
tickN   int64
}

func NewAutoSync(jobs *repository.SyncJobRepo, push *mspush.Pusher, pull *mspull.Runner, settings *repository.SyncSettingsRepo, internalToken string) *AutoSync {
return &AutoSync{
jobs:          jobs,
push:          push,
pull:          pull,
settings:      settings,
internalToken: internalToken,
}
}

// Start всегда запускает фоновый цикл. Включение/выключение управляется
// тумблером в sync_settings (GET/PUT /api/v1/sync/settings).
func (a *AutoSync) Start(ctx context.Context) {
if a.settings == nil {
slog.Info("autosync: no settings repo — disabled")
return
}
slog.Info("autosync: starting (poll every 60s)")
go a.loop(ctx)
}

func (a *AutoSync) loop(ctx context.Context) {
t := time.NewTicker(60 * time.Second)
defer t.Stop()
for {
select {
case <-ctx.Done():
slog.Info("autosync: stopped")
return
case <-t.C:
a.maybeTick(ctx)
}
}
}

func (a *AutoSync) maybeTick(ctx context.Context) {
s, err := a.settings.Get(ctx)
if err != nil {
slog.Warn("autosync: settings.Get", "error", err)
return
}
if !s.Enabled {
return
}
interval := time.Duration(s.IntervalMinutes) * time.Minute
if interval <= 0 {
interval = 10 * time.Minute
}
if s.LastCycleAt != nil && time.Since(*s.LastCycleAt) < interval {
return
}
a.runCycle(ctx)
}

func (a *AutoSync) runCycle(ctx context.Context) {
a.mu.Lock()
if a.running {
a.mu.Unlock()
return
}
a.running = true
a.tickN++
a.mu.Unlock()
defer func() {
a.mu.Lock()
a.running = false
a.mu.Unlock()
}()

runCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
defer cancel()

slog.Info("autosync: cycle start")
_ = a.settings.SetLastCycle(runCtx, "running", "")

// Если уже идёт ручная задача — не мешаем.
if active, err := a.jobs.GetActive(runCtx); err == nil && active != nil {
slog.Info("autosync: skip, active job", "kind", active.Kind, "id", active.ID)
_ = a.settings.SetLastCycle(runCtx, "idle", "active job: "+active.Kind)
return
}

var firstErr error

// 1) PULL из МС.
if a.pull != nil && a.pull.Enabled() {
for _, kind := range pullKinds {
job, err := a.jobs.Create(runCtx, "pull-"+kind, "auto")
if err != nil {
if firstErr == nil {
firstErr = fmt.Errorf("create job %s: %w", kind, err)
}
continue
}
if err := a.jobs.MarkRunning(runCtx, job.ID); err != nil {
if firstErr == nil {
firstErr = fmt.Errorf("mark running %s: %w", kind, err)
}
continue
}
if err := a.pull.PullByKind(runCtx, kind, job.ID); err != nil {
_ = a.jobs.Finish(runCtx, job.ID, "error", err.Error())
if firstErr == nil {
firstErr = fmt.Errorf("pull %s: %w", kind, err)
}
slog.Warn("autosync: pull failed", "kind", kind, "error", err)
} else {
_ = a.jobs.Finish(runCtx, job.ID, "done", "")
}
}
}

// 2) PUSH в МС.
if a.push != nil && a.push.Enabled() {
attempted, succeeded, err := a.push.RetryPending(runCtx, a.internalToken)
if err != nil {
if firstErr == nil {
firstErr = fmt.Errorf("push: %w", err)
}
slog.Warn("autosync: push failed", "error", err)
} else {
slog.Info("autosync: push done", "attempted", attempted, "succeeded", succeeded)
}
}

if firstErr != nil {
_ = a.settings.SetLastCycle(runCtx, "error", firstErr.Error())
slog.Warn("autosync: cycle finished with errors", "error", firstErr)
} else {
_ = a.settings.SetLastCycle(runCtx, "ok", "")
slog.Info("autosync: cycle ok")
}
}

// Status — для GET /api/v1/sync/autosync.
func (a *AutoSync) Status() map[string]any {
a.mu.Lock()
defer a.mu.Unlock()
return map[string]any{
"running":      a.running,
"tick_count":   a.tickN,
"push_enabled": a.push != nil && a.push.Enabled(),
"pull_enabled": a.pull != nil && a.pull.Enabled(),
}
}