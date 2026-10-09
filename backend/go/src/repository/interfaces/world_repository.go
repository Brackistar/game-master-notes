package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type ListWorldsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateWorldParams struct {
	ID              model.ULID
	Name            string
	Description     string
	Status          constants.WorldStatus
	ExpectedVersion model.Version
}

type WorldRepository interface {
	Create(ctx context.Context, world model.World) (model.World, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error)
	List(ctx context.Context, params ListWorldsParams) ([]model.World, error)
	Update(ctx context.Context, params UpdateWorldParams) (model.World, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
