package router

import (
	"github.com/SosisterRapStar/flights/internal/adapter/controller"
	"github.com/SosisterRapStar/flights/internal/config"
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
		if c.Middleware.Monitoring != nil {
			r.Use(c.Middleware.Monitoring)
		}

		r.Route("/flights", func(r chi.Router) {
			r.Post("/", c.V1.Flight.Create)
			r.Get("/", c.V1.Flight.List)
			r.Get("/{id}", c.V1.Flight.GetByID)
			r.Patch("/{id}", c.V1.Flight.Update)
			r.Delete("/{id}", c.V1.Flight.Delete)
		})

		r.Get("/dummy", c.V1.Dummy.GetDummy)
	})
	return r
}
