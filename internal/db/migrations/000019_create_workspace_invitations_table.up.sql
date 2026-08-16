CREATE TABLE
    workspace_invitations (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
        invited_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        invited_by BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'DECLINED', 'CANCELLED')),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (), 
        responded_at TIMESTAMPTZ NULL
    );

CREATE UNIQUE INDEX workspace_invitations_pending_unique 
ON workspace_invitations (workspace_id, invited_user_id)
WHERE status = 'PENDING';