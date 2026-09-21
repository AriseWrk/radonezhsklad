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
)

// SyncJob — состояние одной задачи синхронизации.
type SyncJob struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"` // queued | running | done | error
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Fetched   int        `json:"fetched"`
	Total     int        `json:"total"`
	Error     string     `json:"error,omitempty"`
	LastLine  string     `json:"last_line,omitempty"`
}

// SyncHandler запускает PS-скрипты синхронизации и хранит статус в памяти.
type SyncHandler struct {
	mu         sync.Mutex
	jobs       map[string]*SyncJob
	activeID   string
	scriptsDir string
}

func NewSyncHandler(scriptsDir string) *SyncHandler {
	return &SyncHandler{jobs: map[string]*SyncJob{}, scriptsDir: scriptsDir}
}

var (
	progressRe = regexp.MustCompile(`\[PROGRESS\] fetched=(\d+) total=(\d+)`)
	doneRe     = regexp.MustCompile(`\[DONE\] ok`)
	failRe     = regexp.MustCompile(`\[FAIL\]\s*(.+)$`)
)

// Start — POST /api/v1/sync/pull-orders
func (h *SyncHandler) Start(c *gin.Context) {
	h.mu.Lock()
	if h.activeID != "" {
		if j, ok := h.jobs[h.activeID]; ok && (j.Status == "queued" || j.Status == "running") {
			h.mu.Unlock()
			c.JSON(409, gin.H{"error": gin.H{
				"code":    "conflict",
				"message": "синхронизация уже идёт",
				"job_id":  j.ID,
			}})
			return
		}
	}
	id := uuid.NewString()
	j := &SyncJob{ID: id, Status: "queued", StartedAt: time.Now()}
	h.jobs[id] = j
	h.activeID = id
	h.mu.Unlock()

	go h.run(j)

	httpx.Created(c, j)
}

// Get — GET /api/v1/sync/jobs/:id
func (h *SyncHandler) Get(c *gin.Context) {
	id := c.Param("id")
	h.mu.Lock()
	j, ok := h.jobs[id]
	h.mu.Unlock()
	if !ok {
		c.JSON(404, gin.H{"error": gin.H{"code": "not_found", "message": "job not found"}})
		return
	}
	httpx.OK(c, j)
}

func (h *SyncHandler) run(j *SyncJob) {
	defer func() {
		if r := recover(); r != nil {
			h.finish(j, "error", fmt.Sprintf("panic: %v", r))
		}
	}()

	h.mu.Lock()
	j.Status = "running"
	h.mu.Unlock()

	scriptPath := filepath.Join(h.scriptsDir, "sync-pull-orders.ps1")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// PowerShell 7 (pwsh) — ищем в PATH. Если нет — пробуем powershell.exe.
	shell := "pwsh"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "powershell.exe"
	}
	cmd := exec.CommandContext(ctx, shell, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)

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

	// stderr в фон (собираем в LastLine)
	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			h.setLastLine(j, sc.Text())
		}
	}()

	// stdout — основной поток прогресса
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
			h.mu.Lock()
			j.Fetched = fetched
			j.Total = total
			h.mu.Unlock()
		}
		if doneRe.MatchString(line) {
			gotDone = true
		}
		if m := failRe.FindStringSubmatch(line); m != nil {
			failMsg = strings.TrimSpace(m[1])
		}
	}

	waitErr := cmd.Wait()

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

func (h *SyncHandler) setLastLine(j *SyncJob, line string) {
	h.mu.Lock()
	j.LastLine = line
	h.mu.Unlock()
}

func (h *SyncHandler) finish(j *SyncJob, status, errMsg string) {
	h.mu.Lock()
	now := time.Now()
	j.Status = status
	j.EndedAt = &now
	if errMsg != "" {
		j.Error = errMsg
	}
	h.mu.Unlock()
}
