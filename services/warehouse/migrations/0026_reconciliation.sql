-- 0026_reconciliation.sql
-- История диагностических сверок локального склада с МойСклад.

CREATE TABLE reconciliation_runs (
    id                  UUID PRIMARY KEY,
    status              VARCHAR(20) NOT NULL DEFAULT 'running'
                            CHECK (status IN ('running', 'completed', 'failed')),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at         TIMESTAMPTZ,
    triggered_by        UUID,
    triggered_by_name   VARCHAR(255),
    price_type_name     VARCHAR(100) NOT NULL DEFAULT 'Цена продажи',
    error               TEXT,

    total_products          INTEGER NOT NULL DEFAULT 0,
    matched_products        INTEGER NOT NULL DEFAULT 0,
    price_differences       INTEGER NOT NULL DEFAULT 0,
    stock_differences      INTEGER NOT NULL DEFAULT 0,
    missing_local_products INTEGER NOT NULL DEFAULT 0,
    missing_ms_products    INTEGER NOT NULL DEFAULT 0,
    ambiguous_products     INTEGER NOT NULL DEFAULT 0,
    unmatched_warehouses   INTEGER NOT NULL DEFAULT 0,
    zero_stock_items       INTEGER NOT NULL DEFAULT 0,

    CHECK (total_products >= 0),
    CHECK (matched_products >= 0),
    CHECK (price_differences >= 0),
    CHECK (stock_differences >= 0),
    CHECK (missing_local_products >= 0),
    CHECK (missing_ms_products >= 0),
    CHECK (ambiguous_products >= 0),
    CHECK (unmatched_warehouses >= 0),
    CHECK (zero_stock_items >= 0)
);

CREATE INDEX idx_reconciliation_runs_started
    ON reconciliation_runs (started_at DESC);

CREATE INDEX idx_reconciliation_runs_status
    ON reconciliation_runs (status, started_at DESC);

CREATE TABLE reconciliation_items (
    id                  UUID PRIMARY KEY,
    run_id              UUID NOT NULL REFERENCES reconciliation_runs(id) ON DELETE CASCADE,

    item_type           VARCHAR(30) NOT NULL
                            CHECK (item_type IN ('product', 'price', 'stock', 'warehouse')),
    result_status       VARCHAR(30) NOT NULL
                            CHECK (result_status IN (
                                'matched',
                                'difference',
                                'missing_local',
                                'missing_moysklad',
                                'ambiguous',
                                'unmatched'
                            )),

    match_method        VARCHAR(20)
                            CHECK (match_method IS NULL OR match_method IN ('external_id', 'sku')),

    local_product_id    UUID,
    ms_product_id       VARCHAR(100),
    sku                 VARCHAR(255),
    local_product_name  TEXT,
    ms_product_name     TEXT,

    local_warehouse_id  UUID REFERENCES warehouses(id) ON DELETE SET NULL,
    ms_store_id         VARCHAR(100),
    local_warehouse_name TEXT,
    ms_store_name       TEXT,

    local_price         NUMERIC(18,4),
    ms_price            NUMERIC(18,4),
    price_difference    NUMERIC(18,4),

    local_stock         NUMERIC(18,4),
    ms_stock            NUMERIC(18,4),
    stock_difference    NUMERIC(18,4),
    is_zero_stock       BOOLEAN NOT NULL DEFAULT FALSE,

    details             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reconciliation_items_run
    ON reconciliation_items (run_id);

CREATE INDEX idx_reconciliation_items_run_type_status
    ON reconciliation_items (run_id, item_type, result_status);

CREATE INDEX idx_reconciliation_items_local_product
    ON reconciliation_items (local_product_id)
    WHERE local_product_id IS NOT NULL;

CREATE INDEX idx_reconciliation_items_sku
    ON reconciliation_items (sku)
    WHERE sku IS NOT NULL;

CREATE INDEX idx_reconciliation_items_warehouse
    ON reconciliation_items (local_warehouse_id)
    WHERE local_warehouse_id IS NOT NULL;

COMMENT ON TABLE reconciliation_runs IS
    'История диагностических сверок остатков и цен с МойСклад; без записи во внешнюю систему';

COMMENT ON TABLE reconciliation_items IS
    'Снимки результатов сверки товаров, цен, остатков и сопоставления складов';
-- Одновременно допускается только одна выполняющаяся сверка.
CREATE UNIQUE INDEX idx_reconciliation_runs_single_running
    ON reconciliation_runs (status)
    WHERE status = 'running';