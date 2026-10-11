-- 0015: nullable warehouse_id for purchase docs (order/invoice/facture)
-- Also adds explicit type list documentation.
ALTER TABLE documents ALTER COLUMN warehouse_id DROP NOT NULL;