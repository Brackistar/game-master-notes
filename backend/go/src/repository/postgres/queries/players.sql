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
ORDER BY lower(name) ASC, id ASC
OFFSET $2
LIMIT $3;

-- name: SearchPlayersByName :many
SELECT id, name, created_at, updated_at, deleted_at, version
FROM players
WHERE ($1::boolean OR deleted_at IS NULL)
  AND lower(name) LIKE '%' || lower($2) || '%'
ORDER BY lower(name) ASC, id ASC
OFFSET $3
LIMIT $4;

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

-- name: RestorePlayer :one
UPDATE players
SET
  deleted_at = NULL,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NOT NULL
RETURNING id, name, created_at, updated_at, deleted_at, version;
