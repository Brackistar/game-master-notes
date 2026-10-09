-- name: CreateCampaign :one
INSERT INTO campaigns (
  id, world_id, name, start_date, end_date, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, world_id, name, start_date, end_date, created_at, updated_at, deleted_at, version;

-- name: GetCampaignByID :one
SELECT id, world_id, name, start_date, end_date, created_at, updated_at, deleted_at, version
FROM campaigns
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListCampaigns :many
SELECT id, world_id, name, start_date, end_date, created_at, updated_at, deleted_at, version
FROM campaigns
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdateCampaign :one
UPDATE campaigns
SET
  name = $2,
  start_date = $3,
  end_date = $4,
  updated_at = $5,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $6
RETURNING id, world_id, name, start_date, end_date, created_at, updated_at, deleted_at, version;

-- name: DeleteCampaign :execrows
UPDATE campaigns
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
