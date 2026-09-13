package audit

import (
"strings"

"github.com/gin-gonic/gin"
"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"

mw "github.com/radonezhsklad/shared/middleware"
)

// UserContext парсит Bearer-токен (без отказа при ошибке) и кладёт user_id в контекст,
// чтобы audit-логи знали, кто сделал запрос.
func UserContext(secret string) gin.HandlerFunc {
return func(c *gin.Context) {
header := c.GetHeader("Authorization")
if strings.HasPrefix(header, "Bearer ") {
raw := strings.TrimPrefix(header, "Bearer ")
token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, jwt.ErrSignatureInvalid
}
return []byte(secret), nil
})
if err == nil && token.Valid {
if claims, ok := token.Claims.(jwt.MapClaims); ok {
if sub, ok := claims["sub"].(string); ok {
if id, err := uuid.Parse(sub); err == nil {
c.Set(mw.CtxUserID, id.String())
}
}
if role, ok := claims["role"].(string); ok {
c.Set(mw.CtxRole, role)
}
}
}
}
c.Next()
}
}