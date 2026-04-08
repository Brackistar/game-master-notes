-- name: CreatePlayer :one
INSERT INTO players (
  id, name, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, name, created_at, updated_at, deleted_at, version;

-- name: GetPlayerByID :one
SELECT id, name, created_at, updated_at, deleted_at, version
FROM players
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListPlayers :many
SELECT id, name, created_at, updated_at, deleted_at, version
FROM players
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdatePlayer :one
UPDATE players
SET
  name = $2,
  updated_at = $3,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $4
RETURNING id, name, created_at, updated_at, deleted_at, version;

-- name: DeletePlayer :execrows
UPDATE players
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
