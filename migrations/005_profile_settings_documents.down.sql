DROP INDEX IF EXISTS idx_user_documents_status;
DROP INDEX IF EXISTS idx_user_documents_user_id;
DROP TABLE IF EXISTS user_documents;
DROP TABLE IF EXISTS user_settings;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('owner', 'runner', 'admin'));

ALTER TABLE users
    DROP COLUMN IF EXISTS profile_photo_url;
