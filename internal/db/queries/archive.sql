-- name: GetArchivedCounts :one
SELECT
    (SELECT COUNT(*) FROM workspaces w
     JOIN workspace_members wm ON w.id = wm.workspace_id
     WHERE wm.user_id = $1 AND wm.role = 'ADMIN' AND w.deleted_at IS NOT NULL) AS workspaces_count,
    (SELECT COUNT(*) FROM projects p
     WHERE p.deleted_at IS NOT NULL
       AND (p.lead_id = $1 OR EXISTS (
           SELECT 1 FROM workspace_members wm
           WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
       ))) AS projects_count,
    (SELECT COUNT(*) FROM issues i
     JOIN projects p ON i.project_id = p.id
     WHERE i.deleted_at IS NOT NULL
       AND (i.created_by = $1 OR i.assigned_to = $1 OR p.lead_id = $1 OR EXISTS (
           SELECT 1 FROM workspace_members wm
           WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
       ))) AS issues_count,
    (SELECT COUNT(*) FROM issue_attachments ia
     JOIN issues i ON ia.issue_id = i.id
     JOIN projects p ON i.project_id = p.id
     WHERE ia.deleted_at IS NOT NULL
       AND (ia.uploaded_by = $1 OR p.lead_id = $1 OR EXISTS (
           SELECT 1 FROM workspace_members wm
           WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
       ))) AS attachments_count;

-- name: ListArchivedWorkspaces :many
SELECT w.id, w.name, w.created_at, w.updated_at, w.deleted_at
FROM workspaces w
JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE wm.user_id = $1 AND wm.role = 'ADMIN' AND w.deleted_at IS NOT NULL
  AND (sqlc.narg(search)::text IS NULL OR w.name ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY w.deleted_at DESC;

-- name: ListArchivedProjects :many
SELECT p.id, p.workspace_id, w.name AS workspace_name, p.lead_id, u.fullname AS lead_name,
       p.name, p.description, p.created_at, p.updated_at, p.deleted_at
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
JOIN users u ON p.lead_id = u.id
WHERE p.deleted_at IS NOT NULL
  AND (p.lead_id = $1 OR EXISTS (
      SELECT 1 FROM workspace_members wm
      WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
  ))
  AND (sqlc.narg(search)::text IS NULL OR p.name ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY p.deleted_at DESC;

-- name: ListArchivedIssues :many
SELECT i.id, i.project_id, p.name AS project_name, p.workspace_id, w.name AS workspace_name,
       i.created_by, creator.fullname AS creator_name,
       i.assigned_to, assignee.fullname AS assignee_name,
       i.title, i.description, i.status, i.priority, i.created_at, i.updated_at, i.due_date, i.deleted_at
FROM issues i
JOIN projects p ON i.project_id = p.id
JOIN workspaces w ON p.workspace_id = w.id
JOIN users creator ON i.created_by = creator.id
LEFT JOIN users assignee ON i.assigned_to = assignee.id
WHERE i.deleted_at IS NOT NULL
  AND (i.created_by = $1 OR i.assigned_to = $1 OR p.lead_id = $1 OR EXISTS (
      SELECT 1 FROM workspace_members wm
      WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
  ))
  AND (sqlc.narg(search)::text IS NULL OR i.title ILIKE '%' || sqlc.narg(search) || '%' OR CAST(i.id AS TEXT) ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY i.deleted_at DESC;

-- name: ListArchivedAttachments :many
SELECT ia.id, ia.issue_id, i.title AS issue_title, i.project_id, p.name AS project_name,
       ia.uploaded_by, uploader.fullname AS uploader_name,
       ia.original_name, ia.resource_type, ia.mime_type, ia.file_size, ia.public_url,
       ia.created_at, ia.deleted_at
FROM issue_attachments ia
JOIN issues i ON ia.issue_id = i.id
JOIN projects p ON i.project_id = p.id
JOIN users uploader ON ia.uploaded_by = uploader.id
WHERE ia.deleted_at IS NOT NULL
  AND (ia.uploaded_by = $1 OR p.lead_id = $1 OR EXISTS (
      SELECT 1 FROM workspace_members wm
      WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1 AND wm.role = 'ADMIN'
  ))
  AND (sqlc.narg(search)::text IS NULL OR ia.original_name ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY ia.deleted_at DESC;
