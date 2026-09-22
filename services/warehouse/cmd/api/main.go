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
	"github.com/radonezhsklad/shared/msapi"
	"github.com/radonezhsklad/warehouse/internal/config"
	"github.com/radonezhsklad/warehouse/internal/db"
	"github.com/radonezhsklad/warehouse/internal/handler"
	"github.com/radonezhsklad/warehouse/internal/mspush"
	"github.com/radonezhsklad/warehouse/internal/product"
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
		slog.Error("db connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("db connected")

	repo := repository.New(pool)
	supRepo := repository.NewSupplierRepo(pool)
	intOrderRepo := repository.NewInternalOrderRepo(pool)
	projectRepo := repository.NewProjectRepo(pool)
	invRepo := repository.NewInventoryRepo(pool)
	orgRepo := repository.NewOrganizationRepo(pool)

	svc := service.New(repo)
	supSvc := service.NewSupplierService(supRepo, orgRepo)
	intOrderSvc := service.NewInternalOrderService(intOrderRepo)
	projectSvc := service.NewProjectService(projectRepo)
	invSvc := service.NewInventoryService(invRepo, repo)
	pc := product.New(cfg.ProductURL)

	// --- push в МС ---
	var pusher *mspush.Pusher
	if cfg.MSPushEnabled {
		msCli, err := msapi.New(cfg.MSToken, cfg.MSTokenFile, cfg.MSAPIBase)
		if err != nil {
			slog.Warn("mspush: disabled, msapi init failed", "error", err)
		} else {
			pusher = mspush.New(msCli, pc, repo, intOrderRepo, true)
			slog.Info("mspush: enabled", "base", cfg.MSAPIBase)
		}
	} else {
		pusher = mspush.New(nil, pc, repo, intOrderRepo, false)
		slog.Info("mspush: disabled (MS_PUSH_ENABLED=false)")
	}
	projectSvc.SetMSPusher(pusher)

	syncH := handler.NewSyncHandler(cfg.ScriptsDir)
	slog.Info("sync: scripts dir", "dir", cfg.ScriptsDir)

	h := handler.New(svc, pc, pusher)
	supH := handler.NewSupplierHandler(supSvc)
	intOrderH := handler.NewInternalOrderHandler(intOrderSvc, svc, supSvc, pc, pusher)
	projectH := handler.NewProjectHandler(projectSvc)
	invH := handler.NewInventoryHandler(invSvc, pc)

	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(mw.RequestID(), mw.Recovery(), mw.AccessLog(), mw.CORS(cfg.CORSOrigins), mw.ErrorHandler())

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "warehouse"})
		})

		read := api.Group("")
		read.Use(mw.RequireJWT(cfg.JWTSecret))
		{
			read.GET("/warehouses", h.ListWarehouses)
			read.GET("/warehouses/:id", h.GetWarehouse)
			read.GET("/stock", h.ListStock)
			read.GET("/stock/extended", h.StockExtended)
			read.GET("/stock/product/:id", h.ProductStockDetail)
			read.GET("/inventory/prepare", h.InventoryPrepare)
			read.GET("/inventories", invH.List)
			read.GET("/inventories/:id", invH.Get)
			read.GET("/inventories/:id/neighbors", invH.Neighbors)
			read.GET("/documents", h.ListDocuments)
			read.GET("/documents/:id", h.GetDocument)

			read.GET("/suppliers", supH.List)
			read.GET("/suppliers/:id", supH.Get)
			read.GET("/internal-orders", intOrderH.List)
			read.GET("/internal-orders/:id", intOrderH.Get)
			read.GET("/internal-orders/next-number", intOrderH.NextNumber)
			read.GET("/internal-orders/:id/export", intOrderH.Export)
			read.GET("/organizations", supH.ListOrganizations)
			read.GET("/projects", projectH.List)
			read.GET("/projects/:id", projectH.Get)
			read.GET("/sync/jobs/:id", syncH.Get)
		}

		whWrite := api.Group("")
		whWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			whWrite.POST("/warehouses", h.CreateWarehouse)
			whWrite.PUT("/warehouses/:id", h.UpdateWarehouse)
			whWrite.DELETE("/warehouses/:id", h.DeleteWarehouse)
		}

		supWrite := api.Group("")
		supWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "manager", "warehouse"))
		{
			supWrite.POST("/suppliers", supH.Create)
			supWrite.PUT("/suppliers/:id", supH.Update)
			supWrite.DELETE("/suppliers/:id", supH.Delete)
			supWrite.POST("/projects", projectH.Create)
			supWrite.PUT("/projects/:id", projectH.Update)
			supWrite.DELETE("/projects/:id", projectH.Delete)
			supWrite.POST("/internal-orders", intOrderH.Create)
			supWrite.POST("/sync/pull-orders", syncH.Start)
			supWrite.PUT("/internal-orders/:id", intOrderH.Update)
			supWrite.POST("/internal-orders/:id/post", intOrderH.Post)
			supWrite.POST("/internal-orders/:id/cancel", intOrderH.Cancel)
			supWrite.DELETE("/internal-orders/:id", intOrderH.Delete)
			supWrite.POST("/internal-orders/:id/print", intOrderH.Print)
			supWrite.POST("/internal-orders/:id/send", intOrderH.Send)
		}

		docWrite := api.Group("")
		docWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "manager", "warehouse"))
		{
			docWrite.POST("/documents", h.CreateDocument)
			docWrite.POST("/documents/:id/post", h.PostDocument)
			docWrite.POST("/documents/:id/cancel", h.CancelDocument)
			docWrite.POST("/inventories/:id/create-correction", invH.CreateCorrection)
		}
	}

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		slog.Info("warehouse service listening", "port", cfg.Port, "product", cfg.ProductURL)
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
	slog.Info("warehouse service stopped")
}
