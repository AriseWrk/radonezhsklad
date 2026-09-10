package handler

import (
"errors"
"net/http"

"github.com/gin-gonic/gin"

"github.com/radonezhsklad/auth/internal/middleware"
"github.com/radonezhsklad/auth/internal/service"
)

type AuthHandler struct {
svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
return &AuthHandler{svc: svc}
}

type registerReq struct {
Email    string `json:"email"     binding:"required,email"`
Password string `json:"password"  binding:"required,min=8"`
FullName string `json:"full_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
var req registerReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
if errors.Is(err, service.ErrUserExists) {
c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"})
return
}
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
return
}
c.JSON(http.StatusCreated, u)
}

type loginReq struct {
Email    string `json:"email"    binding:"required"`
Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
var req loginReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

pair, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
if errors.Is(err, service.ErrInvalidCredentials) {
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
return
}
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
return
}
c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Me(c *gin.Context) {
userID, _ := c.Get(middleware.CtxUserID)
role, _ := c.Get(middleware.CtxRole)
c.JSON(http.StatusOK, gin.H{
"user_id": userID,
"role":    role,
})
}