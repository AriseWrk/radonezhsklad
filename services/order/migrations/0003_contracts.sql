CREATE TABLE IF NOT EXISTS contracts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    number          VARCHAR(255) NOT NULL,
    code            VARCHAR(50),
    doc_date        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    customer_id     UUID REFERENCES customers(id) ON DELETE SET NULL,
    organization_id UUID,
    amount          NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(3)   NOT NULL DEFAULT 'RUB',
    paid            NUMERIC(15,2) NOT NULL DEFAULT 0,
    fulfilled       NUMERIC(15,2) NOT NULL DEFAULT 0,
    comment         TEXT,
    printed_at      TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    archived        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contracts_customer ON contracts(customer_id);
CREATE INDEX IF NOT EXISTS idx_contracts_number   ON contracts(number);
CREATE INDEX IF NOT EXISTS idx_contracts_date     ON contracts(doc_date DESC);
CREATE INDEX IF NOT EXISTS idx_contracts_archived ON contracts(archived);

DROP TRIGGER IF EXISTS trg_contracts_updated_at ON contracts;
CREATE TRIGGER trg_contracts_updated_at
BEFORE UPDATE ON contracts FOR EACH ROW EXECUTE FUNCTION set_updated_at();