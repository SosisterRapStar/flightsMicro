package ports

import "github.com/SosisterRapStar/flights/internal/domain/flight"

type FlightRepository interface {
	flight.Repository
}
