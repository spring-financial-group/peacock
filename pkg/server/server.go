package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	logger.Init()

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Unable to initialize config: %v\n", err)
	}
	logger.SetLevel(cfg.LogLevel)

	sources, err := NewDataSources(&cfg.DataSources)
	if err != nil {
		log.Fatalf("Unable to initialise data sources: %v\n", err)
	}
	defer sources.Close(context.Background())

	router, err := inject(cfg, sources)
	if err != nil {
		log.Fatalf("Unable to initialise router: %v\n", err)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Initialising the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error serving: %v\n", err)
		}
	}()
	log.Infof("Server started, listening on %s", srv.Addr)

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server & data sources they have 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Info("Server exiting")
}
