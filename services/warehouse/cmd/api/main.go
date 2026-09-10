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

shcfg "github.com/radonezhsklad/shared/config"
"github.com/radonezhsklad/shared/logger"
mw "github.com/radonezhsklad/shared/middleware"
"github.com/radonezhsklad/warehouse/internal/config"
"github.com/radonezhsklad/warehouse/internal/db"
"github.com/radonezhsklad/warehouse/internal/handler"
"github.com/radonezhsklad/warehouse/internal/repository"
"github.com/radonezhsklad/warehouse/internal/service"
)

func main() {
_ = godotenv.Load()
env := shcfg.GetString("ENV", "dev")
logger.Init(env)

cfg := config.Load()
ctx := context.Background()

pool, err := db.Connect(ctx, cfg.DatabaseURL)
if err != nil {
slog.Error("db connect", "error", err); os.Exit(1)
}
defer pool.Close()
slog.Info("db connected")

repo := repository.New(pool)
svc := service.New(repo)
h := handler.New(svc)

if env == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(
mw.RequestID(), mw.Recovery(), mw.AccessLog(), mw.CORS(cfg.CORSOrigins), mw.ErrorHandler(),
)

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "warehouse"})
})

auth := api.Group("")
auth.Use(mw.RequireJWT(cfg.JWTSecret))
{
auth.GET("/warehouses", h.ListWarehouses)
auth.POST("/warehouses", h.CreateWarehouse)
auth.GET("/warehouses/:id", h.GetWarehouse)
auth.PUT("/warehouses/:id", h.UpdateWarehouse)
auth.DELETE("/warehouses/:id", h.DeleteWarehouse)

auth.GET("/stock", h.ListStock)

auth.GET("/documents", h.ListDocuments)
auth.POST("/documents", h.CreateDocument)
auth.GET("/documents/:id", h.GetDocument)
auth.POST("/documents/:id/post", h.PostDocument)
auth.POST("/documents/:id/cancel", h.CancelDocument)
}
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
go func() {
slog.Info("warehouse service listening", "port", cfg.Port)
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
slog.Info("warehouse service stopped")
}