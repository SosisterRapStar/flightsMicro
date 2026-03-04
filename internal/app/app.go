package app

import (
	"fmt"
	"strings"

	"github.com/SosisterRapStar/flights/internal/adapter/controller"
	"github.com/SosisterRapStar/flights/internal/adapter/controller/middleware"
	v1 "github.com/SosisterRapStar/flights/internal/adapter/controller/v1"
	adapterKafka "github.com/SosisterRapStar/flights/internal/adapter/kafka"
	adapterRepo "github.com/SosisterRapStar/flights/internal/adapter/repository"
	"github.com/SosisterRapStar/flights/internal/config"
	"github.com/SosisterRapStar/flights/internal/domain/flight"
	"github.com/SosisterRapStar/flights/internal/infrastructure/db"
	infrakafka "github.com/SosisterRapStar/flights/internal/infrastructure/kafka"
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

	brokers := cfg.Kafka.Brokers
	if len(brokers) == 0 && cfg.Kafka.URL != "" {
		brokers = strings.Split(cfg.Kafka.URL, ",")
	}

	kafkaCfg := &infrakafka.Config{
		Brokers:          brokers,
		GroupID:          cfg.Kafka.GroupID,
		AckPolicy:        cfg.Kafka.Producer.AckPolicy,
		RetryMax:         cfg.Kafka.Producer.RetryMax,
		AutoCommitEnable: cfg.Kafka.Consumer.AutoCommitEnable,
		MaxWaitTime:      cfg.Kafka.Consumer.MaxWaitTime,
	}

	sagaPubsub, err := adapterKafka.NewSagaPubsub(kafkaCfg)
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
