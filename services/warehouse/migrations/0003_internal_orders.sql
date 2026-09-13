CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS internal_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    number          VARCHAR(50) NOT NULL,
    doc_date        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    warehouse_id    UUID REFERENCES warehouses(id)    ON DELETE SET NULL,
    plan_date       DATE,
    project         VARCHAR(255),
    comment         TEXT,
    total           NUMERIC(15,2) NOT NULL DEFAULT 0,
    vat_enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    vat_included    BOOLEAN NOT NULL DEFAULT TRUE,
    posted_at       TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_int_orders_status  ON internal_orders(status);
CREATE INDEX IF NOT EXISTS idx_int_orders_created ON internal_orders(created_at DESC);

DROP TRIGGER IF EXISTS trg_int_orders_updated_at ON internal_orders;
CREATE TRIGGER trg_int_orders_updated_at
BEFORE UPDATE ON internal_orders FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS internal_order_items (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id    UUID NOT NULL REFERENCES internal_orders(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL,
    quantity    NUMERIC(15,3) NOT NULL,
    price       NUMERIC(15,2) NOT NULL DEFAULT 0,
    vat_rate    NUMERIC(5,2)  NOT NULL DEFAULT 20,
    sum         NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_int_order_items_order ON internal_order_items(order_id);