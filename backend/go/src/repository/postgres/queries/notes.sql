-- name: CreateNote :one
INSERT INTO notes (
  id, title, content_md, note_type, metadata_json, created_at, updated_at, deleted_at, version
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, title, content_md, note_type, metadata_json, created_at, updated_at, deleted_at, version;

-- name: GetNoteByID :one
SELECT id, title, content_md, note_type, metadata_json, created_at, updated_at, deleted_at, version
FROM notes
WHERE id = $1
  AND ($2::boolean OR deleted_at IS NULL);

-- name: ListNotes :many
SELECT id, title, content_md, note_type, metadata_json, created_at, updated_at, deleted_at, version
FROM notes
WHERE ($1::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC, id DESC
OFFSET $2
LIMIT $3;

-- name: UpdateNote :one
UPDATE notes
SET
  title = $2,
  content_md = $3,
  note_type = $4,
  metadata_json = $5,
  updated_at = $6,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL
  AND version = $7
RETURNING id, title, content_md, note_type, metadata_json, created_at, updated_at, deleted_at, version;

-- name: DeleteNote :execrows
UPDATE notes
SET
  deleted_at = $2,
  updated_at = $2,
  version = version + 1
WHERE id = $1
  AND deleted_at IS NULL;
