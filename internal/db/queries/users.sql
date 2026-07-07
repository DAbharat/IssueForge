-- name: CreateOnboardingUser :one
INSERT INTO users(
    email,
    fullname,
    display_name,
    password_hash,
    role
)
VALUES(
    $1, $2, $3, $4, 'MEMBER'
)
RETURNING id, email, fullname, display_name, role, created_at;


--name: AssignUserToWorkspace :one
UPDATE users
SET workspace_id = $1,
    role = COALESCE($2, role)
WHERE id = $3 AND workspace_id IS NULL
RETURNING id, workspace_id, email, display_name, role;


--name: GetUserByEmailAndWorkspace :one
SELECT id, workspace_id, email, password_hash, role
FROM users
WHERE email = $1 AND workspace_id = $2


-- name: GetUserForLogin :one
SELECT id, workspace_id, email, password_hash, role
FROM users
WHERE email = $1
LIMIT 1;


-- name: GetUserByID :one
SELECT id, workspace_id, display_name, fullname, email, role, created_at, updated_at
FROM users
WHERE id=$1
LIMIT 1;