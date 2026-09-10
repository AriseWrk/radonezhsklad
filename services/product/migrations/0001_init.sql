-- ============================================================
-- RadonezhSklad / product / 0001_init
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- общая функция триггера
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------- categories ----------
CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    parent_id  UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);

DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
CREATE TRIGGER trg_categories_updated_at
BEFORE UPDATE ON categories
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------- units ----------
CREATE TABLE IF NOT EXISTS units (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code       VARCHAR(20) NOT NULL UNIQUE,
    name       VARCHAR(50) NOT NULL,
    short_name VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO units (code, name, short_name) VALUES
    ('pcs',  'Штука',     'шт'),
    ('kg',   'Килограмм', 'кг'),
    ('m',    'Метр',      'м'),
    ('l',    'Литр',      'л'),
    ('pack', 'Упаковка',  'упак')
ON CONFLICT (code) DO NOTHING;

-- ---------- products ----------
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(500) NOT NULL,
    sku         VARCHAR(100) UNIQUE,
    barcode     VARCHAR(50),
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    unit_id     UUID REFERENCES units(id)      ON DELETE RESTRICT,
    description TEXT,
    price       NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency    VARCHAR(3)    NOT NULL DEFAULT 'RUB',
    is_archived BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_archived ON products(is_archived);
CREATE INDEX IF NOT EXISTS idx_products_name     ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_barcode  ON products(barcode);

DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();