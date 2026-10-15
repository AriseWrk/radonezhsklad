-- Переход: internal_orders.project теперь хранит наш локальный projects.id (UUID),
-- а не MS external_id. MS external_id остаётся в projects.external_id и резолвится
-- на лету при push/pull в МС.
--
-- Backfill: для всех заказов, у которых project совпадает с projects.external_id,
-- заменяем значение на projects.id. Не-сматчившиеся значения (старые текстовые,
-- мусор) НЕ трогаем, чтобы не потерять данные.
--
-- Идемпотентно: повторный запуск ничего не меняет (o.project <> p.id).

BEGIN;

UPDATE internal_orders o
SET project = p.id::text
FROM projects p
WHERE p.external_id IS NOT NULL
  AND p.external_id::text = o.project
  AND o.project IS NOT NULL
  AND o.project <> ''
  AND o.project <> p.id::text;

COMMIT;