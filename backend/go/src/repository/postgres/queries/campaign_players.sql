-- name: CreateCampaignPlayer :one
SELECT campaign_id, player_id, created_at, updated_at, deleted_at
FROM fn_add_player_to_campaign($1, $2);

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

-- name: DeleteCampaignPlayer :one
SELECT campaign_id, player_id, created_at, updated_at, deleted_at
FROM fn_remove_player_from_campaign($1, $2);
