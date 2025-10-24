package server

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/logger"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Msgf("Unable to initialize config: %v", err)
		return
	}
	logger.SetLevel(cfg.LogLevel)

	sources, err := NewDataSources(&cfg.DataSources)
	if err != nil {
		log.Fatal().Msgf("Unable to initialise data sources: %v", err)
		return
	}

	router := inject(cfg, sources)
	defer sources.Close(context.Background())

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
	}

	// Initialising the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Error serving: %v", err)
		}
	}()
	log.Info().Msgf("Server started, listening on %s", srv.Addr)

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info().Msg("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %v", err)
	}

	log.Info().Msg("Server exiting")
}
