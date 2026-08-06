DROP INDEX IF EXISTS idx_projects_workspace_name_unique;

ALTER TABLE projects
ADD CONSTRAINT projects_workspace_id_name_key
UNIQUE (workspace_id, name);

ALTER TABLE projects
DROP COLUMN deleted_at;