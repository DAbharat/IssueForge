CREATE TABLE
    workspace_members (
        workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
        user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        role user_role NOT NULL DEFAULT 'MEMBER',
        joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        PRIMARY KEY (workspace_id, user_id)
    );

CREATE INDEX idx_workspace_members_user_id On workspace_members (user_id);