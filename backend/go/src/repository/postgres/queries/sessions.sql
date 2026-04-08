-- name: CreateSession :one
INSERT INTO sessions (
  id, campaign_id, played_on, summary_md, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, campaign_id, played_on, summary_md, created_at, updated_at, deleted_at, version;

-- name: GetSessionByID :one
SELECT id, campaign_id, played_on, summary_md, created_at, updated_at, deleted_at, version
FROM sessions
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListSessions :many
SELECT id, campaign_id, played_on, summary_md, created_at, updated_at, deleted_at, version
FROM sessions
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY played_on DESC, created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdateSession :one
UPDATE sessions
SET
  played_on = $2,
  summary_md = $3,
  updated_at = $4,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $5
RETURNING id, campaign_id, played_on, summary_md, created_at, updated_at, deleted_at, version;

-- name: DeleteSession :execrows
UPDATE sessions
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
