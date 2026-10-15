-- Внутренние заказы: системы и фактические выдачи
CREATE TABLE IF NOT EXISTS systems (
  code       TEXT PRIMARY KEY,
  name       TEXT NOT NULL,
  sort_order INT  NOT NULL DEFAULT 0
);

INSERT INTO systems (code, name, sort_order) VALUES
  ('SOTS', 'СОТС', 1),
  ('APS',  'АПС',  2),
  ('SOUE', 'СОУЭ', 3),
  ('SKUD', 'СКУД', 4),
  ('SOT',  'СОТ',  5),
  ('SKS',  'СКС',  6)
ON CONFLICT (code) DO NOTHING;

ALTER TABLE internal_order_items
  ADD COLUMN IF NOT EXISTS system_code TEXT REFERENCES systems(code),
  ADD COLUMN IF NOT EXISTS issued_qty  NUMERIC(15,3) NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS internal_order_issues (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id    UUID NOT NULL REFERENCES internal_orders(id) ON DELETE CASCADE,
  item_id     UUID NOT NULL REFERENCES internal_order_items(id) ON DELETE CASCADE,
  qty         NUMERIC(15,3) NOT NULL CHECK (qty > 0),
  issue_date  DATE NOT NULL DEFAULT CURRENT_DATE,
  issued_by   UUID,
  comment     TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_int_issues_order ON internal_order_issues(order_id);

ALTER TABLE internal_orders
  ADD COLUMN IF NOT EXISTS issue_status VARCHAR(20) NOT NULL DEFAULT 'none';