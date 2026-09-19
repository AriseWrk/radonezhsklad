-- Интеграция с МойСклад: документы (documents, internal_orders, inventories)

-- ============ documents ============
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS external_id   UUID,
    ADD COLUMN IF NOT EXISTS external_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS doc_date      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS total         NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS vat_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS vat_included  BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_external_id
    ON documents (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_documents_external_code
    ON documents (external_code);
CREATE INDEX IF NOT EXISTS idx_documents_doc_date
    ON documents (doc_date DESC);

-- ============ document_items ============
ALTER TABLE document_items
    ADD COLUMN IF NOT EXISTS external_id UUID,
    ADD COLUMN IF NOT EXISTS vat_rate    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS discount    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sum         NUMERIC(15,2) NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_document_items_external_id
    ON document_items (external_id) WHERE external_id IS NOT NULL;

-- ============ internal_orders ============
ALTER TABLE internal_orders
    ADD COLUMN IF NOT EXISTS external_id   UUID,
    ADD COLUMN IF NOT EXISTS external_code VARCHAR(100);

CREATE UNIQUE INDEX IF NOT EXISTS idx_int_orders_external_id
    ON internal_orders (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_int_orders_external_code
    ON internal_orders (external_code);

-- ============ internal_order_items ============
ALTER TABLE internal_order_items
    ADD COLUMN IF NOT EXISTS external_id UUID;

CREATE UNIQUE INDEX IF NOT EXISTS idx_int_order_items_external_id
    ON internal_order_items (external_id) WHERE external_id IS NOT NULL;

-- ============ inventories (инвентаризации) ============
CREATE TABLE IF NOT EXISTS inventories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id     UUID,
    external_code   VARCHAR(100),
    number          VARCHAR(50) NOT NULL,
    doc_date        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    warehouse_id    UUID REFERENCES warehouses(id)    ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    comment         TEXT,
    total           NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_inventories_external_id
    ON inventories (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventories_external_code
    ON inventories (external_code);
CREATE INDEX IF NOT EXISTS idx_inventories_doc_date
    ON inventories (doc_date DESC);

DROP TRIGGER IF EXISTS trg_inventories_updated_at ON inventories;
CREATE TRIGGER trg_inventories_updated_at
BEFORE UPDATE ON inventories FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS inventory_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inventory_id        UUID NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
    external_id         UUID,
    product_id          UUID NOT NULL,
    quantity            NUMERIC(15,3) NOT NULL DEFAULT 0,
    calculated_quantity NUMERIC(15,3) NOT NULL DEFAULT 0,
    correction_amount   NUMERIC(15,3) NOT NULL DEFAULT 0,
    price               NUMERIC(15,2) NOT NULL DEFAULT 0,
    correction_sum      NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_items_inv
    ON inventory_items (inventory_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_product
    ON inventory_items (product_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_items_external_id
    ON inventory_items (external_id) WHERE external_id IS NOT NULL;
