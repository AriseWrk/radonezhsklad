\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE tmp_docs (
    ms_id           uuid,
    number          varchar,
    moment          timestamptz,
    total           numeric,
    comment         text,
    store_uuid      uuid,
    org_uuid        uuid
) ON COMMIT DROP;

CREATE TEMP TABLE tmp_items (
    doc_ms_id           uuid,
    item_ms_id          uuid,
    product_id          uuid,
    quantity            numeric,
    calculated_quantity numeric,
    correction_amount   numeric,
    price               numeric,
    correction_sum      numeric
) ON COMMIT DROP;

\copy tmp_docs  FROM '{{DOCS_PATH}}'  WITH (FORMAT text)
\copy tmp_items FROM '{{ITEMS_PATH}}' WITH (FORMAT text)

INSERT INTO inventories (
    number, doc_date, warehouse_id, organization_id, comment, total,
    external_id, external_code
)
SELECT
    d.number, d.moment, w.id, o.id, d.comment, d.total,
    d.ms_id, d.ms_id::text
FROM tmp_docs d
LEFT JOIN warehouses    w ON w.external_id = d.store_uuid
LEFT JOIN organizations o ON o.external_id = d.org_uuid
ON CONFLICT (external_code) DO UPDATE SET
    doc_date   = EXCLUDED.doc_date,
    comment    = EXCLUDED.comment,
    total      = EXCLUDED.total,
    updated_at = NOW();

INSERT INTO inventory_items (
    inventory_id, product_id, quantity, calculated_quantity, correction_amount,
    price, correction_sum, external_id
)
SELECT inv.id, i.product_id,
    COALESCE(i.quantity,0), COALESCE(i.calculated_quantity,0), COALESCE(i.correction_amount,0),
    COALESCE(i.price,0), COALESCE(i.correction_sum,0), i.item_ms_id
FROM tmp_items i
JOIN inventories inv ON inv.external_code = i.doc_ms_id::text
WHERE i.product_id IS NOT NULL
ON CONFLICT (external_id) DO UPDATE SET
    inventory_id        = EXCLUDED.inventory_id,
    product_id          = EXCLUDED.product_id,
    quantity            = EXCLUDED.quantity,
    calculated_quantity = EXCLUDED.calculated_quantity,
    correction_amount   = EXCLUDED.correction_amount,
    price               = EXCLUDED.price,
    correction_sum      = EXCLUDED.correction_sum;

COMMIT;
