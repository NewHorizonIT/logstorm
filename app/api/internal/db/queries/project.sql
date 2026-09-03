-- name: CreateProject :one
INSERT INTO projects (owner_id, name, slug, description, environment)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1 AND owner_id = $2
LIMIT 1;

-- name: GetProjectByOwnerAndSlug :one
SELECT * FROM projects
WHERE owner_id = $1 AND slug = $2
LIMIT 1;

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET name = $3,
    description = $4,
    updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING *;
