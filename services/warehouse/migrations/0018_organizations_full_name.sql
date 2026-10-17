-- Полное юридическое наименование организации (как в МС: counterparty.legalTitle).
-- name остаётся кратким; в печатные формы (ИНВ-3 и др.) идёт full_name, если заполнено.

BEGIN;

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS full_name VARCHAR(1000);

-- Backfill: пока нет данных из МС — заполняем равным краткому имени.
UPDATE organizations SET full_name = name WHERE full_name IS NULL;

COMMIT;