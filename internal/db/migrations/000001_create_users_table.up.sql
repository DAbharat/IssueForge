DO $$
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('ADMIN', 'MEMBER');
    END IF;
END $$;

CREATE TABLE
    IF NOT EXISTS users (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        workspace_id BIGINT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
        fullname TEXT NOT NULL,
        display_name TEXT NOT NULL,
        email TEXT NOT NULL,
        password_hash TEXT NOT NULL,
        role user_role NOT NULL DEFAULT 'MEMBER',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_users_workspace ON users(workspace_id);
    
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_workspace_email 
ON users (workspace_id, email)
WHERE workspace_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_onboarding_email 
ON users (email)
WHERE workspace_id IS NULL;

CREATE OR REPLACE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();