CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID,
    user_email  VARCHAR(255),
    method      VARCHAR(10) NOT NULL,
    path        VARCHAR(500) NOT NULL,
    resource    VARCHAR(100),
    resource_id VARCHAR(100),
    status      INT NOT NULL,
    request_body TEXT,
    client_ip   VARCHAR(64),
    request_id  VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_user      ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_resource  ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_created   ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_method    ON audit_logs(method);