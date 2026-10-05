-- 0014_ms_sync_errors.sql — append-only журнал ошибок синхронизации с МС.
--
-- Отличие от documents.ms_sync_error:
--   * documents.ms_sync_error  — «последняя» ошибка, перезаписывается
--   * ms_sync_errors           — история попыток (append), для retry и разбора
--
-- Поля:
--   entity      — 'document' | 'internalorder' | 'pull-doc' | ...
--   local_id    — наш UUID (может быть NULL для pull-ошибок)
--   ms_external_id — UUID в МС (для диагностики)
--   op          — 'create' | 'update' | 'cancel' | 'pull' | 'positions'
--   attempt     — номер попытки
--   error       — текст
--   resolved_at — когда проблема решилась (NULL = открыта)

CREATE TABLE IF NOT EXISTS ms_sync_errors (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity         VARCHAR(40)  NOT NULL,
    local_id       UUID,
    ms_external_id UUID,
    op             VARCHAR(20)  NOT NULL,
    attempt        INT          NOT NULL DEFAULT 1,
    error          TEXT         NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    resolved_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ms_sync_errors_open
    ON ms_sync_errors (created_at DESC)
    WHERE resolved_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ms_sync_errors_local
    ON ms_sync_errors (local_id) WHERE local_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ms_sync_errors_ext
    ON ms_sync_errors (ms_external_id) WHERE ms_external_id IS NOT NULL;

COMMENT ON TABLE ms_sync_errors IS 'Append-only журнал ошибок синхронизации с МойСклад';