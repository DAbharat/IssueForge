-- name: CreateActivity :one
INSERT INTO issue_activities(
    issue_id, 
    actor_id, 
    activity_type, 
    field_name, 
    old_value, 
    new_value
) VALUES (
    $1,$2,$3,$4,$5,$6
)
RETURNING id, issue_id, actor_id, activity_type, field_name, old_value, new_value, created_at;


-- name: ListIssueActivities :many
SELECT ia.id, ia.issue_id, ia.actor_id, ia.activity_type, ia.field_name, ia.old_value, ia.new_value, ia.created_at,
       u_actor.fullname AS actor_name, u_actor.username AS actor_username
FROM issue_activities ia
INNER JOIN users u_actor ON ia.actor_id = u_actor.id
INNER JOIN issues i ON ia.issue_id = i.id
WHERE ia.issue_id = $1
ORDER BY ia.created_at DESC
LIMIT $2
OFFSET $3;