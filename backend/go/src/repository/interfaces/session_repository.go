package interfaces

import (
	"context"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListSessionsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateSessionParams struct {
	ID              model.ULID
	PlayedOn        *time.Time
	SummaryMD       string
	ExpectedVersion model.Version
}

type SessionRepository interface {
	Create(ctx context.Context, session model.Session) (model.Session, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error)
	List(ctx context.Context, params ListSessionsParams) ([]model.Session, error)
	Update(ctx context.Context, params UpdateSessionParams) (model.Session, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
