package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/SosisterRapStar/flights/internal/domain/flight"
	"github.com/jmoiron/sqlx"
)

type FlightRepository struct {
	db      *sqlx.DB
	manager *Manager
}

func NewFlightRepository(db *sqlx.DB, manager *Manager) *FlightRepository {
	return &FlightRepository{
		db:      db,
		manager: manager,
	}
}

func (r *FlightRepository) Create(ctx context.Context, input flight.CreateFlightInput) (*flight.Flight, error) {
	availableSeats := input.TotalSeats

	row := flight.Flight{}
	queryer := r.queryer(ctx)
	err := sqlx.GetContext(
		ctx,
		queryer,
		&row,
		createFlightQuery,
		input.Origin,
		input.Destination,
		input.DepartureAt,
		input.ArrivalAt,
		input.TotalSeats,
		availableSeats,
		input.PriceCents,
		input.Currency,
		input.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("executing create flight query: %w", err)
	}
	return &row, nil
}

func (r *FlightRepository) GetByID(ctx context.Context, id string) (*flight.Flight, error) {
	row := flight.Flight{}
	queryer := r.queryer(ctx)
	if err := sqlx.GetContext(ctx, queryer, &row, getFlightByIDQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, flight.ErrFlightNotFound
		}
		return nil, fmt.Errorf("executing get flight by id query: %w", err)
	}
	return &row, nil
}

func (r *FlightRepository) List(ctx context.Context, params flight.ListParams) ([]flight.Flight, error) {
	rows := make([]flight.Flight, 0, params.Limit)
	queryer := r.queryer(ctx)
	if err := sqlx.SelectContext(ctx, queryer, &rows, listFlightsQuery, params.Limit, params.Offset); err != nil {
		return nil, fmt.Errorf("executing list flights query: %w", err)
	}
	return rows, nil
}

func (r *FlightRepository) Update(ctx context.Context, id string, input flight.UpdateFlightInput) (*flight.Flight, error) {
	setClauses := make([]string, 0, 10)
	args := make([]any, 0, 11)

	nextArg := 1
	if input.Origin != nil {
		setClauses = append(setClauses, fmt.Sprintf("origin = $%d", nextArg))
		args = append(args, *input.Origin)
		nextArg++
	}
	if input.Destination != nil {
		setClauses = append(setClauses, fmt.Sprintf("destination = $%d", nextArg))
		args = append(args, *input.Destination)
		nextArg++
	}
	if input.DepartureAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("departure_at = $%d", nextArg))
		args = append(args, *input.DepartureAt)
		nextArg++
	}
	if input.ArrivalAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("arrival_at = $%d", nextArg))
		args = append(args, *input.ArrivalAt)
		nextArg++
	}
	if input.TotalSeats != nil {
		setClauses = append(setClauses, fmt.Sprintf("total_seats = $%d", nextArg))
		args = append(args, *input.TotalSeats)
		nextArg++
	}
	if input.AvailableSeats != nil {
		setClauses = append(setClauses, fmt.Sprintf("available_seats = $%d", nextArg))
		args = append(args, *input.AvailableSeats)
		nextArg++
	}
	if input.PriceCents != nil {
		setClauses = append(setClauses, fmt.Sprintf("price_cents = $%d", nextArg))
		args = append(args, *input.PriceCents)
		nextArg++
	}
	if input.Currency != nil {
		setClauses = append(setClauses, fmt.Sprintf("currency = $%d", nextArg))
		args = append(args, *input.Currency)
		nextArg++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", nextArg))
		args = append(args, *input.Status)
		nextArg++
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
UPDATE flight.flights
SET %s
WHERE id = $%d
RETURNING id, origin, destination, departure_at, arrival_at,
	total_seats, available_seats, price_cents, currency, status,
	created_at, updated_at;
`, strings.Join(setClauses, ", "), nextArg)

	row := flight.Flight{}
	queryer := r.queryer(ctx)
	if err := sqlx.GetContext(ctx, queryer, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, flight.ErrFlightNotFound
		}
		return nil, fmt.Errorf("executing update flight query: %w", err)
	}
	return &row, nil
}

func (r *FlightRepository) Delete(ctx context.Context, id string) error {
	execer := r.execer(ctx)
	result, err := execer.ExecContext(ctx, deleteFlightByIDQuery, id)
	if err != nil {
		return fmt.Errorf("executing delete flight query: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reading delete flight rows affected: %w", err)
	}
	if affected == 0 {
		return flight.ErrFlightNotFound
	}
	return nil
}

func (r *FlightRepository) queryer(ctx context.Context) sqlx.QueryerContext {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *FlightRepository) execer(ctx context.Context) sqlx.ExecerContext {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
