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
if err != nil { slog.Error("db connect", "error", err); os.Exit(1) }
defer pool.Close()
slog.Info("db connected")

repo := repository.New(pool)
svc := service.New(repo)
h := handler.New(svc)

if env == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(mw.RequestID(), mw.Recovery(), mw.AccessLog(), mw.CORS(cfg.CORSOrigins), mw.ErrorHandler())

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "product"})
})

read := api.Group("")
read.Use(mw.RequireJWT(cfg.JWTSecret))
{
read.GET("/units", h.ListUnits)
read.GET("/categories", h.ListCategories)
read.GET("/categories/:id", h.GetCategory)
read.GET("/products", h.ListProducts)
read.GET("/products/list", h.ListByIDs)
read.GET("/products/:id", h.GetProduct)
}

write := api.Group("")
write.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "manager"))
{
write.POST("/categories", h.CreateCategory)
write.PUT("/categories/:id", h.UpdateCategory)
write.DELETE("/categories/:id", h.DeleteCategory)
write.POST("/products", h.CreateProduct)
write.PUT("/products/:id", h.UpdateProduct)
write.DELETE("/products/:id", h.ArchiveProduct)
}
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
go func() {
slog.Info("product service listening", "port", cfg.Port)
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
slog.Info("product service stopped")
}