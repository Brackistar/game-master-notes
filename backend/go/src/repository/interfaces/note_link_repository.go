package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type ListNoteLinksParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateNoteLinkParams struct {
	ID              model.ULID
	LinkType        constants.NoteLinkType
	ExpectedVersion model.Version
}

type NoteLinkRepository interface {
	Create(ctx context.Context, link model.NoteLink) (model.NoteLink, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteLink, error)
	ListBySource(ctx context.Context, sourceNoteID model.ULID, params ListNoteLinksParams) ([]model.NoteLink, error)
	ListByTarget(ctx context.Context, targetNoteID model.ULID, params ListNoteLinksParams) ([]model.NoteLink, error)
	Update(ctx context.Context, params UpdateNoteLinkParams) (model.NoteLink, error)
	Delete(ctx context.Context, id model.ULID) error
}

