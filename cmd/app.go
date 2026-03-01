package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/app"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/infrastructure/router"
)

const shutdownTimeout = 10 * time.Second

func main() {
	cfg := &config.AppConfig{
		Server: config.Server{
			Address: ":8080",
		},
		API: config.API{
			Timeout:           30 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	runServer(cfg)
}

func runServer(cfg *config.AppConfig) {
	a := app.New()
	controllers := a.GetControllers(cfg)
	mux := router.NewMux(cfg, controllers)

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

	log.Info("server stopped gracefully")
}
