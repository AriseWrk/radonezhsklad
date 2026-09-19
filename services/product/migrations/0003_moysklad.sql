-- Интеграция с МойСклад: external_id / external_code / source + weight/volume

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS external_id   UUID,
    ADD COLUMN IF NOT EXISTS external_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS source        VARCHAR(20) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS weight        NUMERIC(10,3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS volume        NUMERIC(10,3) NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_external_id
    ON products (external_id) WHERE external_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_products_source
    ON products (source);

-- Единица измерения "километр" (есть в МойСклад, не было у нас)
INSERT INTO units (code, name, short_name)
VALUES ('km', 'Километр', 'км')
ON CONFLICT (code) DO NOTHING;
