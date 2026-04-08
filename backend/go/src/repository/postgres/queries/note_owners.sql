-- name: CreateNoteOwner :one
INSERT INTO note_owners (
  note_id, owner_type, owner_id, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING note_id, owner_type, owner_id, created_at, updated_at, deleted_at;

-- name: GetNoteOwner :one
SELECT note_id, owner_type, owner_id, created_at, updated_at, deleted_at
FROM note_owners
WHERE note_id = $1
  AND owner_type = $2
  AND owner_id = $3
  AND ($4::boolean OR deleted_at IS NULL);

-- name: ListNoteOwnersByNote :many
SELECT note_id, owner_type, owner_id, created_at, updated_at, deleted_at
FROM note_owners
WHERE note_id = $1
  AND ($2::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $3
LIMIT $4;

-- name: ListNoteOwnersByOwner :many
SELECT note_id, owner_type, owner_id, created_at, updated_at, deleted_at
FROM note_owners
WHERE owner_type = $1
  AND owner_id = $2
  AND ($3::boolean OR deleted_at IS NULL)
ORDER BY created_at DESC
OFFSET $4
LIMIT $5;

-- name: DeleteNoteOwner :execrows
UPDATE note_owners
SET
  deleted_at = $4,
  updated_at = $4
WHERE note_id = $1
  AND owner_type = $2
  AND owner_id = $3
  AND deleted_at IS NULL;

