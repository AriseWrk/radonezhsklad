package handler

import (
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/radonezhsklad/warehouse/internal/product"
    "github.com/radonezhsklad/warehouse/internal/repository"
    "github.com/radonezhsklad/warehouse/internal/service"
    apperr "github.com/radonezhsklad/shared/errors"
    "github.com/radonezhsklad/shared/httpx"
)

type InventoryHandler struct {
    svc        *service.InventoryService
    productCli *product.Client
}

func NewInventoryHandler(svc *service.InventoryService, pc *product.Client) *InventoryHandler {
    return &InventoryHandler{svc: svc, productCli: pc}
}

func (h *InventoryHandler) List(c *gin.Context) {
    f := repository.InventoryFilters{}
    f.WarehouseID = parseUUIDOpt(c.Query("warehouse_id"))
    if v := c.Query("from"); v != "" {
        if t, err := time.Parse("2006-01-02", v); err == nil { f.From = &t }
    }
    if v := c.Query("to"); v != "" {
        if t, err := time.Parse("2006-01-02", v); err == nil { f.To = &t }
    }
    items, err := h.svc.List(c.Request.Context(), f)
    if err != nil { c.Error(err); return }
    httpx.OK(c, gin.H{"items": items})
}

func (h *InventoryHandler) Get(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
    inv, err := h.svc.Get(c.Request.Context(), id)
    if err != nil { c.Error(err); return }

    if len(inv.Items) > 0 {
        token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        ids := make([]uuid.UUID, 0, len(inv.Items))
        seen := map[uuid.UUID]bool{}
        for _, it := range inv.Items {
            if !seen[it.ProductID] {
                seen[it.ProductID] = true
                ids = append(ids, it.ProductID)
            }
        }
        prods, err := h.productCli.ListByIDs(c.Request.Context(), token, ids)
        if err == nil {
            pmap := make(map[uuid.UUID]product.Product, len(prods))
            for _, p := range prods { pmap[p.ID] = p }
            for i := range inv.Items {
                if p, ok := pmap[inv.Items[i].ProductID]; ok {
                    inv.Items[i].ProductName = p.Name
                    if p.SKU != nil { inv.Items[i].ProductSKU = *p.SKU }
                }
            }
        }
    }

    httpx.OK(c, inv)
}

func (h *InventoryHandler) Neighbors(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
    n, err := h.svc.Neighbors(c.Request.Context(), id)
    if err != nil { c.Error(err); return }
    httpx.OK(c, gin.H{
        "prev_id":  n.PrevID,
        "next_id":  n.NextID,
        "position": n.Position,
        "total":    n.Total,
    })
}


func (h *InventoryHandler) CreateCorrection(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
    var req struct {
        Kind string `json:"kind" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(apperr.BadRequest(err.Error())); return
    }
    doc, err := h.svc.CreateCorrection(c.Request.Context(), id, req.Kind)
    if err != nil { c.Error(err); return }
    httpx.Created(c, doc)
}
