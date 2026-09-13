package main

import (
"context"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/gin-gonic/gin"
"github.com/joho/godotenv"

"github.com/radonezhsklad/auth/internal/config"
"github.com/radonezhsklad/auth/internal/db"
"github.com/radonezhsklad/auth/internal/handler"
"github.com/radonezhsklad/auth/internal/middleware"
"github.com/radonezhsklad/auth/internal/repository"
"github.com/radonezhsklad/auth/internal/service"
)

func main() {
_ = godotenv.Load()
cfg := config.Load()
ctx := context.Background()

pool, err := db.Connect(ctx, cfg.DatabaseURL)
if err != nil { log.Fatalf("db connect: %v", err) }
defer pool.Close()
log.Println("db connected")

userRepo := repository.NewUserRepo(pool)
tokenRepo := repository.NewTokenRepo(pool)
authSvc := service.NewAuthService(userRepo, tokenRepo, cfg)
authHandler := handler.NewAuthHandler(authSvc)

if os.Getenv("ENV") == "prod" { gin.SetMode(gin.ReleaseMode) }
r := gin.New()
r.Use(gin.Logger(), gin.Recovery())

api := r.Group("/api/v1")
{
api.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "auth"})
})
api.POST("/auth/register", authHandler.Register)
api.POST("/auth/login", authHandler.Login)

authorized := api.Group("")
authorized.Use(middleware.RequireAuth(authSvc))
{
authorized.GET("/auth/me", authHandler.Me)
}

admin := api.Group("/users")
admin.Use(middleware.RequireAuth(authSvc), middleware.RequireRole("admin"))
{
admin.GET("", authHandler.ListUsers)
admin.POST("", authHandler.CreateUser)
admin.PUT("/:id", authHandler.UpdateUser)
admin.PUT("/:id/role", authHandler.UpdateRole)
admin.PUT("/:id/active", authHandler.UpdateActive)
}
}

srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
go func() {
log.Printf("auth service listening on :%s", cfg.Port)
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
log.Fatalf("listen: %v", err)
}
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
_ = srv.Shutdown(shutdownCtx)
log.Println("auth service stopped")
}