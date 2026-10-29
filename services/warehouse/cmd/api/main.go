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
	"github.com/radonezhsklad/warehouse/internal/mspull"
	"github.com/radonezhsklad/warehouse/internal/mspush"
	"github.com/radonezhsklad/warehouse/internal/product"
	"github.com/radonezhsklad/warehouse/internal/repository"
	"github.com/radonezhsklad/warehouse/internal/service"
	"github.com/radonezhsklad/warehouse/internal/worker"
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
	intIssueRepo := repository.NewInternalOrderIssueRepo(pool)
	projectRepo := repository.NewProjectRepo(pool)
	invRepo := repository.NewInventoryRepo(pool)
	orgRepo := repository.NewOrganizationRepo(pool)
	syncJobRepo := repository.NewSyncJobRepo(pool)
	msSyncErrRepo := repository.NewMsSyncErrorRepo(pool)

	// На старте помечаем «зависшие» job'ы (running/queued от прошлого процесса) как error.
	if n, err := syncJobRepo.MarkStaleAsError(ctx); err != nil {
		slog.Warn("sync: MarkStaleAsError failed", "error", err)
	} else if n > 0 {
		slog.Info("sync: stale jobs marked as error", "count", n)
	}

	svc := service.New(repo)
	supSvc := service.NewSupplierService(supRepo, orgRepo)
	intOrderSvc := service.NewInternalOrderService(intOrderRepo, intIssueRepo, repo)
	projectSvc := service.NewProjectService(projectRepo)
	invSvc := service.NewInventoryService(invRepo, repo, svc)
	pc := product.New(cfg.ProductURL)

	// --- MS-клиент (используется и push'ем, и pull'ом) ---
	var msCli *msapi.Client
	if cfg.MSPushEnabled {
		cli, msErr := msapi.New(cfg.MSToken, cfg.MSTokenFile, cfg.MSAPIBase)
		if msErr != nil {
			slog.Warn("msapi: init failed, MS-функции отключены", "error", msErr)
		} else {
			msCli = cli
			slog.Info("msapi: enabled", "base", cfg.MSAPIBase)
		}
	} else {
		slog.Info("msapi: disabled (MS_PUSH_ENABLED=false)")
	}

	// --- push в МС ---
	var pusher *mspush.Pusher
	if msCli != nil {
		pusher = mspush.New(msCli, pc, repo, intOrderRepo, msSyncErrRepo, true)
		slog.Info("mspush: enabled")
	} else {
		pusher = mspush.New(nil, pc, repo, intOrderRepo, msSyncErrRepo, false)
		slog.Info("mspush: disabled")
	}
	projectSvc.SetMSPusher(pusher)

	// --- pull справочников из МС ---
	msPullRepo := repository.NewMsPullRepo(pool)
	internalToken := os.Getenv("INTERNAL_TOKEN")
	pullRunner := mspull.NewRunner(msCli, msPullRepo, syncJobRepo, pc, internalToken)
	if pullRunner.Enabled() {
		slog.Info("mspull: enabled")
	} else {
		slog.Info("mspull: disabled")
	}

	// --- автосинк ---
	autoSyncInterval := 15 * time.Minute
	if v := os.Getenv("MS_AUTOSYNC_INTERVAL"); v != "" {
		if d, perr := time.ParseDuration(v); perr == nil && d > 0 {
			autoSyncInterval = d
		}
	}
	autoSyncEnabled := os.Getenv("MS_AUTOSYNC_ENABLED") == "true"
	autoSync := worker.NewAutoSync(syncJobRepo, pusher, internalToken, autoSyncInterval, autoSyncEnabled)
	autoSync.Start(context.Background())

	syncH := handler.NewSyncHandler(syncJobRepo, msSyncErrRepo, pullRunner, pusher, autoSync, internalToken, cfg.ScriptsDir)
	slog.Info("sync: scripts dir", "dir", cfg.ScriptsDir)

	h := handler.New(svc, pc, pusher)
	supH := handler.NewSupplierHandler(supSvc)
	intOrderH := handler.NewInternalOrderHandler(intOrderSvc, svc, supSvc, pc, pusher)
	projectH := handler.NewProjectHandler(projectSvc)
	invH := handler.NewInventoryHandler(invSvc, svc, supSvc, pc)

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

		// readPts — что видит ПТС (админ/склад/ПТС)
		readPts := api.Group("")
		readPts.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse", "pts"))
		{
			readPts.GET("/warehouses", h.ListWarehouses)
			readPts.GET("/warehouses/:id", h.GetWarehouse)
			readPts.GET("/stock", h.ListStock)
			readPts.GET("/stock/extended", h.StockExtended)
			readPts.GET("/stock/product/:id", h.ProductStockDetail)
			readPts.GET("/internal-orders", intOrderH.List)
			readPts.GET("/internal-orders/next-number", intOrderH.NextNumber)
			readPts.GET("/internal-orders/:id", intOrderH.Get)
			readPts.GET("/internal-orders/:id/export", intOrderH.Export)
			readPts.GET("/internal-orders/:id/issues", intOrderH.ListIssues)
			readPts.GET("/organizations", supH.ListOrganizations)
			readPts.GET("/projects", projectH.List)
			readPts.GET("/projects/:id", projectH.Get)
		}

		// readCore — только админ/склад
		readCore := api.Group("")
		readCore.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			readCore.GET("/reports/turnover", h.TurnoverReport)
			readCore.GET("/inventory/prepare", h.InventoryPrepare)
			readCore.GET("/inventories", invH.List)
			readCore.GET("/inventories/:id", invH.Get)
			readCore.GET("/inventories/:id/neighbors", invH.Neighbors)
			readCore.GET("/inventories/:id/export", invH.Export)
			readCore.GET("/inventories/:id/export-inv3", invH.ExportInv3)
			readCore.GET("/documents", h.ListDocuments)
			readCore.GET("/documents/:id", h.GetDocument)
			readCore.GET("/suppliers", supH.List)
			readCore.GET("/suppliers/:id", supH.Get)
			readCore.GET("/sync/jobs", syncH.List)
			readCore.GET("/sync/errors", syncH.ListErrors)
			readCore.GET("/sync/autosync", syncH.AutoSyncStatus)
			readCore.GET("/sync/jobs/:id", syncH.Get)
		}

		whWrite := api.Group("")
		whWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			whWrite.POST("/warehouses", h.CreateWarehouse)
			whWrite.PUT("/warehouses/:id", h.UpdateWarehouse)
			whWrite.DELETE("/warehouses/:id", h.DeleteWarehouse)
		}

		// internal-orders CRUD + print — админ/склад/ПТС
		intOrderWrite := api.Group("")
		intOrderWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse", "pts"))
		{
			intOrderWrite.POST("/internal-orders", intOrderH.Create)
			intOrderWrite.PUT("/internal-orders/:id", intOrderH.Update)
			intOrderWrite.POST("/internal-orders/:id/print", intOrderH.Print)
		}

		// internal-orders: проведение/отмена/отправка/выдача/удаление — только админ/склад
		intOrderPost := api.Group("")
		intOrderPost.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			intOrderPost.POST("/internal-orders/:id/post", intOrderH.Post)
			intOrderPost.POST("/internal-orders/:id/cancel", intOrderH.Cancel)
			intOrderPost.POST("/internal-orders/:id/send", intOrderH.Send)
			intOrderPost.POST("/internal-orders/:id/issues", intOrderH.Issue)
			intOrderPost.DELETE("/internal-orders/:id", intOrderH.Delete)
		}

		supWrite := api.Group("")
		supWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			supWrite.POST("/suppliers", supH.Create)
			supWrite.PUT("/suppliers/:id", supH.Update)
			supWrite.DELETE("/suppliers/:id", supH.Delete)
			supWrite.POST("/projects", projectH.Create)
			supWrite.PUT("/projects/:id", projectH.Update)
			supWrite.DELETE("/projects/:id", projectH.Delete)
			supWrite.POST("/sync/pull-orders", syncH.Start)
			supWrite.POST("/sync/retry-pending", syncH.StartRetryPending)
			supWrite.POST("/sync/retry/document/:id", syncH.StartRetryDocument)
			supWrite.POST("/sync/pull/:kind", syncH.StartPull)
		}

		docWrite := api.Group("")
		docWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
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
