-- name: CreateNoteTag :one
INSERT INTO note_tags (
  note_id, tag_id, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING note_id, tag_id, created_at, updated_at, deleted_at;

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

-- name: DeleteNoteTag :execrows
UPDATE note_tags
SET
  deleted_at = $3,
  updated_at = $3
WHERE note_id = $1
  AND tag_id = $2
  AND deleted_at IS NULL;

