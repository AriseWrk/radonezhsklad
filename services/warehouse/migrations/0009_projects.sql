-- Справочник проектов из МойСклад (Этап 44)
CREATE TABLE IF NOT EXISTS projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id     UUID UNIQUE,
    name            VARCHAR(500) NOT NULL,
    archived        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_external_id ON projects (external_id);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects (name);