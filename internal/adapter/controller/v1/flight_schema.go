package v1

import "github.com/SosisterRapStar/flights/internal/domain/flight"

type createFlightRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	DepartureAt string `json:"departure_at"`
	ArrivalAt   string `json:"arrival_at"`
	TotalSeats  int    `json:"total_seats"`
	PriceCents  int    `json:"price_cents"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
}

type updateFlightRequest struct {
	Origin         *string `json:"origin,omitempty"`
	Destination    *string `json:"destination,omitempty"`
	DepartureAt    *string `json:"departure_at,omitempty"`
	ArrivalAt      *string `json:"arrival_at,omitempty"`
	TotalSeats     *int    `json:"total_seats,omitempty"`
	AvailableSeats *int    `json:"available_seats,omitempty"`
	PriceCents     *int    `json:"price_cents,omitempty"`
	Currency       *string `json:"currency,omitempty"`
	Status         *string `json:"status,omitempty"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type listFlightsResponse struct {
	Items []flight.Flight `json:"items"`
}
