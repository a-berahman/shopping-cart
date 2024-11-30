CREATE TYPE item_status AS ENUM (
    'PENDING',
    'AVAILABILITY_CHECK',
    'AVAILABLE',
    'UNAVAILABLE',
    'RESERVED',
    'FAILED'
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    reservation_id TEXT,
    status item_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_status ON items(status);

