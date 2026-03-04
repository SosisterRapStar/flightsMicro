package flight

import (
	"context"
	"errors"
	"fmt"
)

const (
	maxListLimit = 100
)

type Repository interface {
	Create(context.Context, CreateFlightInput) (*Flight, error)
	GetByID(context.Context, string) (*Flight, error)
	List(context.Context, ListParams) ([]Flight, error)
	Update(context.Context, string, UpdateFlightInput) (*Flight, error)
	Delete(context.Context, string) error
}

type Module interface {
	Create(context.Context, CreateFlightInput) (*Flight, error)
	GetByID(context.Context, string) (*Flight, error)
	List(context.Context, ListParams) ([]Flight, error)
	Update(context.Context, string, UpdateFlightInput) (*Flight, error)
	Delete(context.Context, string) error
}

type module struct {
	repository Repository
}

func NewModule(repository Repository) Module {
	return &module{repository: repository}
}

func (m *module) Create(ctx context.Context, input CreateFlightInput) (*Flight, error) {
	if input.Origin == "" || input.Destination == "" {
		return nil, errors.New("origin and destination are required")
	}
	if !input.ArrivalAt.After(input.DepartureAt) {
		return nil, errors.New("arrival_at must be after departure_at")
	}
	if input.TotalSeats <= 0 {
		return nil, errors.New("total_seats must be positive")
	}
	if input.PriceCents < 0 {
		return nil, errors.New("price_cents must be non-negative")
	}
	if input.Currency == "" {
		input.Currency = DefaultCurrency
	}
	if input.Status == "" {
		input.Status = DefaultStatus
	}

	created, err := m.repository.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("creating flight: %w", err)
	}
	return created, nil
}

func (m *module) GetByID(ctx context.Context, id string) (*Flight, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	item, err := m.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting flight by id: %w", err)
	}
	return item, nil
}

func (m *module) List(ctx context.Context, params ListParams) ([]Flight, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > maxListLimit {
		params.Limit = maxListLimit
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	items, err := m.repository.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("listing flights: %w", err)
	}
	return items, nil
}

func (m *module) Update(ctx context.Context, id string, input UpdateFlightInput) (*Flight, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if !input.HasChanges() {
		return nil, errors.New("no fields to update")
	}
	item, err := m.repository.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("updating flight: %w", err)
	}
	return item, nil
}

func (m *module) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	if err := m.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting flight: %w", err)
	}
	return nil
}
