-- name: CreatePlane :one
INSERT INTO planes (
  id, name, description, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, name, description, created_at, updated_at, deleted_at, version;

-- name: GetPlaneByID :one
SELECT id, name, description, created_at, updated_at, deleted_at, version
FROM planes
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListPlanes :many
SELECT id, name, description, created_at, updated_at, deleted_at, version
FROM planes
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdatePlane :one
UPDATE planes
SET
  name = $2,
  description = $3,
  updated_at = $4,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $5
RETURNING id, name, description, created_at, updated_at, deleted_at, version;

-- name: DeletePlane :execrows
UPDATE planes
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
