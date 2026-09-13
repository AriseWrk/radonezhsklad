package main

import (
"context"
"log/slog"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/gin-gonic/gin"
"github.com/joho/godotenv"

"github.com/radonezhsklad/gateway/internal/audit"
"github.com/radonezhsklad/gateway/internal/config"
"github.com/radonezhsklad/gateway/internal/proxy"
shcfg "github.com/radonezhsklad/shared/config"
"github.com/radonezhsklad/shared/logger"
mw "github.com/radonezhsklad/shared/middleware"
)

func main() {
_ = godotenv.Load()
env := shcfg.GetString("ENV", "dev")
logger.Init(env)
cfg := config.Load()

if env == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(
mw.RequestID(),
mw.Recovery(),
mw.AccessLog(),
mw.CORS(cfg.CORSOrigins),
)

auditLogger := audit.NewLogger(cfg.AuditURL, cfg.InternalToken)

r.GET("/api/v1/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
})

authProxy := proxy.New(cfg.AuthURL)
productProxy := proxy.New(cfg.ProductURL)
warehouseProxy := proxy.New(cfg.WarehouseURL)
orderProxy := proxy.New(cfg.OrderURL)
auditProxy := proxy.New(cfg.AuditURL)

api := r.Group("/api/v1")
api.Use(auditLogger.Middleware())
{
api.Any("/auth/*path", authProxy)
api.Any("/users", authProxy)
api.Any("/users/*path", authProxy)

api.Any("/units", productProxy)
api.Any("/units/*path", productProxy)
api.Any("/categories", productProxy)
api.Any("/categories/*path", productProxy)
api.Any("/products", productProxy)
api.Any("/products/*path", productProxy)

api.Any("/warehouses", warehouseProxy)
api.Any("/warehouses/*path", warehouseProxy)
api.Any("/stock", warehouseProxy)
api.Any("/stock/*path", warehouseProxy)
api.Any("/documents", warehouseProxy)
api.Any("/documents/*path", warehouseProxy)
api.Any("/inventory", warehouseProxy)
api.Any("/inventory/*path", warehouseProxy)

api.Any("/customers", orderProxy)
api.Any("/customers/*path", orderProxy)
api.Any("/orders", orderProxy)
api.Any("/orders/*path", orderProxy)

// audit list
api.Any("/audit", auditProxy)
api.Any("/audit/*path", auditProxy)
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 10 * time.Second}
go func() {
slog.Info("gateway listening",
"port", cfg.Port,
"auth", cfg.AuthURL, "product", cfg.ProductURL,
"warehouse", cfg.WarehouseURL, "order", cfg.OrderURL, "audit", cfg.AuditURL,
)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error("listen", "error", err); os.Exit(1)
}
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = srv.Shutdown(shutdownCtx)
slog.Info("gateway stopped")
}