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

"github.com/radonezhsklad/audit/internal/config"
"github.com/radonezhsklad/audit/internal/db"
"github.com/radonezhsklad/audit/internal/handler"
"github.com/radonezhsklad/audit/internal/repository"
)

func main() {
_ = godotenv.Load()
env := shcfg.GetString("ENV", "dev")
logger.Init(env)
cfg := config.Load()
ctx := context.Background()

pool, err := db.Connect(ctx, cfg.DatabaseURL)
if err != nil { slog.Error("db connect", "error", err); os.Exit(1) }
defer pool.Close()
slog.Info("db connected")

repo := repository.New(pool)
h := handler.New(repo, cfg.InternalToken)

if env == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(mw.RequestID(), mw.Recovery(), mw.AccessLog(), mw.CORS(cfg.CORSOrigins), mw.ErrorHandler())

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "audit"})
})

// внутренний приём от gateway
api.POST("/internal/audit", h.InternalLog)

// список — только admin
read := api.Group("/audit")
read.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin"))
{
read.GET("", h.List)
}
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
go func() {
slog.Info("audit service listening", "port", cfg.Port)
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
slog.Info("audit service stopped")
}