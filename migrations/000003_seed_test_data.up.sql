-- Seed data for flights service (PostgreSQL, schema flight)

INSERT INTO flight.flights (
    id,
    origin,
    destination,
    departure_at,
    arrival_at,
    total_seats,
    available_seats,
    price_cents,
    currency,
    status,
    created_at,
    updated_at
) VALUES (
    '33333333-3333-3333-3333-333333333333',
    'LED',
    'MOW',
    NOW() + INTERVAL '1 day',
    NOW() + INTERVAL '1 day' + INTERVAL '1 hour',
    100,
    100,
    300000,
    'RUB',
    'scheduled',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

