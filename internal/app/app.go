package app

import (
	"fmt"
	"strings"

	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/middleware"
	v1 "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/v1"
	adapterKafka "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/kafka"
	adapterRepo "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/repository"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/config"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/domain/flight"
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/infrastructure/db"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) GetControllers(cfg *config.AppConfig) (*controller.Controller, error) {
	postgres, err := db.NewPostgres(&cfg.Repository)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	manager := adapterRepo.NewManager(postgres)
	flightRepository := adapterRepo.NewFlightRepository(postgres, manager)
	flightModule := flight.NewModule(flightRepository)

	kafkaBrokers := cfg.Kafka.Brokers
	if len(kafkaBrokers) == 0 && cfg.Kafka.URL != "" {
		kafkaBrokers = strings.Split(cfg.Kafka.URL, ",")
	}

	sagaPubsub, err := adapterKafka.NewSagaPubsub(adapterKafka.Config{
		Brokers: kafkaBrokers,
		GroupID: cfg.Kafka.GroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("create saga pubsub: %w", err)
	}

	_ = sagaPubsub

	return &controller.Controller{
		Middleware: middleware.NewMiddleware(cfg),
		V1: v1.Controller{
			Flight: v1.NewFlightController(flightModule),
			Dummy:  v1.NewDummyController(),
		},
	}, nil
}
