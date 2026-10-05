-- 0013_sync_jobs.sql — persistent-история задач синхронизации с МойСклад.
--
-- Заменяет in-memory map в SyncHandler. Даёт:
--   * историю запусков (ручных и авто)
--   * возможность узнать "идёт ли синк" после рестарта api.exe
--   * курсоры last_success для инкрементального pull
--   * терминальные состояния (done/error/cancelled)
--
-- Семантика:
--   status: queued | running | done | error | cancelled
--   kind:   pull-orders | pull-docs-enter | pull-docs-demand | pull-docs-loss | pull-docs-move
--           pull-products | pull-counterparties | pull-warehouses | pull-organizations | pull-projects
--           autosync | retry-pending
--   cursor_after: значение, до которого прошли (ISO8601 updated, либо NULL)
--   cursor_before: значение, с которого начали
--   external_state: JSONB для агрегатных курсоров (напр. {"enter":"...","demand":"..."})

CREATE TABLE IF NOT EXISTS sync_jobs (
    id             UUID PRIMARY KEY,
    kind           VARCHAR(40)  NOT NULL,
    status         VARCHAR(20)  NOT NULL,
    started_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    ended_at       TIMESTAMPTZ,
    last_line      TEXT,
    error          TEXT,
    fetched        INTEGER      NOT NULL DEFAULT 0,
    total          INTEGER,
    cursor_before  TEXT,
    cursor_after   TEXT,
    external_state JSONB,
    triggered_by   VARCHAR(40)  NOT NULL DEFAULT 'manual'  -- manual | auto | api
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_started
    ON sync_jobs (started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_kind_started
    ON sync_jobs (kind, started_at DESC);

-- Быстрый поиск "активных" — для блокировки параллельного запуска.
CREATE INDEX IF NOT EXISTS idx_sync_jobs_active
    ON sync_jobs (status)
    WHERE status IN ('queued', 'running');

-- Курсор последнего успешного job'а по kind — для инкрементального pull.
-- Используем view, чтобы не плодить таблицу-состояние.
CREATE OR REPLACE VIEW v_sync_cursor AS
SELECT DISTINCT ON (kind)
       kind,
       cursor_after  AS last_cursor,
       ended_at      AS last_success_at
FROM sync_jobs
WHERE status = 'done' AND cursor_after IS NOT NULL
ORDER BY kind, ended_at DESC;

COMMENT ON TABLE  sync_jobs IS 'История задач синхронизации с МойСклад (pull/push/retry)';
COMMENT ON VIEW   v_sync_cursor IS 'Курсор последнего успешного pull по каждому kind';