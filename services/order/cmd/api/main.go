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

"github.com/radonezhsklad/order/internal/config"
"github.com/radonezhsklad/order/internal/db"
"github.com/radonezhsklad/order/internal/handler"
"github.com/radonezhsklad/order/internal/repository"
"github.com/radonezhsklad/order/internal/service"
"github.com/radonezhsklad/order/internal/warehouse"
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
wh := warehouse.New(cfg.WarehouseURL)
svc := service.New(repo, wh)
h := handler.New(svc)

if env == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(mw.RequestID(), mw.Recovery(), mw.AccessLog(), mw.CORS(cfg.CORSOrigins), mw.ErrorHandler())

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "order"})
})

auth := api.Group("")
auth.Use(mw.RequireJWT(cfg.JWTSecret))
{
auth.GET("/customers", h.ListCustomers)
auth.POST("/customers", h.CreateCustomer)
auth.GET("/customers/:id", h.GetCustomer)
auth.DELETE("/customers/:id", h.DeleteCustomer)

auth.GET("/orders", h.ListOrders)
auth.POST("/orders", h.CreateOrder)
auth.GET("/orders/:id", h.GetOrder)
auth.POST("/orders/:id/confirm", h.ConfirmOrder)
auth.POST("/orders/:id/ship", h.ShipOrder)
auth.POST("/orders/:id/cancel", h.CancelOrder)
}
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
go func() {
slog.Info("order service listening", "port", cfg.Port, "warehouse", cfg.WarehouseURL)
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
slog.Info("order service stopped")
}