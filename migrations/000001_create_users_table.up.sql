CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,

    CONSTRAINT chk_users_role CHECK ( role IN ('admin','user') ),
    CONSTRAINT chk_username_len    CHECK ( length(username) BETWEEN 3 AND 32 ),
    CONSTRAINT chk_username_format CHECK ( username ~ '^[a-zA-Z0-9_-]+$' )
)