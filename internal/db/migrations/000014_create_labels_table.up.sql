CREATE TABLE
    labels (
        id BIGSERIAL PRIMARY KEY,
        project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        color TEXT NOT NULL DEFAULT '#525252',
        created_by BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMPTZ
    );

CREATE UNIQUE INDEX idx_labels_project_name
ON labels(project_id, LOWER(name))
WHERE deleted_at IS NULL;

CREATE INDEX idx_labels_created_by 
ON labels (created_by)
WHERE deleted_at IS NULL;

CREATE TRIGGER update_labels_updated_at
    BEFORE UPDATE ON labels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();