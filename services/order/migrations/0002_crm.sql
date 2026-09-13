ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS external_code     VARCHAR(100),
  ADD COLUMN IF NOT EXISTS status           VARCHAR(50)  NOT NULL DEFAULT 'Новый',
  ADD COLUMN IF NOT EXISTS comment          TEXT,
  ADD COLUMN IF NOT EXISTS group_name       VARCHAR(255),
  ADD COLUMN IF NOT EXISTS full_name        VARCHAR(500),
  ADD COLUMN IF NOT EXISTS last_name        VARCHAR(100),
  ADD COLUMN IF NOT EXISTS first_name       VARCHAR(100),
  ADD COLUMN IF NOT EXISTS middle_name      VARCHAR(100),
  ADD COLUMN IF NOT EXISTS legal_address    TEXT,
  ADD COLUMN IF NOT EXISTS actual_address   TEXT,
  ADD COLUMN IF NOT EXISTS inn              VARCHAR(20),
  ADD COLUMN IF NOT EXISTS kpp              VARCHAR(20),
  ADD COLUMN IF NOT EXISTS ogrn             VARCHAR(20),
  ADD COLUMN IF NOT EXISTS okpo             VARCHAR(20),
  ADD COLUMN IF NOT EXISTS fax              VARCHAR(50),
  ADD COLUMN IF NOT EXISTS counterparty_type VARCHAR(100),
  ADD COLUMN IF NOT EXISTS archived         BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_customers_status   ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_archived ON customers(archived);
CREATE INDEX IF NOT EXISTS idx_customers_inn      ON customers(inn);