package saga

const (
	insertFlightBookingQuery = `
INSERT INTO flight.flight_bookings (user_id, flight_id, status)
VALUES ($1, $2, 'reserved')
RETURNING booking_id;
`

	decrementFlightSeatsQuery = `
UPDATE flight.flights
SET available_seats = available_seats - 1, updated_at = NOW()
WHERE id = $1 AND available_seats > 0;
`

	cancelFlightBookingQuery = `
UPDATE flight.flight_bookings
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND status = 'reserved';
`

	incrementFlightSeatsQuery = `
UPDATE flight.flights
SET available_seats = available_seats + 1, updated_at = NOW()
WHERE id = (SELECT flight_id FROM flight.flight_bookings WHERE id = $1);
`
)
