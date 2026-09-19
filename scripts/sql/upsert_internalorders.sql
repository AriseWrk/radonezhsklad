\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE tmp_docs (
    ms_id           uuid,
    number          varchar,
    moment          timestamptz,
    applicable      boolean,
    total           numeric,
    vat_enabled     boolean,
    vat_included    boolean,
    comment         text,
    plan_date       date,
    store_uuid      uuid,
    org_uuid        uuid,
    project         varchar
) ON COMMIT DROP;

CREATE TEMP TABLE tmp_items (
    doc_ms_id  uuid,
    item_ms_id uuid,
    product_id uuid,
    quantity   numeric,
    price      numeric,
    vat        numeric,
    sum        numeric
) ON COMMIT DROP;

\copy tmp_docs  FROM '{{DOCS_PATH}}'  WITH (FORMAT text)
\copy tmp_items FROM '{{ITEMS_PATH}}' WITH (FORMAT text)

INSERT INTO internal_orders (
    number, doc_date, status, organization_id, warehouse_id,
    plan_date, project, comment, total, vat_enabled, vat_included,
    posted_at, external_id, external_code
)
SELECT
    d.number, d.moment,
    CASE WHEN COALESCE(d.applicable,false) THEN 'posted' ELSE 'draft' END,
    o.id, w.id,
    d.plan_date, d.project, d.comment,
    d.total, COALESCE(d.vat_enabled,false), COALESCE(d.vat_included,false),
    CASE WHEN COALESCE(d.applicable,false) THEN d.moment ELSE NULL END,
    d.ms_id, d.ms_id::text
FROM tmp_docs d
LEFT JOIN warehouses    w ON w.external_id = d.store_uuid
LEFT JOIN organizations o ON o.external_id = d.org_uuid
ON CONFLICT (external_code) DO UPDATE SET
    doc_date     = EXCLUDED.doc_date,
    status       = EXCLUDED.status,
    plan_date    = EXCLUDED.plan_date,
    project      = EXCLUDED.project,
    comment      = EXCLUDED.comment,
    total        = EXCLUDED.total,
    posted_at    = EXCLUDED.posted_at,
    updated_at   = NOW();

INSERT INTO internal_order_items (
    order_id, product_id, quantity, price, vat_rate, sum, external_id
)
SELECT o.id, i.product_id,
    COALESCE(i.quantity,0), COALESCE(i.price,0),
    COALESCE(i.vat,0), COALESCE(i.sum,0), i.item_ms_id
FROM tmp_items i
JOIN internal_orders o ON o.external_code = i.doc_ms_id::text
WHERE i.product_id IS NOT NULL
ON CONFLICT (external_id) DO UPDATE SET
    order_id   = EXCLUDED.order_id,
    product_id = EXCLUDED.product_id,
    quantity   = EXCLUDED.quantity,
    price      = EXCLUDED.price,
    vat_rate   = EXCLUDED.vat_rate,
    sum        = EXCLUDED.sum;

COMMIT;
