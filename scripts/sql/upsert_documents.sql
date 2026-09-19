\set ON_ERROR_STOP on
BEGIN;

CREATE TEMP TABLE tmp_docs (
    ms_id             uuid,
    number            varchar,
    moment            timestamptz,
    applicable        boolean,
    total             numeric,
    vat_enabled       boolean,
    vat_included      boolean,
    comment           text,
    incoming_number   varchar,
    incoming_date     date,
    store_uuid        uuid,
    target_store_uuid uuid,
    agent_uuid        uuid,
    org_uuid          uuid
) ON COMMIT DROP;

CREATE TEMP TABLE tmp_items (
    doc_ms_id  uuid,
    item_ms_id uuid,
    product_id uuid,
    quantity   numeric,
    price      numeric,
    vat        numeric,
    discount   numeric,
    sum        numeric
) ON COMMIT DROP;

\copy tmp_docs  FROM '{{DOCS_PATH}}'  WITH (FORMAT text)
\copy tmp_items FROM '{{ITEMS_PATH}}' WITH (FORMAT text)

SELECT
    COUNT(*)                                                        AS total_docs,
    COUNT(*) FILTER (WHERE w.id IS NULL)                            AS no_store,
    COUNT(*) FILTER (WHERE s.id IS NULL AND agent_uuid IS NOT NULL) AS no_agent,
    COUNT(*) FILTER (WHERE o.id IS NULL AND org_uuid   IS NOT NULL) AS no_org
FROM tmp_docs d
LEFT JOIN warehouses    w ON w.external_id = d.store_uuid
LEFT JOIN suppliers     s ON s.external_id = d.agent_uuid
LEFT JOIN organizations o ON o.external_id = d.org_uuid;

INSERT INTO documents (
    type, number, status, warehouse_id, target_warehouse_id, comment,
    supplier_id, organization_id, incoming_number, incoming_date, paid_amount,
    posted_at, external_id, external_code, doc_date, total, vat_enabled, vat_included
)
SELECT
    '{{DOC_TYPE}}', d.number,
    CASE WHEN COALESCE(d.applicable, false) THEN 'posted' ELSE 'draft' END,
    w.id, tw.id, d.comment,
    s.id, o.id,
    d.incoming_number, d.incoming_date,
    0,
    CASE WHEN COALESCE(d.applicable, false) THEN d.moment ELSE NULL END,
    d.ms_id, d.ms_id::text, d.moment, d.total, COALESCE(d.vat_enabled, false), COALESCE(d.vat_included, false)
FROM tmp_docs d
LEFT JOIN warehouses    w  ON w.external_id  = d.store_uuid
LEFT JOIN warehouses    tw ON tw.external_id = d.target_store_uuid
LEFT JOIN suppliers     s  ON s.external_id  = d.agent_uuid
LEFT JOIN organizations o  ON o.external_id  = d.org_uuid
WHERE w.id IS NOT NULL
ON CONFLICT (external_code) DO UPDATE SET
    status          = EXCLUDED.status,
    total           = EXCLUDED.total,
    doc_date        = EXCLUDED.doc_date,
    posted_at       = EXCLUDED.posted_at,
    comment         = EXCLUDED.comment,
    incoming_number = EXCLUDED.incoming_number,
    incoming_date   = EXCLUDED.incoming_date,
    updated_at      = NOW();

INSERT INTO document_items (
    document_id, product_id, quantity, price, external_id, vat_rate, discount, sum
)
SELECT d.id, i.product_id, i.quantity, i.price, i.item_ms_id, i.vat, i.discount, i.sum
FROM tmp_items i
JOIN documents d ON d.external_code = i.doc_ms_id::text
WHERE i.product_id IS NOT NULL
ON CONFLICT (external_id) DO UPDATE SET
    document_id = EXCLUDED.document_id,
    product_id  = EXCLUDED.product_id,
    quantity    = EXCLUDED.quantity,
    price       = EXCLUDED.price,
    vat_rate    = EXCLUDED.vat_rate,
    discount    = EXCLUDED.discount,
    sum         = EXCLUDED.sum;

COMMIT;
