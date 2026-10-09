package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListMapNotePlacementsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateMapNotePlacementParams struct {
	ID              model.ULID
	X               uint8
	Y               uint8
	ExpectedVersion model.Version
}

type MapNotePlacementRepository interface {
	Create(ctx context.Context, placement model.MapNotePlacement) (model.MapNotePlacement, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.MapNotePlacement, error)
	ListByMapNote(ctx context.Context, mapNoteID model.ULID, params ListMapNotePlacementsParams) ([]model.MapNotePlacement, error)
	ListByTargetNote(ctx context.Context, targetNoteID model.ULID, params ListMapNotePlacementsParams) ([]model.MapNotePlacement, error)
	Update(ctx context.Context, params UpdateMapNotePlacementParams) (model.MapNotePlacement, error)
	Delete(ctx context.Context, id model.ULID) error
}
