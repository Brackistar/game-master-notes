-- name: CreateNoteAsset :one
INSERT INTO note_assets (
  id, note_id, asset_type, storage_path, mime_type, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, note_id, asset_type, storage_path, mime_type, created_at, updated_at, deleted_at, version;

-- name: GetNoteAssetByID :one
SELECT id, note_id, asset_type, storage_path, mime_type, created_at, updated_at, deleted_at, version
FROM note_assets
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListNoteAssetsByNote :many
SELECT id, note_id, asset_type, storage_path, mime_type, created_at, updated_at, deleted_at, version
FROM note_assets
WHERE note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $3
LIMIT $4;

-- name: UpdateNoteAsset :one
UPDATE note_assets
SET
  asset_type = $2,
  storage_path = $3,
  mime_type = $4,
  updated_at = $5,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $6
RETURNING id, note_id, asset_type, storage_path, mime_type, created_at, updated_at, deleted_at, version;

-- name: DeleteNoteAsset :execrows
UPDATE note_assets
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;

