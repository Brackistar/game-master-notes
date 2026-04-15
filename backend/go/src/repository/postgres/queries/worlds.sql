-- name: CreateWorld :one
INSERT INTO worlds (
  id, plane_id, name, description, status, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, plane_id, name, description, status, created_at, updated_at, deleted_at, version;

-- name: GetWorldByID :one
SELECT id, plane_id, name, description, status, created_at, updated_at, deleted_at, version
FROM worlds
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListWorlds :many
SELECT id, plane_id, name, description, status, created_at, updated_at, deleted_at, version
FROM worlds
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdateWorld :one
UPDATE worlds
SET
  name = $2,
  description = $3,
  status = $4,
  updated_at = $5,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $6
RETURNING id, plane_id, name, description, status, created_at, updated_at, deleted_at, version;

-- name: DeleteWorld :execrows
UPDATE worlds
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
