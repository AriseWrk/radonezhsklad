// Package mspull — pull справочников из МойСклад в нашу БД.
//
// Каждая функция — отдельный kind в sync_jobs (pull-counterparties,
// pull-warehouses, pull-organizations, pull-projects). Прогресс виден в UI
// через GET /sync/jobs/:id.
//
// Курсор: v_sync_cursor.last_cursor по kind. Читаем от него (или с -1 года,
// если курсора нет). Фильтр МС: updated>=<cursor>. После успешного прогона
// пишем cursor_after = now()-5s (перекрытие на случай гонок).
package mspull

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/radonezhsklad/shared/msapi"
	"github.com/radonezhsklad/warehouse/internal/product"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

const (
	pageSize      = 1000
	initialWindow = 365 * 24 * time.Hour // если курсора нет — за год
	overlap       = 5 * time.Second      // перекрытие курсора
	msMomentFmt   = "2006-01-02 15:04:05"
)

// Runner — общий контекст для pull'ов.
type Runner struct {
	ms    *msapi.Client
	repo  *repository.MsPullRepo
	jobs  *repository.SyncJobRepo
	pc    *product.Client
	token string
}

func NewRunner(ms *msapi.Client, repo *repository.MsPullRepo, jobs *repository.SyncJobRepo, pc *product.Client, token string) *Runner {
	return &Runner{ms: ms, repo: repo, jobs: jobs, pc: pc, token: token}
}

func (r *Runner) Enabled() bool { return r.ms != nil }

// ---------- MS-ответы ----------

