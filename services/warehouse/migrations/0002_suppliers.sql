CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ---------- suppliers ----------
CREATE TABLE IF NOT EXISTS suppliers (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    inn        VARCHAR(20),
    phone      VARCHAR(50),
    email      VARCHAR(255),
    address    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers(name);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_suppliers_updated_at ON suppliers;
CREATE TRIGGER trg_suppliers_updated_at
BEFORE UPDATE ON suppliers FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------- organizations ----------
CREATE TABLE IF NOT EXISTS organizations (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    inn        VARCHAR(20),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO organizations (name, inn, is_default)
SELECT 'ООО Ромашка', '7701234567', TRUE
WHERE NOT EXISTS (SELECT 1 FROM organizations);

-- ---------- documents: новые поля ----------
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS supplier_id     UUID REFERENCES suppliers(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS incoming_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS incoming_date   DATE,
    ADD COLUMN IF NOT EXISTS paid_amount     NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS printed_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sent_at         TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_documents_supplier     ON documents(supplier_id);
CREATE INDEX IF NOT EXISTS idx_documents_organization ON documents(organization_id);