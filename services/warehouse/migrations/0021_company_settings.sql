CREATE TABLE IF NOT EXISTS company_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    numbering VARCHAR(20) NOT NULL DEFAULT 'year',
    reply_address VARCHAR(20) NOT NULL DEFAULT 'employee',
    forbid_ship_no_stock BOOLEAN NOT NULL DEFAULT FALSE,
    auto_min_price BOOLEAN NOT NULL DEFAULT FALSE,
    move_deleted_to_bin BOOLEAN NOT NULL DEFAULT TRUE,
    use_lots BOOLEAN NOT NULL DEFAULT FALSE,
    show_extra_fields BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by UUID,
    updated_by_name VARCHAR(255),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Начальная запись
INSERT INTO company_settings (id) VALUES (1) ON CONFLICT DO NOTHING;