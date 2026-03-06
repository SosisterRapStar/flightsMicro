package saga

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SosisterRapStar/LETI-paper/backoff"
	paperBroker "github.com/SosisterRapStar/LETI-paper/broker"
	"github.com/SosisterRapStar/LETI-paper/controller"
	"github.com/SosisterRapStar/LETI-paper/database"
	"github.com/SosisterRapStar/LETI-paper/message"
	"github.com/SosisterRapStar/LETI-paper/retry"
	"github.com/SosisterRapStar/LETI-paper/step"
)

type flightPayload struct {
	BookingID       string `json:"booking_id"`
	UserID          string `json:"user_id"`
	FlightID        string `json:"flight_id"`
	HotelID         string `json:"hotel_id"`
	RoomID          string `json:"room_id"`
	CheckIn         string `json:"check_in"`
	CheckOut        string `json:"check_out"`
	AmountCents     int    `json:"amount_cents"`
	Currency        string `json:"currency"`
	FlightBookingID string `json:"flight_booking_id"`
}

func parsePayload[T any](msg message.Message) (*T, error) {
	raw, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return &result, nil
}

type FlightSaga struct {
	Controller        *controller.Controller
	StepFlightReserve *step.Step
	ErrCh             chan error
}

func InitFlightSaga(
	ctx context.Context,
	db *sql.DB,
	pubsub paperBroker.Pubsub,
) (*FlightSaga, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	if pubsub == nil {
		return nil, fmt.Errorf("pubsub is required")
	}

	dbCtx := database.NewDBContext(db, database.SQLDialectPostgres)

	errCh := make(chan error, 128)

	ctrl, err := controller.New(&controller.Config{
		Subscriber: pubsub,
		Publisher:  pubsub,
		DB:         dbCtx,
		InfraRetry: &retry.Retrier{
			BackoffOptions: retry.BackoffOptions{
				BackoffPolicy: backoff.Expontential{},
				MinBackoff:    50 * time.Millisecond,
				MaxBackoff:    5 * time.Second,
			},
			MaxRetries: 10,
		},
		PollInterval:  1 * time.Second,
		BatchSize:     10,
		BackoffPolicy: backoff.Expontential{},
		BackoffMin:    100 * time.Millisecond,
		BackoffMax:    1 * time.Minute,
		ErrCh:         errCh,
	})
	if err != nil {
		return nil, fmt.Errorf("create controller: %w", err)
	}

	flightReserveStep, err := step.New(&step.StepParams{
		Name: "flight-reserve",
		Routing: step.RoutingConfig{
			NextStepTopics: []string{TopicFlightReserved},
			ErrorTopics:    []string{TopicFlightFailed},
		},
		Execute: func(ctx context.Context, tx database.TxQueryer, msg message.Message) (message.Message, error) {
			p, err := parsePayload[flightPayload](msg)
			if err != nil {
				return msg, err
			}
			if p.UserID == "" || p.FlightID == "" {
				return msg, fmt.Errorf("user_id and flight_id are required")
			}

			var bookingID string
			row := tx.QueryRowContext(ctx, insertFlightBookingQuery, p.UserID, p.FlightID)
			if err := row.Scan(&bookingID); err != nil {
				return msg, fmt.Errorf("insert flight booking: %w", err)
			}

			result, err := tx.ExecContext(ctx, decrementFlightSeatsQuery, p.FlightID)
			if err != nil {
				return msg, fmt.Errorf("decrement seats: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return msg, fmt.Errorf("rows affected: %w", err)
			}
			if affected == 0 {
				return msg, fmt.Errorf("no seats available for flight %s", p.FlightID)
			}

			if msg.Payload == nil {
				msg.Payload = make(map[string]any)
			}
			msg.Payload["flight_booking_id"] = bookingID
			return msg, nil
		},
		Compensate: func(ctx context.Context, tx database.TxQueryer, msg message.Message) (message.Message, error) {
			p, err := parsePayload[flightPayload](msg)
			if err != nil {
				return msg, err
			}
			if p.FlightBookingID == "" {
				return msg, fmt.Errorf("flight_booking_id is required for compensation")
			}

			result, err := tx.ExecContext(ctx, cancelFlightBookingQuery, p.FlightBookingID)
			if err != nil {
				return msg, fmt.Errorf("cancel flight booking: %w", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return msg, fmt.Errorf("rows affected: %w", err)
			}
			if affected == 0 {
				return msg, nil
			}

			if _, err := tx.ExecContext(ctx, incrementFlightSeatsQuery, p.FlightBookingID); err != nil {
				return msg, fmt.Errorf("increment seats: %w", err)
			}
			return msg, nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create flight-reserve step: %w", err)
	}

	return &FlightSaga{
		Controller:        ctrl,
		StepFlightReserve: flightReserveStep,
		ErrCh:             errCh,
	}, nil
}
