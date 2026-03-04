package ports

import "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/domain/flight"

type FlightRepository interface {
	flight.Repository
}
