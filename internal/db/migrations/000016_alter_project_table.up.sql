ALTER TABLE projects
ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE projects
DROP CONSTRAINT projects_workspace_id_name_key;

CREATE UNIQUE INDEX idx_projects_workspace_name_unique
ON projects(workspace_id, name)
WHERE deleted_at IS NULL;