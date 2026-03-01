package app

import (
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/middleware"
	v1 "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/v1"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) GetControllers(cfg *config.AppConfig) *controller.Controller {
	return &controller.Controller{
		Middleware: middleware.NewMiddleware(cfg),
		V1: v1.Controller{
			Dummy: v1.NewDummyController(),
		},
	}
}
