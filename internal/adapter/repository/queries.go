package repository

const (
	createFlightQuery = `
INSERT INTO flight.flights (
	origin,
	destination,
	departure_at,
	arrival_at,
	total_seats,
	available_seats,
	price_cents,
	currency,
	status
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, origin, destination, departure_at, arrival_at,
	total_seats, available_seats, price_cents, currency, status,
	created_at, updated_at;
`

	getFlightByIDQuery = `
SELECT id, origin, destination, departure_at, arrival_at,
	total_seats, available_seats, price_cents, currency, status,
	created_at, updated_at
FROM flight.flights
WHERE id = $1;
`

	listFlightsQuery = `
SELECT id, origin, destination, departure_at, arrival_at,
	total_seats, available_seats, price_cents, currency, status,
	created_at, updated_at
FROM flight.flights
ORDER BY departure_at ASC
LIMIT $1 OFFSET $2;
`

	deleteFlightByIDQuery = `
DELETE FROM flight.flights
WHERE id = $1;
`
)
