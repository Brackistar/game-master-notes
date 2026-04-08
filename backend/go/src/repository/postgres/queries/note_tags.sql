-- name: CreateNoteTag :one
SELECT note_id, tag_id, created_at, updated_at, deleted_at
FROM fn_add_note_tag($1, $2);

-- name: GetNoteTag :one
SELECT note_id, tag_id, created_at, updated_at, deleted_at
FROM note_tags
WHERE note_id = $1
  AND tag_id = $2
  AND ($3::boolean OR deleted_at IS NULL);

-- name: ListNoteTagsByNote :many
SELECT note_id, tag_id, created_at, updated_at, deleted_at
FROM note_tags
WHERE note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $3
LIMIT $4;

-- name: ListNoteTagsByTag :many
SELECT note_id, tag_id, created_at, updated_at, deleted_at
FROM note_tags
WHERE tag_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $3
LIMIT $4;

-- name: DeleteNoteTag :one
SELECT note_id, tag_id, created_at, updated_at, deleted_at
FROM fn_remove_note_tag($1, $2);
