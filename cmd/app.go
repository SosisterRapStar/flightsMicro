package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/SosisterRapStar/flights/internal/app"
	"github.com/SosisterRapStar/flights/internal/config"
	_ "github.com/SosisterRapStar/flights/internal/docs"
	"github.com/SosisterRapStar/flights/internal/infrastructure/router"
	"github.com/SosisterRapStar/flights/internal/infrastructure/telemetry"
)

const shutdownTimeout = 10 * time.Second

// @title Flights API
// @version 1.0
// @description API для управления рейсами (flights microservice).
// @BasePath /api/v1
func main() {
	cfg := config.MustLoad("config.yaml")
	runServer(cfg)
}

func runServer(cfg *config.AppConfig) {
	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("building app: %v", err)
	}
	mux := router.NewMux(cfg, a.Controller)

	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           mux,
		ReadHeaderTimeout: cfg.API.ReadHeaderTimeout,
	}

	go func() {
		log.Infof("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listening on %s: %v", srv.Addr, err)
		}
	}()

	shutdown(srv)
}

func shutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Infof("received signal %s, shutting down", sig)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	if err := telemetry.Shutdown(ctx); err != nil {
		log.Warnf("telemetry shutdown: %v", err)
	}

	log.Info("server stopped gracefully")
}
