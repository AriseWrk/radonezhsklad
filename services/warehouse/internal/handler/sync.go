package handler

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/radonezhsklad/shared/httpx"
	"github.com/radonezhsklad/warehouse/internal/mspull"
	"github.com/radonezhsklad/warehouse/internal/mspush"
	"github.com/radonezhsklad/warehouse/internal/repository"
	"github.com/radonezhsklad/warehouse/internal/worker"
)

// SyncHandler запускает PS-скрипты синхронизации и хранит состояние в sync_jobs.
type SyncHandler struct {
	autoSync      *worker.AutoSync
	push          *mspush.Pusher
	internalToken string
	errs          *repository.MsSyncErrorRepo
	mu            sync.Mutex // защита от двух одновременных Start в одном процессе
	repo          *repository.SyncJobRepo
	pull          *mspull.Runner
	scriptsDir    string
	lastLineAt    map[uuid.UUID]time.Time
}

func NewSyncHandler(repo *repository.SyncJobRepo, errs *repository.MsSyncErrorRepo, pull *mspull.Runner, push *mspush.Pusher, autoSync *worker.AutoSync, internalToken, scriptsDir string) *SyncHandler {
	return &SyncHandler{
		repo:          repo,
		errs:          errs,
		push:          push,
		autoSync:      autoSync,
		internalToken: internalToken,
		pull:          pull,
		scriptsDir:    scriptsDir,
		lastLineAt:    map[uuid.UUID]time.Time{},
	}
}

var (
	progressRe = regexp.MustCompile(`\[PROGRESS\] fetched=(\d+) total=(\d+)`)
	doneRe     = regexp.MustCompile(`\[DONE\] ok`)
	failRe     = regexp.MustCompile(`\[FAIL\]\s*(.+)$`)
)

// Start — POST /api/v1/sync/pull-orders
func (h *SyncHandler) Start(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Проверяем, нет ли активной задачи в БД (в т.ч. оставшейся от прошлого процесса).
	if active, err := h.repo.GetActive(c.Request.Context()); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	} else if active != nil {
		c.JSON(409, gin.H{"error": gin.H{
			"code":    "conflict",
			"message": "синхронизация уже идёт",
			"job_id":  active.ID,
		}})
		return
	}

	job, err := h.repo.Create(c.Request.Context(), "pull-orders", "manual")
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}

	go h.run(job)

	httpx.Created(c, job)
}

// StartPull — POST /api/v1/sync/pull/:kind
// Поддерживаемые kind: counterparties|warehouses|organizations|projects
// (можно с префиксом pull-, он обрезается).
func (h *SyncHandler) StartPull(c *gin.Context) {
	if h.pull == nil || !h.pull.Enabled() {
		c.JSON(503, gin.H{"error": gin.H{
			"code":    "unavailable",
			"message": "MS pull отключён (нет MS_TOKEN/MS_TOKEN_FILE или MS_PUSH_ENABLED=false)",
		}})
		return
	}
	kind := strings.ToLower(c.Param("kind"))
	kind = strings.TrimPrefix(kind, "pull-")
	if kind == "" {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "kind обязателен"}})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if active, err := h.repo.GetActive(c.Request.Context()); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	} else if active != nil {
		c.JSON(409, gin.H{"error": gin.H{
			"code":    "conflict",
			"message": "синхронизация уже идёт",
			"job_id":  active.ID,
		}})
		return
	}

	job, err := h.repo.Create(c.Request.Context(), "pull-"+kind, "manual")
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}

	go h.runPull(job, kind)

	httpx.Created(c, job)
}

func (h *SyncHandler) runPull(job *repository.SyncJob, kind string) {
	defer func() {
		if r := recover(); r != nil {
			h.finish(job, "error", fmt.Sprintf("panic: %v", r))
		}
	}()

	ctx := context.Background()

	if err := h.repo.MarkRunning(ctx, job.ID); err != nil {
		h.finish(job, "error", fmt.Sprintf("mark running: %v", err))
		return
	}

	if err := h.pull.PullByKind(ctx, kind, job.ID); err != nil {
		h.finish(job, "error", err.Error())
		return
	}
	h.finish(job, "done", "")
}

// ListErrors — GET /api/v1/sync/errors?limit=100
func (h *SyncHandler) ListErrors(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
	}
	if h.errs == nil {
		httpx.OK(c, gin.H{"items": []any{}, "count_open": 0})
		return
	}
	items, err := h.errs.ListOpen(c.Request.Context(), limit)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}
	n, _ := h.errs.CountOpen(c.Request.Context())
	httpx.OK(c, gin.H{"items": items, "count_open": n})
}

// StartRetryPending — POST /api/v1/sync/retry-pending
// Перебирает documents/internal_orders с ms_sync_error и пробует запушить заново.
func (h *SyncHandler) StartRetryPending(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if active, err := h.repo.GetActive(c.Request.Context()); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	} else if active != nil {
		c.JSON(409, gin.H{"error": gin.H{
			"code":    "conflict",
			"message": "синхронизация уже идёт",
			"job_id":  active.ID,
		}})
		return
	}

	job, err := h.repo.Create(c.Request.Context(), "retry-pending", "manual")
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}

	go h.runRetryPending(job)
	httpx.Created(c, job)
}

