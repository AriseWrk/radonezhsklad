package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	shcfg "github.com/radonezhsklad/shared/config"
	"github.com/radonezhsklad/shared/httpx"
	"github.com/radonezhsklad/shared/logger"
	mw "github.com/radonezhsklad/shared/middleware"
	"github.com/radonezhsklad/shared/msapi"
	"github.com/radonezhsklad/warehouse/internal/config"
	"github.com/radonezhsklad/warehouse/internal/db"
	"github.com/radonezhsklad/warehouse/internal/handler"
	"github.com/radonezhsklad/warehouse/internal/mspull"
	"github.com/radonezhsklad/warehouse/internal/mspush"
	"github.com/radonezhsklad/warehouse/internal/printing"
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
	companySettingsRepo := repository.NewCompanySettingsRepo(pool)
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

	// --- MS-клиент (используется push'ем) ---
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

	// --- pull runner + sync settings ---
	internalToken := os.Getenv("INTERNAL_TOKEN")
	syncSettingsRepo := repository.NewSyncSettingsRepo(pool)
	msPullRepo := repository.NewMsPullRepo(pool)
	var pullRunner *mspull.Runner
	if msCli != nil {
		pullRunner = mspull.NewRunner(msCli, msPullRepo, syncJobRepo, pc, internalToken)
		slog.Info("mspull: enabled")
	} else {
		slog.Info("mspull: disabled (no MS client)")
	}

	// --- автосинк (pull + push, управляется тумблером sync_settings.enabled) ---
	autoSync := worker.NewAutoSync(syncJobRepo, pusher, pullRunner, syncSettingsRepo, internalToken)
	autoSync.Start(context.Background())

	templatesDir := cfg.UploadsDir
	if _, err := os.Stat(templatesDir); err != nil {
		exe, _ := os.Executable()
		templatesDir = filepath.Join(filepath.Dir(exe), "internal", "uploads", "templates")
	}
	slog.Info("templates dir", "dir", templatesDir)
	printer := printing.New(templatesDir)
	h := handler.New(svc, pc, pusher, printer)
	supH := handler.NewSupplierHandler(supSvc)
	intOrderH := handler.NewInternalOrderHandler(intOrderSvc, svc, supSvc, pc, pusher)
	projectH := handler.NewProjectHandler(projectSvc)
	invH := handler.NewInventoryHandler(invSvc, svc, supSvc, pc)
	companySettingsH := handler.NewCompanySettingsHandler(companySettingsRepo)

	syncH := handler.NewSyncHandler(syncJobRepo, msSyncErrRepo, pullRunner, pusher, autoSync, syncSettingsRepo, internalToken, "./scripts")

	fileRepo, err := repository.NewFileRepo(cfg.UploadsDir)
	if err != nil {
		slog.Error("file repo init", "error", err)
		os.Exit(1)
	}
	filesH := handler.NewFilesHandler(fileRepo)
	slog.Info("uploads dir", "dir", cfg.UploadsDir)

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
			readPts.GET("/internal-orders/:id/reserves", intOrderH.Reserves)
			readPts.GET("/organizations", supH.ListOrganizations)
			readPts.GET("/projects", projectH.List)
			readPts.GET("/projects/:id", projectH.Get)
			readPts.GET("/company/settings", companySettingsH.Get)
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
			readCore.GET("/inventories/:id/export-inv19", invH.ExportInv19)
			readCore.GET("/documents", h.ListDocuments)
			readCore.GET("/documents/:id", h.GetDocument)
			readCore.GET("/documents/:id/print/:template", h.PrintDocument)
			readCore.GET("/suppliers", supH.List)
			readCore.GET("/suppliers/:id", supH.Get)
		}

		whWrite := api.Group("")
		whWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			whWrite.POST("/warehouses", h.CreateWarehouse)
			whWrite.PUT("/warehouses/:id", h.UpdateWarehouse)
			whWrite.DELETE("/warehouses/:id", h.DeleteWarehouse)
		}

		intOrderWrite := api.Group("")
		intOrderWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse", "pts"))
		{
			intOrderWrite.POST("/internal-orders", intOrderH.Create)
			intOrderWrite.PUT("/internal-orders/:id", intOrderH.Update)
			intOrderWrite.POST("/internal-orders/:id/print", intOrderH.Print)
		}

		intOrderPost := api.Group("")
		intOrderPost.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			intOrderPost.POST("/internal-orders/:id/post", intOrderH.Post)
			intOrderPost.POST("/internal-orders/:id/cancel", intOrderH.Cancel)
			intOrderPost.POST("/internal-orders/:id/send", intOrderH.Send)
			intOrderPost.POST("/internal-orders/:id/issues", intOrderH.Issue)
			intOrderPost.DELETE("/internal-orders/:id", intOrderH.SoftDelete)
			intOrderPost.POST("/internal-orders/:id/restore", intOrderH.Restore)
			intOrderPost.DELETE("/internal-orders/:id/hard", intOrderH.HardDelete)
		}

		supWrite := api.Group("")
		supWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			supWrite.POST("/organizations", supH.CreateOrganization)
			supWrite.PUT("/organizations/:id", supH.UpdateOrganization)
			supWrite.DELETE("/organizations/:id", supH.DeleteOrganization)
			supWrite.POST("/suppliers", supH.Create)
			supWrite.PUT("/suppliers/:id", supH.Update)
			supWrite.DELETE("/suppliers/:id", supH.Delete)
			supWrite.POST("/projects", projectH.Create)
			supWrite.PUT("/projects/:id", projectH.Update)
			supWrite.DELETE("/projects/:id", projectH.Delete)
		}

		docWrite := api.Group("")
		docWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			docWrite.POST("/documents", h.CreateDocument)
			docWrite.PUT("/documents/:id", h.UpdateDocument)
			docWrite.POST("/documents/:id/post", h.PostDocument)
			docWrite.POST("/documents/:id/cancel", h.CancelDocument)
			docWrite.POST("/inventories/:id/create-correction", invH.CreateCorrection)
			docWrite.DELETE("/documents/:id", h.SoftDeleteDocument)
			docWrite.POST("/documents/:id/restore", h.RestoreDocument)
			docWrite.DELETE("/documents/:id/hard", h.HardDeleteDocument)
		}

		// company settings: только админ
		adminWrite := api.Group("")
		adminWrite.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin"))
		{
			adminWrite.PUT("/company/settings", companySettingsH.Update)
			adminWrite.GET("/files", filesH.List)
			adminWrite.POST("/files", filesH.Upload)
			adminWrite.GET("/files/:name/download", filesH.Download)
			adminWrite.DELETE("/files/:name", filesH.Delete)
			adminWrite.PATCH("/files/:name", filesH.UpdateCategory)
			adminWrite.DELETE("/trash/clear", func(c *gin.Context) {
				nDocs, err := h.ClearTrashDocumentsCount(c.Request.Context())
				if err != nil {
					c.Error(err)
					return
				}
				nOrders, err := intOrderSvc.ClearTrash(c.Request.Context())
				if err != nil {
					c.Error(err)
					return
				}
				httpx.OK(c, gin.H{"documents": nDocs, "internal_orders": nOrders})
			})
		}

		// sync: pull/push, jobs, settings
		syncGroup := api.Group("")
		syncGroup.Use(mw.RequireJWT(cfg.JWTSecret), mw.RequireRole("admin", "warehouse"))
		{
			syncGroup.GET("/sync/settings", syncH.GetSettings)
			syncGroup.PUT("/sync/settings", syncH.SetSettings)
			syncGroup.GET("/sync/jobs", syncH.List)
			syncGroup.GET("/sync/jobs/:id", syncH.Get)
			syncGroup.GET("/sync/autosync", syncH.AutoSyncStatus)
			syncGroup.GET("/sync/errors", syncH.ListErrors)
			syncGroup.POST("/sync/pull/:kind", syncH.StartPull)
			syncGroup.POST("/sync/pull-orders", syncH.Start)
			syncGroup.POST("/sync/retry-pending", syncH.StartRetryPending)
			syncGroup.POST("/sync/retry/document/:id", syncH.StartRetryDocument)
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
