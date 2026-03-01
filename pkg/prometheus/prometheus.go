package prometheus

import (
	"net/http"
	"time"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/pkg/transport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	Registry *prometheus.Registry
	Server   *transport.HTTPServer
}

func GetPromService(cfg *config.AppConfig) *Prometheus {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{
		Addr:         cfg.Prometheus.Host,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	return &Prometheus{
		Registry: reg,
		Server: &transport.HTTPServer{
			Mux:    mux,
			Server: srv,
		},
	}
}
