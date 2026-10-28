package main

import (
	"context"
	"errors"
	"fmt"
	"group/app/api"
	"group/app/api/groups"
	"group/app/config"
	"group/app/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	wa := groups.NewClient(
		cfg.OpenWAURL, cfg.OpenWAAdminAPIKey,
	)
	
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	router := gin.New()
	router.Use(gin.Recovery())
	// webhook URL carries the token in the query string, keep it out of the logs
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/api/webhooks/openwa"}}))

	// Handlers
	groupHandler := groups.GroupHandler{WA: wa, Pool: pool, WebhookToken: cfg.WebhookToken}

	// Publics routes
	// Health endpoint with ping in database
	router.GET("/health", func(c *gin.Context) {
		err := pool.Ping(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})
	router.POST("/api/webhooks/openwa", groupHandler.Webhook)

	// Protected routes
	protected := router.Group("/api", api.APIKey(cfg.APIKey))
	protected.GET("/groups/:sessionID", groupHandler.List)
	protected.GET("/groups/:sessionID/:groupID", groupHandler.Get)
	protected.PATCH("/groups/:sessionID/:groupID", groupHandler.SetManaged)
	protected.DELETE("/groups/:sessionID/:groupID/participants", groupHandler.RemoveParticipant)
	protected.POST("/groups/:sessionID/sync", groupHandler.Sync)
	protected.POST("/groups/:sessionID/:groupID/sync", groupHandler.SyncGroup)
	protected.GET("/groups/:sessionID/:groupID/members", groupHandler.Members)
	protected.GET("/groups/:sessionID/:groupID/members/returned", groupHandler.Returned)
	protected.POST("/groups/:sessionID/:groupID/removals", groupHandler.CreateRemoval)
	protected.GET("/groups/:sessionID/:groupID/removals", groupHandler.ListRemovals)
	protected.GET("/removals/:id", groupHandler.GetRemoval)
	protected.POST("/removals/:id/cancel", groupHandler.CancelRemoval)


	// ctx is canceled on Ctrl+C (SIGINT) or on stop from docker/systemd (SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workerDone := make(chan struct{})
	go func() {
		groups.Worker{WA: wa, Pool: pool}.Run(ctx)
		close(workerDone)
	}()

	syncDone := make(chan struct{})
	go func() {
		groupHandler.RunAutoSync(ctx, cfg.OpenWASessionID, cfg.SyncInterval)
		close(syncDone)
	}()

	srv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port), Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	// stop accepting requests and wait the ones in progress (sync can take a while)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	// pool is closed by the defer only after the background jobs are done with it
	<-workerDone
	<-syncDone
	log.Println("bye")
}
