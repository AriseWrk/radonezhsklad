-- Связь корректирующего документа с инвентаризацией (Этап 42)
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS source_inventory_id UUID REFERENCES inventories(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_documents_source_inventory
    ON documents (source_inventory_id) WHERE source_inventory_id IS NOT NULL;