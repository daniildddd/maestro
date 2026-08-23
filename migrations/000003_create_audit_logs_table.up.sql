CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    action TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    connector_id TEXT,
    connector_name TEXT,
    state_before JSONB,
    state_after JSONB
)
