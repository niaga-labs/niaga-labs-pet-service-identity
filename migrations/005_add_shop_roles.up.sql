CREATE TABLE IF NOT EXISTS user_roles (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(30) NOT NULL,
    scope_type VARCHAR(20) NOT NULL DEFAULT 'global',
    scope_id   UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role, scope_type, scope_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_scope ON user_roles(user_id, scope_type);
CREATE INDEX IF NOT EXISTS idx_user_roles_scope ON user_roles(scope_type, scope_id);
