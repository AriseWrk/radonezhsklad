-- is_printed + warehouse_name из МС (Этап 45)
ALTER TABLE internal_orders
    ADD COLUMN IF NOT EXISTS is_printed BOOLEAN NOT NULL DEFAULT FALSE;