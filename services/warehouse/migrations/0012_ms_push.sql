-- 0012_ms_push.sql — обратная синхронизация: push наших документов и заказов в МойСклад
--
-- Отличие "наших" от импортированных:
--   * импорт: external_id = UUID из МС, source = 'ms'
--   * созданные у нас: external_id IS NULL, source = 'manual'
--   * из инвентаризации: source = 'inventory', source_inventory_id IS NOT NULL
--
-- При успешном POST в МС пишем external_id (ответ MS) и ms_synced_at = NOW().
-- При ошибке — ms_sync_error (текст), external_id остаётся NULL.

-- ============ documents ============
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS source        VARCHAR(20) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS ms_synced_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ms_sync_error TEXT;

CREATE INDEX IF NOT EXISTS idx_documents_ms_pending
    ON documents (created_at)
    WHERE external_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_documents_source
    ON documents (source);

-- ============ internal_orders ============
ALTER TABLE internal_orders
    ADD COLUMN IF NOT EXISTS source        VARCHAR(20) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS ms_synced_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ms_sync_error TEXT;

CREATE INDEX IF NOT EXISTS idx_int_orders_ms_pending
    ON internal_orders (created_at)
    WHERE external_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_int_orders_source
    ON internal_orders (source);