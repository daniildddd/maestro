DROP TABLE audit_logs;

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action          TEXT NOT NULL,
    outcome         TEXT NOT NULL,
    failure_reason  TEXT,
    actor_id        UUID,
    actor_login     TEXT,
    subject_type    TEXT NOT NULL,
    subject_id      TEXT,
    subject_name    TEXT,
    state_before    JSONB,
    state_after     JSONB,
    request_id      TEXT,
    ip              TEXT,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_logs_actor      ON audit_logs (actor_login, created_at DESC);
CREATE INDEX idx_audit_logs_action     ON audit_logs (action, created_at DESC);
