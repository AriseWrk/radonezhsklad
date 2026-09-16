package handler

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/order/internal/models"
apperr "github.com/radonezhsklad/shared/errors"
"github.com/radonezhsklad/shared/httpx"
)

type contractReq struct {
Number         string  `json:"number" binding:"required"`
ContractType   string  `json:"contract_type"`
Code           string  `json:"code"`
DocDate        string  `json:"doc_date"`
CustomerID     string  `json:"customer_id"`
OrganizationID string  `json:"organization_id"`
Amount         float64 `json:"amount"`
Currency       string  `json:"currency"`
Paid           float64 `json:"paid"`
Fulfilled      float64 `json:"fulfilled"`
Comment        string  `json:"comment"`
Archived       bool    `json:"archived"`
}

func (r contractReq) toModel() *models.Contract {
c := &models.Contract{
Number:       r.Number,
ContractType: r.ContractType,
Amount:       r.Amount,
Paid:         r.Paid,
Fulfilled:    r.Fulfilled,
Currency:     r.Currency,
Archived:     r.Archived,
}
if r.Code != "" { c.Code = &r.Code }
if r.Comment != "" { c.Comment = &r.Comment }
if r.CustomerID != "" {
if id, err := uuid.Parse(r.CustomerID); err == nil { c.CustomerID = &id }
}
if r.OrganizationID != "" {
if id, err := uuid.Parse(r.OrganizationID); err == nil { c.OrganizationID = &id }
}
if r.DocDate != "" {
if t, err := time.Parse(time.RFC3339, r.DocDate); err == nil {
c.DocDate = t
} else if t, err := time.Parse("2006-01-02T15:04:05", r.DocDate); err == nil {
c.DocDate = t
} else if t, err := time.Parse("2006-01-02", r.DocDate); err == nil {
c.DocDate = t
}
}
return c
}

func (h *Handler) CreateContract(c *gin.Context) {
var req contractReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
item, err := h.contractSvc.Create(c.Request.Context(), req.toModel())
if err != nil { c.Error(err); return }
httpx.Created(c, item)
}

func (h *Handler) ListContracts(c *gin.Context) {
includeArchived := c.Query("include_archived") == "true"
items, err := h.contractSvc.List(c.Request.Context(), includeArchived)
if err != nil { c.Error(err); return }
httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) GetContract(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
item, err := h.contractSvc.Get(c.Request.Context(), id)
if err != nil { c.Error(err); return }
httpx.OK(c, item)
}

func (h *Handler) UpdateContract(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req contractReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
item, err := h.contractSvc.Update(c.Request.Context(), id, req.toModel())
if err != nil { c.Error(err); return }
httpx.OK(c, item)
}

func (h *Handler) DeleteContract(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
if err := h.contractSvc.Delete(c.Request.Context(), id); err != nil { c.Error(err); return }
c.Status(http.StatusNoContent)
}

func (h *Handler) ContractNeighbors(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }

cur, err := h.contractSvc.Get(c.Request.Context(), id)
if err != nil { c.Error(err); return }

nb, err := h.contractSvc.Neighbors(c.Request.Context(), id)
if err != nil { c.Error(err); return }

next, _ := h.contractSvc.GetNext(c.Request.Context(), cur.DocDate, id)
prev, _ := h.contractSvc.GetPrev(c.Request.Context(), cur.DocDate, id)

var nextID, prevID string
if next != nil { nextID = next.ID.String() }
if prev != nil { prevID = prev.ID.String() }

httpx.OK(c, gin.H{
"index":   nb.Index,
"total":   nb.Total,
"next_id": nextID,
"prev_id": prevID,
})
}