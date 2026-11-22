package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	mw "github.com/radonezhsklad/shared/middleware"
	"github.com/radonezhsklad/warehouse/internal/exporter"
	"github.com/radonezhsklad/warehouse/internal/reconciliation"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type ReconciliationHandler struct {
	service *reconciliation.Service
	runs    *repository.ReconciliationRepo
}

func NewReconciliationHandler(
	service *reconciliation.Service,
	runs *repository.ReconciliationRepo,
) *ReconciliationHandler {
	return &ReconciliationHandler{
		service: service,
		runs:    runs,
	}
}

func (h *ReconciliationHandler) ListRuns(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	runs, err := h.runs.ListRuns(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": runs})
}

func (h *ReconciliationHandler) GetActiveRun(c *gin.Context) {
	run, err := h.runs.GetActiveRun(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": run})
}

func (h *ReconciliationHandler) GetRun(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "bad_request", "message": "некорректный ID запуска"},
		})
		return
	}

	run, err := h.runs.GetRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "not_found", "message": "запуск не найден"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": run})
}

func (h *ReconciliationHandler) ListItems(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "bad_request", "message": "некорректный ID запуска"},
		})
		return
	}

	run, err := h.runs.GetRun(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "not_found", "message": "запуск не найден"},
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.runs.ListItems(
		c.Request.Context(),
		id,
		c.Query("type"),
		c.Query("status"),
		c.Query("search"),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *ReconciliationHandler) Start(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"code":    "unavailable",
				"message": "клиент МойСклад не настроен",
			},
		})
		return
	}

	active, err := h.runs.GetActiveRun(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}
	if active != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "conflict",
				"message": "сверка уже выполняется",
				"run_id":  active.ID,
			},
		})
		return
	}

	userID := mw.CurrentUserID(c)
	userName := c.GetString("user_name")
	if userName == "" {
		userName = "Пользователь"
	}

	run, err := h.runs.CreateRun(c.Request.Context(), userID, userName)
	if err != nil {
		active, activeErr := h.runs.GetActiveRun(c.Request.Context())
		if activeErr == nil && active != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "conflict",
					"message": "сверка уже выполняется",
					"run_id":  active.ID,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}

	go func() {
		_ = h.service.Run(context.Background(), run.ID)
	}()

	c.JSON(http.StatusCreated, gin.H{"item": run})
}

// Export выгружает все строки завершённой сверки в Excel.
func (h *ReconciliationHandler) Export(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "bad_request", "message": "некорректный ID запуска"},
		})
		return
	}

	ctx := c.Request.Context()
	run, err := h.runs.GetRun(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "internal", "message": err.Error()},
		})
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "not_found", "message": "запуск не найден"},
		})
		return
	}
	if run.Status == "running" {
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "conflict", "message": "нельзя экспортировать выполняющуюся сверку"},
		})
		return
	}

	items := make([]repository.ReconciliationItem, 0)
	offset := 0

	for {
		batch, total, listErr := h.runs.ListItems(ctx, id, "", "", "", 500, offset)
		if listErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "internal", "message": listErr.Error()},
			})
			return
		}
		if len(batch) == 0 {
			break
		}

		items = append(items, batch...)
		offset += len(batch)
		if offset >= total {
			break
		}
	}

	data, err := exporter.ReconciliationXLSX(run, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "export_failed", "message": err.Error()},
		})
		return
	}

	filename := fmt.Sprintf("reconciliation-%s.xlsx", run.ID.String())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data,
	)
}
