-- 0019: расширение складских операций (резервы, денормализация movements, связи документов)

-- ============ stock_reserves: резервы под внутренние заказы ============
CREATE TABLE IF NOT EXISTS stock_reserves (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    product_id   UUID NOT NULL,
    order_id     UUID REFERENCES internal_orders(id) ON DELETE CASCADE,
    quantity     NUMERIC(15,3) NOT NULL CHECK (quantity > 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reserves_wh_prod ON stock_reserves(warehouse_id, product_id);
CREATE INDEX IF NOT EXISTS idx_reserves_order   ON stock_reserves(order_id);

-- ============ stock_movements: денормализация для отчётов ============
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS doc_type      VARCHAR(20);
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS user_id       UUID;
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS warehouse_from UUID REFERENCES warehouses(id) ON DELETE SET NULL;
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS warehouse_to   UUID REFERENCES warehouses(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_movements_doc_type ON stock_movements(doc_type);
CREATE INDEX IF NOT EXISTS idx_movements_user     ON stock_movements(user_id);

-- ============ documents: связь родитель-потомок и причина списания ============
ALTER TABLE documents ADD COLUMN IF NOT EXISTS parent_document_id UUID REFERENCES documents(id) ON DELETE SET NULL;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS reason             TEXT;
CREATE INDEX IF NOT EXISTS idx_documents_parent ON documents(parent_document_id) WHERE parent_document_id IS NOT NULL;

-- ============ internal_orders: склад-приёмник и статусы shipped/closed ============
ALTER TABLE internal_orders ADD COLUMN IF NOT EXISTS target_warehouse_id UUID REFERENCES warehouses(id) ON DELETE SET NULL;
ALTER TABLE internal_orders ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMPTZ;
ALTER TABLE internal_orders ADD COLUMN IF NOT EXISTS closed_at  TIMESTAMPTZ;
