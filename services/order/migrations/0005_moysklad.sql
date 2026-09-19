-- Интеграция с МойСклад: external_id / source для контрагентов (customers)

ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS external_id UUID,
    ADD COLUMN IF NOT EXISTS source      VARCHAR(20) NOT NULL DEFAULT 'manual';

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_external_id
    ON customers (external_id) WHERE external_id IS NOT NULL;

-- external_code уже существует с прошлых миграций; нужен full-unique для UPSERT
-- (partial не подходит для ON CONFLICT (external_code))
DROP INDEX IF EXISTS idx_customers_external_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_external_code
    ON customers (external_code);

CREATE INDEX IF NOT EXISTS idx_customers_source
    ON customers (source);
