CREATE TYPE activity_type AS ENUM (
    'ISSUE_CREATED',
    'ISSUE_DETAILS_UPDATED',
    'ISSUE_STATUS_CHANGED',
    'ISSUE_PRIORITY_CHANGED',
    'ISSUE_ASSIGNEE_CHANGED',
    'ISSUE_DELETED',
    'ISSUE_RESTORED',
    'COMMENT_CREATED',
    'COMMENT_UPDATED',
    'COMMENT_DELETED'
);

CREATE TABLE
    issue_activities (
        id BIGSERIAL PRIMARY KEY,
        issue_id BIGINT NOT NULL REFERENCES issues (id) ON DELETE RESTRICT,
        actor_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        activity_type activity_type NOT NULL,
        field_name TEXT,
        old_value TEXT,
        new_value TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_issue_activities_lookup ON issue_activities (issue_id, created_at DESC);

CREATE INDEX idx_issue_activities_actor_created ON issue_activities(actor_id, created_at DESC);