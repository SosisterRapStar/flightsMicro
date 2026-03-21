CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS flight;

CREATE TABLE IF NOT EXISTS flight.flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin TEXT NOT NULL,
    destination TEXT NOT NULL,
    departure_at TIMESTAMP NOT NULL,
    arrival_at TIMESTAMP NOT NULL,
    total_seats INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'RUB',
    status TEXT NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (arrival_at > departure_at),
    CHECK (total_seats > 0),
    CHECK (available_seats >= 0),
    CHECK (available_seats <= total_seats)
);

CREATE INDEX IF NOT EXISTS idx_flights_departure_at ON flight.flights(departure_at);
CREATE INDEX IF NOT EXISTS idx_flights_origin_destination ON flight.flights(origin, destination);
CREATE INDEX IF NOT EXISTS idx_flights_status ON flight.flights(status);
