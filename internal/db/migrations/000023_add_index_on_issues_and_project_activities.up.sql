CREATE INDEX idx_issues_project_created ON issues(project_id, created_at)
WHERE deleted_at IS NULL;

CREATE INDEX idx_issue_activities_created ON issue_activities(created_at DESC);

CREATE INDEX idx_project_activities_created ON project_activities(created_at DESC);