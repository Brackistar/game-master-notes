-- name: CreateMapNotePlacement :one
SELECT id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version
FROM fn_upsert_map_note_placement($1, $2, $3, $4, $5);

-- name: GetMapNotePlacementByID :one
SELECT id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version
FROM map_note_placements
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListMapNotePlacementsByMap :many
SELECT id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version
FROM map_note_placements
WHERE map_note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $3
LIMIT $4;

-- name: ListMapNotePlacementsByTarget :many
SELECT id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version
FROM map_note_placements
WHERE target_note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $3
LIMIT $4;

-- name: UpdateMapNotePlacement :one
UPDATE map_note_placements
SET
  x = $2,
  y = $3,
  updated_at = $4,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $5
RETURNING id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version;

-- name: DeleteMapNotePlacement :one
SELECT id, map_note_id, target_note_id, x, y, created_at, updated_at, deleted_at, version
FROM fn_remove_map_note_placement($1, $2);
