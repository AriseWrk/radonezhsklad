package middleware

import (
"strings"

"github.com/gin-gonic/gin"
"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"

apperr "github.com/radonezhsklad/shared/errors"
)

const (
CtxUserID = "user_id"
CtxRole   = "role"
)

func RequireJWT(secret string) gin.HandlerFunc {
return func(c *gin.Context) {
header := c.GetHeader("Authorization")
if header == "" || !strings.HasPrefix(header, "Bearer ") {
c.Error(apperr.Unauthorized("missing bearer token"))
c.Abort()
return
}
raw := strings.TrimPrefix(header, "Bearer ")

token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, jwt.ErrSignatureInvalid
}
return []byte(secret), nil
})
if err != nil || !token.Valid {
c.Error(apperr.Unauthorized("invalid token"))
c.Abort()
return
}

claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
c.Error(apperr.Unauthorized("invalid token claims"))
c.Abort()
return
}
sub, _ := claims["sub"].(string)
uid, err := uuid.Parse(sub)
if err != nil {
c.Error(apperr.Unauthorized("invalid token subject"))
c.Abort()
return
}
role, _ := claims["role"].(string)

c.Set(CtxUserID, uid.String())
c.Set(CtxRole, role)
c.Next()
}
}

func CurrentUserID(c *gin.Context) uuid.UUID {
s := c.GetString(CtxUserID)
id, _ := uuid.Parse(s)
return id
}

func CurrentRole(c *gin.Context) string {
return c.GetString(CtxRole)
}

func RequireRole(roles ...string) gin.HandlerFunc {
allow := map[string]bool{}
for _, r := range roles {
allow[r] = true
}
return func(c *gin.Context) {
role := CurrentRole(c)
if !allow[role] {
c.Error(apperr.Forbidden("insufficient permissions"))
c.Abort()
return
}
c.Next()
}
}