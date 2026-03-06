CREATE TABLE IF NOT EXISTS flight.flight_bookings (
    booking_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    flight_id UUID NOT NULL REFERENCES flight.flights(id) ON DELETE CASCADE,
    seat_number TEXT,
    status TEXT NOT NULL DEFAULT 'reserved' CHECK (status IN ('reserved', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flight_bookings_user_id ON flight.flight_bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_flight_bookings_flight_id ON flight.flight_bookings(flight_id);
CREATE INDEX IF NOT EXISTS idx_flight_bookings_status ON flight.flight_bookings(status);
