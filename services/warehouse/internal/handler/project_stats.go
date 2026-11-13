package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/radonezhsklad/shared/httpx"
	"github.com/radonezhsklad/warehouse/internal/product"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type ProjectStatsHandler struct {
	repo       *repository.ProjectStatsRepo
	productCli *product.Client
}

func NewProjectStatsHandler(repo *repository.ProjectStatsRepo, productCli *product.Client) *ProjectStatsHandler {
	return &ProjectStatsHandler{repo: repo, productCli: productCli}
}

// ListOrders — GET /api/v1/projects/:id/internal-orders
func (h *ProjectStatsHandler) ListOrders(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	orders, err := h.repo.ListOrdersByProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpx.OK(c, gin.H{"items": orders})
}

// SystemsSummary — GET /api/v1/projects/:id/systems-summary
func (h *ProjectStatsHandler) SystemsSummary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	items, err := h.repo.SystemsSummaryByProject(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ids := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, it := range items {
		if _, ok := seen[it.ProductID]; ok {
			continue
		}
		seen[it.ProductID] = struct{}{}
		ids = append(ids, it.ProductID)
	}

	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	products, perr := h.productCli.ListByIDs(c.Request.Context(), token, ids)
	if perr != nil {
		c.Header("X-Product-Enrich-Warning", perr.Error())
	}

	pdata := make(map[uuid.UUID]product.Product, len(products))
	for _, p := range products {
		pdata[p.ID] = p
	}

	type enriched struct {
		repository.SystemItem
		Name   string     `json:"name"`
		SKU    *string    `json:"sku,omitempty"`
		UnitID *uuid.UUID `json:"unit_id,omitempty"`
	}
	out := make([]enriched, len(items))
	for i, it := range items {
		out[i].SystemItem = it
		if p, ok := pdata[it.ProductID]; ok {
			out[i].Name = p.Name
			out[i].SKU = p.SKU
			out[i].UnitID = p.UnitID
		}
	}

	httpx.OK(c, gin.H{"items": out})
}
