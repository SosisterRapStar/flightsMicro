package app

import (
	"context"
	"fmt"
	"strings"

	paperController "github.com/SosisterRapStar/LETI-paper/controller"
	"github.com/SosisterRapStar/flights/internal/adapter/controller"
	"github.com/SosisterRapStar/flights/internal/adapter/controller/middleware"
	v1 "github.com/SosisterRapStar/flights/internal/adapter/controller/v1"
	adapterKafka "github.com/SosisterRapStar/flights/internal/adapter/kafka"
	adapterRepo "github.com/SosisterRapStar/flights/internal/adapter/repository"
	"github.com/SosisterRapStar/flights/internal/config"
	"github.com/SosisterRapStar/flights/internal/domain/flight"
	"github.com/SosisterRapStar/flights/internal/infrastructure/db"
	infrakafka "github.com/SosisterRapStar/flights/internal/infrastructure/kafka"
	"github.com/SosisterRapStar/flights/internal/infrastructure/telemetry"
	"github.com/SosisterRapStar/flights/internal/saga"
)

type App struct {
	Controller *controller.Controller
}

func New(cfg *config.AppConfig) (*App, error) {
	postgres, err := db.NewPostgres(&cfg.Repository)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	flightRepository := adapterRepo.NewFlightRepository(postgres)
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

	if err := telemetry.Init(cfg); err != nil {
		return nil, fmt.Errorf("init telemetry: %w", err)
	}

	ctx := context.Background()
	var sagaTracing *paperController.TracingConfig
	if cfg.Tracing.Enabled {
		sagaTracing = &paperController.TracingConfig{
			Tracer:     telemetry.Tracer("flights-saga"),
			TracerName: "flights",
		}
	}
	flightSaga, err := saga.InitFlightSaga(ctx, postgres.DB, sagaPubsub, sagaTracing)
	if err != nil {
		return nil, fmt.Errorf("init flight saga: %w", err)
	}

	if err := flightSaga.Controller.Register(saga.TopicBookingCreated, flightSaga.StepFlightReserve); err != nil {
		return nil, fmt.Errorf("register %s step: %w", saga.TopicBookingCreated, err)
	}

	if err := flightSaga.Controller.Init(ctx); err != nil {
		return nil, fmt.Errorf("init flight saga controller: %w", err)
	}
	if err := sagaPubsub.Run(ctx); err != nil {
		return nil, fmt.Errorf("run saga pubsub: %w", err)
	}
	return &App{
		Controller: &controller.Controller{
			Middleware: middleware.NewMiddleware(cfg),
			V1: v1.Controller{
				Flight: v1.NewFlightController(flightModule),
				Dummy:  v1.NewDummyController(),
			},
		},
	}, nil
}
