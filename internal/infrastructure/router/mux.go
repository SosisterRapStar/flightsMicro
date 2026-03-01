package router

import (
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
	chi "github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewMux(cfg *config.AppConfig, c *controller.Controller) chi.Router {
	r := chi.NewRouter()

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(c.Middleware.Logger)
		r.Use(c.Middleware.Monitoring)

		r.Get("/dummy", c.V1.Dummy.GetDummy)
	})
	return r
}
