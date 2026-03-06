package flight

import (
	"errors"
	"time"
)

const (
	DefaultCurrency   = "RUB"
	DefaultStatus     = "scheduled"
	DefaultTotalSeats = 100
)

var (
	ErrFlightNotFound = errors.New("flight not found")
)

type Flight struct {
	ID             string    `db:"id" json:"id"`
	Origin         string    `db:"origin" json:"origin"`
	Destination    string    `db:"destination" json:"destination"`
	DepartureAt    time.Time `db:"departure_at" json:"departure_at"`
	ArrivalAt      time.Time `db:"arrival_at" json:"arrival_at"`
	TotalSeats     int       `db:"total_seats" json:"total_seats"`
	AvailableSeats int       `db:"available_seats" json:"available_seats"`
	PriceCents     int       `db:"price_cents" json:"price_cents"`
	Currency       string    `db:"currency" json:"currency"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type CreateFlightInput struct {
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	DepartureAt    time.Time `json:"departure_at"`
	ArrivalAt      time.Time `json:"arrival_at"`
	TotalSeats     int       `json:"total_seats"`
	PriceCents     int       `json:"price_cents"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
}

type UpdateFlightInput struct {
	Origin         *string    `json:"origin,omitempty"`
	Destination    *string    `json:"destination,omitempty"`
	DepartureAt    *time.Time `json:"departure_at,omitempty"`
	ArrivalAt      *time.Time `json:"arrival_at,omitempty"`
	TotalSeats     *int       `json:"total_seats,omitempty"`
	AvailableSeats *int       `json:"available_seats,omitempty"`
	PriceCents     *int       `json:"price_cents,omitempty"`
	Currency       *string    `json:"currency,omitempty"`
	Status         *string    `json:"status,omitempty"`
}

func (u UpdateFlightInput) HasChanges() bool {
	return u.Origin != nil ||
		u.Destination != nil ||
		u.DepartureAt != nil ||
		u.ArrivalAt != nil ||
		u.TotalSeats != nil ||
		u.AvailableSeats != nil ||
		u.PriceCents != nil ||
		u.Currency != nil ||
		u.Status != nil
}

type ListParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

const (
	FlightBookingStatusReserved  = "reserved"
	FlightBookingStatusCancelled = "cancelled"
)

var (
	ErrFlightBookingNotFound = errors.New("flight booking not found")
	ErrNoSeatsAvailable      = errors.New("no seats available")
)

type FlightBooking struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	FlightID   string    `db:"flight_id" json:"flight_id"`
	SeatNumber string    `db:"seat_number" json:"seat_number"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
