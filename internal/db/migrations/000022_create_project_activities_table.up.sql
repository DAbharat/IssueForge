CREATE TYPE project_activity_type AS ENUM (
    'PROJECT_CREATED',
    'PROJECT_DETAILS_UPDATED',
    'PROJECT_MEMBER_ADDED',
    'PROJECT_MEMBER_REMOVED',
    'NEW_LEAD_ASSIGNED',
    'PROJECT_ARCHIVED',
    'PROJECT_DELETED',
    'PROJECT_RESTORED',
    'NEW_ISSUE_ADDED'
);

CREATE TABLE
    project_activities (
        id BIGSERIAL PRIMARY KEY,
        project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE RESTRICT,
        actor_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        target_user_id BIGINT REFERENCES users(id) ON DELETE RESTRICT,
        activity_type project_activity_type NOT NULL,
        field_name TEXT,
        old_value TEXT,
        new_value TEXT,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_project_activities_lookup ON project_activities (project_id, created_at DESC);

CREATE INDEX idx_project_activities_actor_created ON project_activities(actor_id, created_at DESC);