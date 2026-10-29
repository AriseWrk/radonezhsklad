-- 0020: связь движений остатков с внутренними заказами (выдачи ВЗ)
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS internal_order_id UUID REFERENCES internal_orders(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_movements_int_order ON stock_movements(internal_order_id) WHERE internal_order_id IS NOT NULL;
