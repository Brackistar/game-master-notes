package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type ListNoteOwnersParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type NoteOwnerRepository interface {
	Create(ctx context.Context, rel model.NoteOwner) (model.NoteOwner, error)
	Get(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, includeDeleted bool) (model.NoteOwner, error)
	ListByNote(ctx context.Context, noteID model.ULID, params ListNoteOwnersParams) ([]model.NoteOwner, error)
	ListByOwner(ctx context.Context, ownerType constants.OwnerType, ownerID model.ULID, params ListNoteOwnersParams) ([]model.NoteOwner, error)
	Delete(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID) error
}
