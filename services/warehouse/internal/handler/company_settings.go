package handler

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/repository"
)

type CompanySettingsHandler struct {
repo *repository.CompanySettingsRepo
}

func NewCompanySettingsHandler(repo *repository.CompanySettingsRepo) *CompanySettingsHandler {
return &CompanySettingsHandler{repo: repo}
}

func (h *CompanySettingsHandler) Get(c *gin.Context) {
s, err := h.repo.Get(c.Request.Context())
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, s)
}

type updateCompanySettingsReq struct {
Numbering         string `json:"numbering"`
ReplyAddress      string `json:"reply_address"`
ForbidShipNoStock bool   `json:"forbid_ship_no_stock"`
AutoMinPrice      bool   `json:"auto_min_price"`
MoveDeletedToBin  bool   `json:"move_deleted_to_bin"`
UseLots           bool   `json:"use_lots"`
ShowExtraFields   bool   `json:"show_extra_fields"`
}

func (h *CompanySettingsHandler) Update(c *gin.Context) {
var req updateCompanySettingsReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
return
}

var userID uuid.UUID
if s := c.GetString("user_id"); s != "" {
userID, _ = uuid.Parse(s)
}
userName := c.GetString("user_name")
if userName == "" {
userName = "Администратор"
}

err := h.repo.Update(c.Request.Context(), repository.CompanySettings{
Numbering:         req.Numbering,
ReplyAddress:      req.ReplyAddress,
ForbidShipNoStock: req.ForbidShipNoStock,
AutoMinPrice:      req.AutoMinPrice,
MoveDeletedToBin:  req.MoveDeletedToBin,
UseLots:           req.UseLots,
ShowExtraFields:   req.ShowExtraFields,
}, userID, userName)

if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"status": "ok"})
}