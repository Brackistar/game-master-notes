-- name: CreateTag :one
INSERT INTO tags (
  id, name, campaign_id, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, name, campaign_id, created_at, updated_at, deleted_at, version;

-- name: GetTagByID :one
SELECT id, name, campaign_id, created_at, updated_at, deleted_at, version
FROM tags
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListTags :many
SELECT id, name, campaign_id, created_at, updated_at, deleted_at, version
FROM tags
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdateTag :one
UPDATE tags
SET
  name = $2,
  campaign_id = $3,
  updated_at = $4,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $5
RETURNING id, name, campaign_id, created_at, updated_at, deleted_at, version;

-- name: DeleteTag :execrows
UPDATE tags
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
