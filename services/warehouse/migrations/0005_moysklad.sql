-- Интеграция с МойСклад: suppliers / organizations / warehouses

-- suppliers (те же контрагенты, что и customers, роль поставщика)
ALTER TABLE suppliers
    ADD COLUMN IF NOT EXISTS external_id       UUID,
    ADD COLUMN IF NOT EXISTS external_code     VARCHAR(100),
    ADD COLUMN IF NOT EXISTS counterparty_type VARCHAR(100),
    ADD COLUMN IF NOT EXISTS legal_address     TEXT,
    ADD COLUMN IF NOT EXISTS actual_address    TEXT,
    ADD COLUMN IF NOT EXISTS kpp               VARCHAR(20),
    ADD COLUMN IF NOT EXISTS ogrn              VARCHAR(20),
    ADD COLUMN IF NOT EXISTS okpo              VARCHAR(20),
    ADD COLUMN IF NOT EXISTS fax               VARCHAR(50),
    ADD COLUMN IF NOT EXISTS comment           TEXT,
    ADD COLUMN IF NOT EXISTS archived          BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS source            VARCHAR(20) NOT NULL DEFAULT 'manual';

CREATE UNIQUE INDEX IF NOT EXISTS idx_suppliers_external_id
    ON suppliers (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_suppliers_external_code
    ON suppliers (external_code);
CREATE INDEX IF NOT EXISTS idx_suppliers_source
    ON suppliers (source);

-- organizations
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS external_id   UUID,
    ADD COLUMN IF NOT EXISTS external_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS kpp           VARCHAR(20),
    ADD COLUMN IF NOT EXISTS ogrn          VARCHAR(20),
    ADD COLUMN IF NOT EXISTS okpo          VARCHAR(20),
    ADD COLUMN IF NOT EXISTS legal_address TEXT,
    ADD COLUMN IF NOT EXISTS email         VARCHAR(255),
    ADD COLUMN IF NOT EXISTS archived      BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS source        VARCHAR(20) NOT NULL DEFAULT 'manual';

CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_external_id
    ON organizations (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_external_code
    ON organizations (external_code);
CREATE INDEX IF NOT EXISTS idx_organizations_source
    ON organizations (source);

-- warehouses
ALTER TABLE warehouses
    ADD COLUMN IF NOT EXISTS external_id   UUID,
    ADD COLUMN IF NOT EXISTS external_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS source        VARCHAR(20) NOT NULL DEFAULT 'manual';

CREATE UNIQUE INDEX IF NOT EXISTS idx_warehouses_external_id
    ON warehouses (external_id) WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_warehouses_external_code
    ON warehouses (external_code);
CREATE INDEX IF NOT EXISTS idx_warehouses_source
    ON warehouses (source);
