-- name: CreateCampaignPlayer :one
INSERT INTO campaign_players (
  campaign_id, player_id, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING campaign_id, player_id, created_at, updated_at, deleted_at;

-- name: GetCampaignPlayer :one
SELECT campaign_id, player_id, created_at, updated_at, deleted_at
FROM campaign_players
WHERE campaign_id = $1
  AND player_id = $2
  AND ($3::boolean OR deleted_at IS NULL);

-- name: ListCampaignPlayersByCampaign :many
SELECT campaign_id, player_id, created_at, updated_at, deleted_at
FROM campaign_players
WHERE campaign_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $3
LIMIT $4;

-- name: ListCampaignPlayersByPlayer :many
SELECT campaign_id, player_id, created_at, updated_at, deleted_at
FROM campaign_players
WHERE player_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $3
LIMIT $4;

-- name: DeleteCampaignPlayer :execrows
UPDATE campaign_players
SET
  deleted_at = $3,
  updated_at = $3
WHERE campaign_id = $1
  AND player_id = $2
  AND deleted_at IS NULL;

