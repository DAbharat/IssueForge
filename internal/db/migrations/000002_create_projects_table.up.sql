CREATE TABLE
    IF NOT EXISTS projects (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
        lead_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        name TEXT NOT NULL,
        description TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        CONSTRAINT projects_workspace_id_name_key UNIQUE (workspace_id, name)
    );

CREATE INDEX IF NOT EXISTS idx_projects_lead_id ON projects (lead_id);

CREATE INDEX idx_projects_workspace ON projects(workspace_id);

CREATE
OR REPLACE TRIGGER update_projects_updated_at BEFORE
UPDATE ON projects FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column ();