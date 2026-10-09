package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type ListNotesParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateNoteParams struct {
	ID              model.ULID
	Title           string
	ContentMD       string
	NoteType        constants.NoteType
	MetadataJSON    []byte
	ExpectedVersion model.Version
}

type NoteRepository interface {
	Create(ctx context.Context, note model.Note) (model.Note, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Note, error)
	List(ctx context.Context, params ListNotesParams) ([]model.Note, error)
	Update(ctx context.Context, params UpdateNoteParams) (model.Note, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
