CREATE TABLE IF NOT EXISTS events (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255) NOT NULL,
    event_date TIMESTAMPTZ NOT NULL,
    total_seats INTEGER NOT NULL DEFAULT 0,
    available_seats INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'sold_out', 'cancelled', 'completed')),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_date ON events(event_date);
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_status_date ON events(status, event_date);

COMMENT ON TABLE events IS 'Мероприятия/события системы';