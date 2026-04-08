-- name: CreateNoteOwner :one
SELECT note_id, owner_type, owner_id, created_at, updated_at, deleted_at
FROM fn_add_note_owner($1, $2, $3);

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

-- name: DeleteNoteOwner :one
SELECT note_id, owner_type, owner_id, created_at, updated_at, deleted_at
FROM fn_remove_note_owner($1, $2, $3);
