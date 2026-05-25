ALTER TABLE users
    ADD COLUMN IF NOT EXISTS profile_photo_url TEXT;

UPDATE users
SET profile_photo_url = avatar_url
WHERE profile_photo_url IS NULL
  AND avatar_url IS NOT NULL;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('owner', 'runner', 'admin', 'shop', 'support_agent'));

CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    notifications_enabled JSONB NOT NULL DEFAULT '{}'::jsonb,
    language VARCHAR(2) NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'ms', 'zh')),
    theme VARCHAR(10) NOT NULL DEFAULT 'system' CHECK (theme IN ('system', 'light', 'dark')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind VARCHAR(20) NOT NULL CHECK (kind IN ('ic', 'license', 'selfie', 'vehicle_reg')),
    storage_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    reviewed_by_user_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_documents_user_id ON user_documents(user_id);
CREATE INDEX IF NOT EXISTS idx_user_documents_status ON user_documents(status);
