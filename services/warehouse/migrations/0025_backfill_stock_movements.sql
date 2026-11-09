-- Backfill stock_movements из documents/document_items
-- 2026-11-09: восстановление истории движений для работы отчёта "Остатки на дату"
-- Идемпотентен: повторный запуск не задвоит строки (NOT EXISTS по document_id+product_id+warehouse_id+doc_type).

BEGIN;

-- 1) Движения из документов
INSERT INTO stock_movements (warehouse_id, product_id, document_id, quantity_delta, created_at, doc_type, user_id, warehouse_from, warehouse_to)
WITH mv AS (
  SELECT d.id AS doc_id, d.type::varchar AS doc_type, d.warehouse_id, di.product_id,
         di.quantity AS delta, d.doc_date, d.created_by, NULL::uuid AS wf, NULL::uuid AS wt
  FROM documents d JOIN document_items di ON di.document_id = d.id
  WHERE d.status='posted' AND d.deleted_at IS NULL AND d.type IN ('receipt','enter')
  UNION ALL
  SELECT d.id, d.type::varchar, d.warehouse_id, di.product_id,
         -di.quantity, d.doc_date, d.created_by, NULL::uuid, NULL::uuid
  FROM documents d JOIN document_items di ON di.document_id = d.id
  WHERE d.status='posted' AND d.deleted_at IS NULL AND d.type IN ('shipment','writeoff')
  UNION ALL
  SELECT d.id, 'transfer'::varchar, d.warehouse_id, di.product_id,
         -di.quantity, d.doc_date, d.created_by, d.warehouse_id, d.target_warehouse_id
  FROM documents d JOIN document_items di ON di.document_id = d.id
  WHERE d.status='posted' AND d.deleted_at IS NULL AND d.type='transfer'
  UNION ALL
  SELECT d.id, 'transfer'::varchar, d.target_warehouse_id, di.product_id,
         di.quantity, d.doc_date, d.created_by, d.warehouse_id, d.target_warehouse_id
  FROM documents d JOIN document_items di ON di.document_id = d.id
  WHERE d.status='posted' AND d.deleted_at IS NULL AND d.type='transfer' AND d.target_warehouse_id IS NOT NULL
)
SELECT warehouse_id, product_id, doc_id, delta,
       COALESCE(doc_date, now())::timestamptz, doc_type, created_by, wf, wt
FROM mv
WHERE warehouse_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM stock_movements sm
    WHERE sm.document_id = mv.doc_id
      AND sm.product_id  = mv.product_id
      AND sm.warehouse_id = mv.warehouse_id
      AND sm.doc_type     = mv.doc_type
  );

-- 2) Компенсация расхождений со снапшотом (initial_balance, чтобы «на дату = сейчас» совпадало с stock_balances)
WITH agg AS (SELECT warehouse_id, product_id, SUM(quantity_delta) AS qty FROM stock_movements GROUP BY 1,2)
INSERT INTO stock_movements (warehouse_id, product_id, quantity_delta, doc_type, created_at)
SELECT COALESCE(sb.warehouse_id, agg.warehouse_id),
       COALESCE(sb.product_id,  agg.product_id),
       COALESCE(sb.quantity,0) - COALESCE(agg.qty,0),
       'initial_balance',
       '2000-01-01 00:00:00+00'::timestamptz
FROM stock_balances sb
FULL OUTER JOIN agg ON agg.warehouse_id=sb.warehouse_id AND agg.product_id=sb.product_id
WHERE abs(COALESCE(sb.quantity,0) - COALESCE(agg.qty,0)) > 0.001
  AND COALESCE(sb.warehouse_id, agg.warehouse_id) IS NOT NULL;

COMMIT;

-- Проверка после запуска: sum_all должно равняться snapshot_total
SELECT
  (SELECT COUNT(*) FROM stock_movements) AS total_mv,
  (SELECT SUM(quantity_delta) FROM stock_movements) AS sum_all,
  (SELECT SUM(quantity) FROM stock_balances) AS snapshot_total;