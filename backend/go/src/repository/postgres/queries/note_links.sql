-- name: CreateNoteLink :one
SELECT id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version
FROM fn_add_note_link($1, $2, $3, $4);

-- name: GetNoteLinkByID :one
SELECT id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version
FROM note_links
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListNoteLinksBySource :many
SELECT id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version
FROM note_links
WHERE source_note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $3
LIMIT $4;

-- name: ListNoteLinksByTarget :many
SELECT id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version
FROM note_links
WHERE target_note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $3
LIMIT $4;

-- name: UpdateNoteLink :one
UPDATE note_links
SET
  link_type = $2,
  updated_at = $3,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $4
RETURNING id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version;

-- name: DeleteNoteLink :one
SELECT id, source_note_id, target_note_id, link_type, created_at, updated_at, deleted_at, version
FROM fn_remove_note_link($1, $2, $3);
