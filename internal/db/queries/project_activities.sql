-- name: CreateProjectActivity :one
INSERT INTO project_activities(
    project_id, 
    actor_id, 
    target_user_id, 
    activity_type, 
    field_name, 
    old_value, 
    new_value
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, project_id, actor_id, target_user_id, activity_type, field_name, old_value, new_value, created_at;


-- name: ListProjectActivities :many
SELECT pa.id, pa.project_id, pa.actor_id, pa.target_user_id, pa.activity_type, pa.field_name, pa.old_value, pa.new_value, pa.created_at,
        u_actor.fullname, u_actor.username, u_target_user.fullname, u_target_user.username
FROM project_activities pa
INNER JOIN users u_actor ON pa.actor_id = u_actor.id
LEFT JOIN users u_target_user on pa.target_user_id = u_target_user.id
WHERE pa.project_id = $1
ORDER BY pa.created_at DESC
LIMIT $2
OFFSET $3;