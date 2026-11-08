CREATE TABLE IF NOT EXISTS sync_settings (
    id               INT PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT false,
    interval_minutes INT NOT NULL DEFAULT 10,
    last_cycle_at    TIMESTAMPTZ,
    last_status      VARCHAR(20) NOT NULL DEFAULT 'idle',
    last_error       TEXT,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT sync_settings_singleton CHECK (id = 1)
);
INSERT INTO sync_settings (id) VALUES (1) ON CONFLICT DO NOTHING;