func (h *SyncHandler) runRetryPending(job *repository.SyncJob) {
	defer func() {
		if r := recover(); r != nil {
			h.finish(job, "error", fmt.Sprintf("panic: %v", r))
		}
	}()

	ctx := context.Background()
	if err := h.repo.MarkRunning(ctx, job.ID); err != nil {
		h.finish(job, "error", fmt.Sprintf("mark running: %v", err))
		return
	}
	if h.push == nil || !h.push.Enabled() {
		h.finish(job, "error", "mspush disabled")
		return
	}
	attempted, succeeded, err := h.push.RetryPending(ctx, h.internalToken)
	if err != nil {
		h.finish(job, "error", err.Error())
		return
	}
	_ = h.repo.SetProgress(ctx, job.ID, attempted, attempted)
	_ = h.repo.SetLastLine(ctx, job.ID,
		fmt.Sprintf("[DONE] attempted=%d succeeded=%d", attempted, succeeded))
	h.finish(job, "done", "")
}

// StartRetryDocument — POST /api/v1/sync/retry/document/:id
func (h *SyncHandler) StartRetryDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid id"}})
		return
	}
	if h.push == nil || !h.push.Enabled() {
		c.JSON(503, gin.H{"error": gin.H{"code": "unavailable", "message": "mspush disabled"}})
		return
	}
	if err := h.push.RetryDocument(c.Request.Context(), id, h.internalToken); err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}
	httpx.OK(c, gin.H{"id": id, "status": "ok"})
}

// AutoSyncStatus — GET /api/v1/sync/autosync
func (h *SyncHandler) AutoSyncStatus(c *gin.Context) {
	if h.autoSync == nil {
		httpx.OK(c, gin.H{"enabled": false, "push_enabled": false})
		return
	}
	httpx.OK(c, h.autoSync.Status())
}

// List — GET /api/v1/sync/jobs?limit=50
func (h *SyncHandler) List(c *gin.Context) {
	limit := 50
	if v := c.Query("limit"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
	}
	jobs, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}
	httpx.OK(c, gin.H{"items": jobs})
}

// Get — GET /api/v1/sync/jobs/:id
func (h *SyncHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{"code": "bad_request", "message": "invalid id"}})
		return
	}
	j, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": gin.H{"code": "internal", "message": err.Error()}})
		return
	}
	if j == nil {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "job not found"}})
		return
	}
	httpx.OK(c, j)
}

func (h *SyncHandler) run(j *repository.SyncJob) {
	defer func() {
		if r := recover(); r != nil {
			h.finish(j, "error", fmt.Sprintf("panic: %v", r))
		}
	}()

	ctx := context.Background()

	if err := h.repo.MarkRunning(ctx, j.ID); err != nil {
		h.finish(j, "error", fmt.Sprintf("mark running: %v", err))
		return
	}

	scriptPath := filepath.Join(h.scriptsDir, "sync-pull-orders.ps1")

	runCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	shell := "pwsh"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "powershell.exe"
	}
	cmd := exec.CommandContext(runCtx, shell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		h.finish(j, "error", fmt.Sprintf("stdout pipe: %v", err))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		h.finish(j, "error", fmt.Sprintf("stderr pipe: %v", err))
		return
	}

	if err := cmd.Start(); err != nil {
		h.finish(j, "error", fmt.Sprintf("start: %v", err))
		return
	}

	var stderrWG sync.WaitGroup
	stderrWG.Add(1)
	go func() {
		defer stderrWG.Done()
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			h.setLastLine(j, sc.Text())
		}
	}()

	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)

	gotDone := false
	var failMsg string
	for sc.Scan() {
		line := sc.Text()
		h.setLastLine(j, line)

		if m := progressRe.FindStringSubmatch(line); m != nil {
			var fetched, total int
			fmt.Sscanf(m[1], "%d", &fetched)
			fmt.Sscanf(m[2], "%d", &total)
			_ = h.repo.SetProgress(ctx, j.ID, fetched, total)
		}
		if doneRe.MatchString(line) {
			gotDone = true
		}
		if m := failRe.FindStringSubmatch(line); m != nil {
			failMsg = strings.TrimSpace(m[1])
		}
	}

	waitErr := cmd.Wait()
	stderrWG.Wait()

	if waitErr != nil {
		msg := waitErr.Error()
		if failMsg != "" {
			msg = failMsg
		}
		h.finish(j, "error", msg)
		return
	}
	if !gotDone {
		h.finish(j, "error", "скрипт завершился без [DONE] ok")
		return
	}
	h.finish(j, "done", "")
}

// setLastLine — с троттлингом: не чаще 300 мс на job, чтобы не долбить БД.
func (h *SyncHandler) setLastLine(j *repository.SyncJob, line string) {
	h.mu.Lock()
	last := h.lastLineAt[j.ID]
	now := time.Now()
	if now.Sub(last) < 300*time.Millisecond {
		h.mu.Unlock()
		return
	}
	h.lastLineAt[j.ID] = now
	h.mu.Unlock()

	_ = h.repo.SetLastLine(context.Background(), j.ID, line)
}

func (h *SyncHandler) finish(j *repository.SyncJob, status, errMsg string) {
	_ = h.repo.Finish(context.Background(), j.ID, status, errMsg)
}
