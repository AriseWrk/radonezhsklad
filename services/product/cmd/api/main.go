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

"github.com/radonezhsklad/product/internal/config"
"github.com/radonezhsklad/product/internal/db"
"github.com/radonezhsklad/product/internal/handler"
"github.com/radonezhsklad/product/internal/repository"
"github.com/radonezhsklad/product/internal/service"
shcfg "github.com/radonezhsklad/shared/config"
"github.com/radonezhsklad/shared/logger"
mw "github.com/radonezhsklad/shared/middleware"
)

func main() {
_ = godotenv.Load()

env := shcfg.GetString("ENV", "dev")
logger.Init(env)

cfg := config.Load()
ctx := context.Background()

pool, err := db.Connect(ctx, cfg.DatabaseURL)
if err != nil {
slog.Error("db connect", "error", err)
os.Exit(1)
}
defer pool.Close()
slog.Info("db connected")

repo := repository.New(pool)
svc := service.New(repo)
h := handler.New(svc)

if env == "prod" {
gin.SetMode(gin.ReleaseMode)
}
r := gin.New()
r.Use(
mw.RequestID(),
mw.Recovery(),
mw.AccessLog(),
mw.CORS(cfg.CORSOrigins),
mw.ErrorHandler(),
)

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "product"})
})

authorized := api.Group("")
authorized.Use(mw.RequireJWT(cfg.JWTSecret))
{
authorized.GET("/units", h.ListUnits)

authorized.POST("/categories", h.CreateCategory)
authorized.GET("/categories", h.ListCategories)
authorized.GET("/categories/:id", h.GetCategory)
authorized.PUT("/categories/:id", h.UpdateCategory)
authorized.DELETE("/categories/:id", h.DeleteCategory)

authorized.POST("/products", h.CreateProduct)
authorized.GET("/products", h.ListProducts)
authorized.GET("/products/:id", h.GetProduct)
authorized.PUT("/products/:id", h.UpdateProduct)
authorized.DELETE("/products/:id", h.ArchiveProduct)
}
}

srv := &http.Server{
Addr:              ":" + cfg.Port,
Handler:           r,
ReadHeaderTimeout: 5 * time.Second,
}

go func() {
slog.Info("product service listening", "port", cfg.Port)
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
slog.Info("product service stopped")
}