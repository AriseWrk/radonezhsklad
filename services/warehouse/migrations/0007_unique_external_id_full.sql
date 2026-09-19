-- Full UNIQUE на external_id вместо partial (нужно для ON CONFLICT)

DROP INDEX IF EXISTS idx_documents_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_documents_external_id ON documents (external_id);

DROP INDEX IF EXISTS idx_document_items_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_document_items_external_id ON document_items (external_id);

DROP INDEX IF EXISTS idx_int_orders_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_int_orders_external_id ON internal_orders (external_id);

DROP INDEX IF EXISTS idx_int_order_items_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_int_order_items_external_id ON internal_order_items (external_id);

DROP INDEX IF EXISTS idx_inventories_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_inventories_external_id ON inventories (external_id);

DROP INDEX IF EXISTS idx_inventory_items_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_inventory_items_external_id ON inventory_items (external_id);
