package handler

import (
"errors"
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"

"github.com/radonezhsklad/auth/internal/middleware"
"github.com/radonezhsklad/auth/internal/service"
apperr "github.com/radonezhsklad/shared/errors"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

type registerReq struct {
Email    string `json:"email"     binding:"required,email"`
Password string `json:"password"  binding:"required,min=8"`
FullName string `json:"full_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
var req registerReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
}
u, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
if errors.Is(err, service.ErrUserExists) {
c.JSON(http.StatusConflict, gin.H{"error": "user with this email already exists"}); return
}
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"}); return }
c.JSON(http.StatusCreated, u)
}

type loginReq struct {
Email    string `json:"email"    binding:"required"`
Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
var req loginReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
}
pair, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
if errors.Is(err, service.ErrInvalidCredentials) {
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"}); return
}
if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"}); return }
c.JSON(http.StatusOK, pair)
}

func (h *AuthHandler) Me(c *gin.Context) {
userID, _ := c.Get(middleware.CtxUserID)
role, _ := c.Get(middleware.CtxRole)
c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
}

// ---------- admin ----------

type createUserReq struct {
Email       string `json:"email"        binding:"required,email"`
Password    string `json:"password"     binding:"required,min=8"`
LastName    string `json:"last_name"    binding:"required"`
FirstName   string `json:"first_name"   binding:"required"`
MiddleName  string `json:"middle_name"`
Phone       string `json:"phone"`
Login       string `json:"login"`
Description string `json:"description"`
Role        string `json:"role"         binding:"required"`
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
var req createUserReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
u, err := h.svc.CreateUserWithRole(c.Request.Context(), service.CreateUserInput{
Email: req.Email, Password: req.Password,
LastName: req.LastName, FirstName: req.FirstName, MiddleName: req.MiddleName,
Phone: req.Phone, Login: req.Login, Description: req.Description, Role: req.Role,
})
if errors.Is(err, service.ErrUserExists) { c.Error(apperr.Conflict("user already exists")); return }
if errors.Is(err, service.ErrInvalidRole) { c.Error(apperr.BadRequest("invalid role")); return }
if err != nil { c.Error(apperr.Internal("create user", err)); return }
c.JSON(http.StatusCreated, u)
}

func (h *AuthHandler) ListUsers(c *gin.Context) {
items, err := h.svc.ListUsers(c.Request.Context())
if err != nil { c.Error(apperr.Internal("list users", err)); return }
c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AuthHandler) UpdateUser(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req createUserReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
u, err := h.svc.UpdateUser(c.Request.Context(), id, service.CreateUserInput{
Email: req.Email, Password: req.Password,
LastName: req.LastName, FirstName: req.FirstName, MiddleName: req.MiddleName,
Phone: req.Phone, Login: req.Login, Description: req.Description, Role: req.Role,
})
if errors.Is(err, service.ErrInvalidRole) { c.Error(apperr.BadRequest("invalid role")); return }
if err != nil { c.Error(apperr.Internal("update user", err)); return }
c.JSON(http.StatusOK, u)
}

type updateRoleReq struct {
Role string `json:"role" binding:"required"`
}

func (h *AuthHandler) UpdateRole(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req updateRoleReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
if err := h.svc.UpdateRole(c.Request.Context(), id, req.Role); err != nil {
if errors.Is(err, service.ErrInvalidRole) { c.Error(apperr.BadRequest("invalid role")); return }
c.Error(apperr.Internal("update role", err)); return
}
c.Status(http.StatusNoContent)
}

type updateActiveReq struct {
IsActive bool `json:"is_active"`
}

func (h *AuthHandler) UpdateActive(c *gin.Context) {
id, err := uuid.Parse(c.Param("id"))
if err != nil { c.Error(apperr.BadRequest("invalid id")); return }
var req updateActiveReq
if err := c.ShouldBindJSON(&req); err != nil { c.Error(apperr.BadRequest(err.Error())); return }
if err := h.svc.UpdateActive(c.Request.Context(), id, req.IsActive); err != nil {
c.Error(apperr.Internal("update active", err)); return
}
c.Status(http.StatusNoContent)
}