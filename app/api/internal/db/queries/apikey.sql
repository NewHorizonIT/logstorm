-- name: CreateAPIKey :one
INSERT INTO api_keys (project_id, owner_id, name, key_hash, key_prefix)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1
  AND revoked_at IS NULL
LIMIT 1;

-- name: ListAPIKeysByOwner :many
SELECT * FROM api_keys
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: ListAPIKeysByProject :many
SELECT * FROM api_keys
WHERE project_id = $1
  AND owner_id = $2
ORDER BY created_at DESC;

-- name: RevokeAPIKey :one
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1
  AND owner_id = $2
  AND revoked_at IS NULL
RETURNING *;
