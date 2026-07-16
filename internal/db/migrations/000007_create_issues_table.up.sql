CREATE TYPE issue_status AS ENUM ('TODO', 'IN_PROGRESS', 'DONE');

CREATE TYPE issue_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH');

CREATE TABLE
    issues (
        id BIGSERIAL PRIMARY KEY,
        project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
        created_by BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
        assigned_to BIGINT REFERENCES users (id) ON DELETE SET NULL,
        title VARCHAR(100) NOT NULL,
        description TEXT NOT NULL,
        status issue_status NOT NULL DEFAULT 'TODO',
        priority issue_priority NOT NULL DEFAULT 'MEDIUM',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_issues_project_id ON issues(project_id);
CREATE INDEX idx_issues_assigned_to ON issues(assigned_to) WHERE assigned_to IS NOT NULL;

CREATE TRIGGER update_issues_modtime
    BEFORE UPDATE ON issues
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();
    