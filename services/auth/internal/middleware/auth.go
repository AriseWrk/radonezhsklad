package middleware

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"

"github.com/radonezhsklad/auth/internal/service"
)

const (
CtxUserID = "user_id"
CtxRole   = "role"
)

func RequireAuth(svc *service.AuthService) gin.HandlerFunc {
return func(c *gin.Context) {
header := c.GetHeader("Authorization")
if header == "" || !strings.HasPrefix(header, "Bearer ") {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
return
}
token := strings.TrimPrefix(header, "Bearer ")
claims, err := svc.ParseAccessToken(token)
if err != nil {
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
return
}
c.Set(CtxUserID, claims.UserID.String())
c.Set(CtxRole, claims.Role)
c.Next()
}
}

func RequireRole(roles ...string) gin.HandlerFunc {
allowed := map[string]bool{}
for _, r := range roles { allowed[r] = true }
return func(c *gin.Context) {
role, _ := c.Get(CtxRole)
r, _ := role.(string)
if !allowed[r] {
c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
return
}
c.Next()
}
}