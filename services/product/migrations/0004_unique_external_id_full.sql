-- Full UNIQUE на products.external_id (partial не работает с ON CONFLICT)

DROP INDEX IF EXISTS idx_products_external_id;
CREATE UNIQUE INDEX IF NOT EXISTS uq_products_external_id ON products (external_id);

