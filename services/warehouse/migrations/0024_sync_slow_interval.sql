ALTER TABLE sync_settings ADD COLUMN IF NOT EXISTS last_slow_cycle_at TIMESTAMPTZ;