type msListMeta struct {
	Size   int `json:"size"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type msListResponse[T any] struct {
	Meta msListMeta `json:"meta"`
	Rows []T        `json:"rows"`
}

// ---------- MS-строки ----------

type msRowCounterparty struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ExternalCode  *string   `json:"externalCode"`
	INN           *string   `json:"inn"`
	KPP           *string   `json:"kpp"`
	OGRN          *string   `json:"ogrn"`
	OKPO          *string   `json:"okpo"`
	CompanyType   *string   `json:"companyType"`
	LegalAddress  *string   `json:"legalAddress"`
	ActualAddress *string   `json:"actualAddress"`
	Fax           *string   `json:"fax"`
	Description   *string   `json:"description"`
	Archived      bool      `json:"archived"`
	Updated       *string   `json:"updated"`
}

type msRowStore struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ExternalCode *string   `json:"externalCode"`
	Address      *string   `json:"address"`
	Archived     bool      `json:"archived"`
	Updated      *string   `json:"updated"`
}

type msRowOrganization struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ExternalCode *string   `json:"externalCode"`
	KPP          *string   `json:"kpp"`
	OGRN         *string   `json:"ogrn"`
	OKPO         *string   `json:"okpo"`
	LegalAddress *string   `json:"legalAddress"`
	Email        *string   `json:"email"`
	Archived     bool      `json:"archived"`
	Updated      *string   `json:"updated"`
}

type msRowProject struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Archived bool      `json:"archived"`
	Updated  *string   `json:"updated"`
}

// ---------- Курсор ----------

func (r *Runner) cursorFor(ctx context.Context, kind string) string {
	c, err := r.jobs.GetLastCursor(ctx, kind)
	if err != nil {
		slog.Warn("mspull: cursor read failed", "kind", kind, "error", err)
	}
	if c != "" {
		return c
	}
	return time.Now().UTC().Add(-initialWindow).In(mskZone()).Format(msMomentFmt)
}

func mskZone() *time.Location {
	return time.FixedZone("MSK", 3*60*60)
}

// afterSuccess пишет cursor_after = now() - overlap.
func (r *Runner) afterSuccess(ctx context.Context, jobID uuid.UUID) {
	cursor := time.Now().In(mskZone()).Add(-overlap).Format(msMomentFmt)
	if err := r.jobs.SetCursorAfter(ctx, jobID, cursor); err != nil {
		slog.Warn("mspull: SetCursorAfter failed", "job_id", jobID, "error", err)
	}
}

// parseMSMoment — "2024-01-15 12:30:45.123" → time.Time.
func parseMSMoment(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05.000", "2006-01-02 15:04:05"} {
		if t, err := time.ParseInLocation(layout, *s, mskZone()); err == nil {
			return &t
		}
	}
	return nil
}

// setProgress пишет fetched/total и last_line.
func (r *Runner) setProgress(ctx context.Context, jobID uuid.UUID, fetched, total int) {
	if err := r.jobs.SetProgress(ctx, jobID, fetched, total); err != nil {
		slog.Warn("mspull: SetProgress failed", "job_id", jobID, "error", err)
	}
	_ = r.jobs.SetLastLine(ctx, jobID,
		fmt.Sprintf("[PROGRESS] fetched=%d total=%d", fetched, total))
}

// ---------- Counterparties ----------

func (r *Runner) PullCounterparties(ctx context.Context, jobID uuid.UUID) error {
	kind := "pull-counterparties"
	cursor := r.cursorFor(ctx, kind)
	slog.Info("mspull: counterparties", "since", cursor)

	fetched := 0
	total := 0
	offset := 0
	for {
		q := url.Values{}
		q.Set("limit", fmt.Sprint(pageSize))
		q.Set("offset", fmt.Sprint(offset))
		q.Set("filter", "updated>="+cursor)

		var page msListResponse[msRowCounterparty]
		if err := r.ms.Get(ctx, "/entity/counterparty", q, &page); err != nil {
			return fmt.Errorf("GET /entity/counterparty: %w", err)
		}
		if total == 0 {
			total = page.Meta.Size
		}
		if len(page.Rows) == 0 {
			break
		}
		for _, row := range page.Rows {
			cp := repository.MsCounterparty{
				ExternalID:    row.ID,
				Name:          row.Name,
				ExternalCode:  row.ExternalCode,
				INN:           row.INN,
				KPP:           row.KPP,
				OGRN:          row.OGRN,
				OKPO:          row.OKPO,
				CompanyType:   row.CompanyType,
				LegalAddress:  row.LegalAddress,
				ActualAddress: row.ActualAddress,
				Fax:           row.Fax,
				Description:   row.Description,
				Archived:      row.Archived,
				Updated:       parseMSMoment(row.Updated),
			}
			if err := r.repo.UpsertCounterparty(ctx, cp); err != nil {
				return fmt.Errorf("upsert counterparty %s: %w", row.ID, err)
			}
			fetched++
		}
		r.setProgress(ctx, jobID, fetched, total)
		offset += len(page.Rows)
		if offset >= page.Meta.Size {
			break
		}
	}
	r.afterSuccess(ctx, jobID)
	slog.Info("mspull: counterparties done", "fetched", fetched, "total", total)
	return nil
}

// ---------- Warehouses (stores) ----------

func (r *Runner) PullWarehouses(ctx context.Context, jobID uuid.UUID) error {
	kind := "pull-warehouses"
	cursor := r.cursorFor(ctx, kind)
	slog.Info("mspull: stores", "since", cursor)

	fetched := 0
	total := 0
	offset := 0
	for {
		q := url.Values{}
		q.Set("limit", fmt.Sprint(pageSize))
		q.Set("offset", fmt.Sprint(offset))
		q.Set("filter", "updated>="+cursor)

		var page msListResponse[msRowStore]
		if err := r.ms.Get(ctx, "/entity/store", q, &page); err != nil {
			return fmt.Errorf("GET /entity/store: %w", err)
		}
		if total == 0 {
			total = page.Meta.Size
		}
		if len(page.Rows) == 0 {
			break
		}
		for _, row := range page.Rows {
			ws := repository.MsStore{
				ExternalID:   row.ID,
				Name:         row.Name,
				ExternalCode: row.ExternalCode,
				Address:      row.Address,
				Archived:     row.Archived,
				Updated:      parseMSMoment(row.Updated),
			}
			if err := r.repo.UpsertWarehouse(ctx, ws); err != nil {
				return fmt.Errorf("upsert store %s: %w", row.ID, err)
			}
			fetched++
		}
		r.setProgress(ctx, jobID, fetched, total)
		offset += len(page.Rows)
		if offset >= page.Meta.Size {
			break
		}
	}
	r.afterSuccess(ctx, jobID)
	slog.Info("mspull: stores done", "fetched", fetched, "total", total)
	return nil
}

// ---------- Organizations ----------

func (r *Runner) PullOrganizations(ctx context.Context, jobID uuid.UUID) error {
	kind := "pull-organizations"
	cursor := r.cursorFor(ctx, kind)
	slog.Info("mspull: organizations", "since", cursor)

	fetched := 0
	total := 0
	offset := 0
	for {
		q := url.Values{}
		q.Set("limit", fmt.Sprint(pageSize))
		q.Set("offset", fmt.Sprint(offset))
		q.Set("filter", "updated>="+cursor)

		var page msListResponse[msRowOrganization]
		if err := r.ms.Get(ctx, "/entity/organization", q, &page); err != nil {
			return fmt.Errorf("GET /entity/organization: %w", err)
		}
		if total == 0 {
			total = page.Meta.Size
		}
		if len(page.Rows) == 0 {
			break
		}
		for _, row := range page.Rows {
			org := repository.MsOrganization{
				ExternalID:   row.ID,
				Name:         row.Name,
				ExternalCode: row.ExternalCode,
				KPP:          row.KPP,
				OGRN:         row.OGRN,
				OKPO:         row.OKPO,
				LegalAddress: row.LegalAddress,
				Email:        row.Email,
				Archived:     row.Archived,
				Updated:      parseMSMoment(row.Updated),
			}
			if err := r.repo.UpsertOrganization(ctx, org); err != nil {
				return fmt.Errorf("upsert organization %s: %w", row.ID, err)
			}
			fetched++
		}
		r.setProgress(ctx, jobID, fetched, total)
		offset += len(page.Rows)
		if offset >= page.Meta.Size {
			break
		}
	}
	r.afterSuccess(ctx, jobID)
	slog.Info("mspull: organizations done", "fetched", fetched, "total", total)
	return nil
}

// ---------- Projects ----------

func (r *Runner) PullProjects(ctx context.Context, jobID uuid.UUID) error {
	kind := "pull-projects"
	cursor := r.cursorFor(ctx, kind)
	slog.Info("mspull: projects", "since", cursor)

	fetched := 0
	total := 0
	offset := 0
	for {
		q := url.Values{}
		q.Set("limit", fmt.Sprint(pageSize))
		q.Set("offset", fmt.Sprint(offset))
		q.Set("filter", "updated>="+cursor)

		var page msListResponse[msRowProject]
		if err := r.ms.Get(ctx, "/entity/project", q, &page); err != nil {
			return fmt.Errorf("GET /entity/project: %w", err)
		}
		if total == 0 {
			total = page.Meta.Size
		}
		if len(page.Rows) == 0 {
			break
		}
		for _, row := range page.Rows {
			pr := repository.MsProject{
				ExternalID: row.ID,
				Name:       row.Name,
				Archived:   row.Archived,
				Updated:    parseMSMoment(row.Updated),
			}
			if err := r.repo.UpsertProject(ctx, pr); err != nil {
				return fmt.Errorf("upsert project %s: %w", row.ID, err)
			}
			fetched++
		}
		r.setProgress(ctx, jobID, fetched, total)
		offset += len(page.Rows)
		if offset >= page.Meta.Size {
			break
		}
	}
	r.afterSuccess(ctx, jobID)
	slog.Info("mspull: projects done", "fetched", fetched, "total", total)
	return nil
}

// ---------- Dispatcher ----------

// PullByKind — вход для handler'а. Возвращает ошибку для неизвестного kind.
func (r *Runner) PullByKind(ctx context.Context, kind string, jobID uuid.UUID) error {
	switch strings.ToLower(kind) {
	case "counterparties", "counterparty", "suppliers":
		return r.PullCounterparties(ctx, jobID)
	case "warehouses", "stores", "warehouse":
		return r.PullWarehouses(ctx, jobID)
	case "organizations", "organization":
		return r.PullOrganizations(ctx, jobID)
	case "projects", "project":
		return r.PullProjects(ctx, jobID)
	case "docs-enter", "enter", "receipts":
		return r.PullEnter(ctx, jobID)
	case "docs-demand", "demand", "shipments":
		return r.PullDemand(ctx, jobID)
	case "docs-loss", "loss", "writeoffs":
		return r.PullLoss(ctx, jobID)
	case "docs-move", "move", "transfers":
		return r.PullMove(ctx, jobID)
	default:
		return fmt.Errorf("неизвестный kind для pull: %q", kind)
	}
}
