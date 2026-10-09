package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListPlayersParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type SearchPlayersParams struct {
	Query          string
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdatePlayerParams struct {
	ID              model.ULID
	Name            string
	ExpectedVersion model.Version
}

type PlayerRepository interface {
	Create(ctx context.Context, player model.Player) (model.Player, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error)
	List(ctx context.Context, params ListPlayersParams) ([]model.Player, error)
	SearchByName(ctx context.Context, params SearchPlayersParams) ([]model.Player, error)
	Update(ctx context.Context, params UpdatePlayerParams) (model.Player, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
	Restore(ctx context.Context, id model.ULID) (model.Player, error)
}
