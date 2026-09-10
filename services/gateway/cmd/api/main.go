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

if env == "prod" {
gin.SetMode(gin.ReleaseMode)
}
r := gin.New()
r.Use(
mw.RequestID(),
mw.Recovery(),
mw.AccessLog(),
mw.CORS(cfg.CORSOrigins),
)

// health самого gateway
r.GET("/api/v1/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
})

// прокси-маршруты
authProxy := proxy.New(cfg.AuthURL)
productProxy := proxy.New(cfg.ProductURL)
warehouseProxy := proxy.New(cfg.WarehouseURL)
orderProxy := proxy.New(cfg.OrderURL)

api := r.Group("/api/v1")
{
// auth — как есть
api.Any("/auth/*path", authProxy)

// product
api.Any("/units", productProxy)
api.Any("/units/*path", productProxy)
api.Any("/categories", productProxy)
api.Any("/categories/*path", productProxy)
api.Any("/products", productProxy)
api.Any("/products/*path", productProxy)

// warehouse (сервис появится на Этапе 11)
api.Any("/warehouse", warehouseProxy)
api.Any("/warehouse/*path", warehouseProxy)

// order (сервис появится на Этапе 12)
api.Any("/orders", orderProxy)
api.Any("/orders/*path", orderProxy)
}

srv := &http.Server{
Addr:              ":" + cfg.Port,
Handler:           r,
ReadHeaderTimeout: 10 * time.Second,
}

go func() {
slog.Info("gateway listening",
"port", cfg.Port,
"auth", cfg.AuthURL,
"product", cfg.ProductURL,
)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
slog.Error("listen", "error", err)
os.Exit(1)
}
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(shutdownCtx); err != nil {
slog.Error("shutdown", "error", err)
}
slog.Info("gateway stopped")
}