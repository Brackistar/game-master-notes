package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListPlanesParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdatePlaneParams struct {
	ID              model.ULID
	Name            string
	Description     string
	ExpectedVersion model.Version
}

type PlaneRepository interface {
	Create(ctx context.Context, plane model.Plane) (model.Plane, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error)
	List(ctx context.Context, params ListPlanesParams) ([]model.Plane, error)
	Update(ctx context.Context, params UpdatePlaneParams) (model.Plane, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
