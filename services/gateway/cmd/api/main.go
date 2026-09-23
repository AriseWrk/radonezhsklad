package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	auditLogger := audit.NewLogger(cfg.AuditURL, cfg.InternalToken)

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	 r.GET("/api/v1/health/all", func(c *gin.Context) {
		 type target struct{ name, url string }
		 targets := []target{
			 {"gateway", ""},
			 {"auth", cfg.AuthURL},
			 {"product", cfg.ProductURL},
			 {"warehouse", cfg.WarehouseURL},
			 {"order", cfg.OrderURL},
			 {"audit", cfg.AuditURL},
		 }
		 type result struct {
			 Name      string `json:"name"`
			 Status    string `json:"status"`
			 LatencyMs int64  `json:"latency_ms"`
		 }
		 out := make([]result, len(targets))
		 var wg sync.WaitGroup
		 client := &http.Client{Timeout: 1500 * time.Millisecond}
		 for i, t := range targets {
			 wg.Add(1)
			 go func(i int, name, url string) {
				 defer wg.Done()
				 out[i].Name = name
				 if url == "" {
					 out[i].Status = "ok"
					 return
				 }
				 start := time.Now()
				 resp, err := client.Get(url + "/api/v1/health")
				 out[i].LatencyMs = time.Since(start).Milliseconds()
				 if err != nil {
					 out[i].Status = "down"
					 return
				 }
				 defer resp.Body.Close()
				 if resp.StatusCode == 200 {
					 out[i].Status = "ok"
				 } else {
					 out[i].Status = "degraded"
				 }
			 }(i, t.name, t.url)
		 }
		 wg.Wait()
		 c.JSON(http.StatusOK, gin.H{"services": out})
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
		api.Any("/inventories", warehouseProxy)
		api.Any("/inventories/*path", warehouseProxy)
		api.Any("/suppliers", warehouseProxy)
		api.Any("/suppliers/*path", warehouseProxy)
		api.Any("/organizations", warehouseProxy)
		api.Any("/organizations/*path", warehouseProxy)
		api.Any("/projects", warehouseProxy)
		api.Any("/projects/*path", warehouseProxy)
		api.Any("/internal-orders", warehouseProxy)
		api.Any("/internal-orders/*path", warehouseProxy)
		api.Any("/sync", warehouseProxy)
		api.Any("/sync/*path", warehouseProxy)

		api.Any("/customers", orderProxy)
		api.Any("/customers/*path", orderProxy)
		api.Any("/contracts", orderProxy)
		api.Any("/contracts/*path", orderProxy)
		api.Any("/orders", orderProxy)
		api.Any("/orders/*path", orderProxy)
		api.Any("/analytics", orderProxy)
		api.Any("/analytics/*path", orderProxy)

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
			slog.Error("listen", "error", err)
			os.Exit(1)
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
