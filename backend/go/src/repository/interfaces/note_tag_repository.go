package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListNoteTagsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type NoteTagRepository interface {
	Create(ctx context.Context, rel model.NoteTag) (model.NoteTag, error)
	Get(ctx context.Context, noteID, tagID model.ULID, includeDeleted bool) (model.NoteTag, error)
	ListByNote(ctx context.Context, noteID model.ULID, params ListNoteTagsParams) ([]model.NoteTag, error)
	ListByTag(ctx context.Context, tagID model.ULID, params ListNoteTagsParams) ([]model.NoteTag, error)
	Delete(ctx context.Context, noteID, tagID model.ULID) error
}
