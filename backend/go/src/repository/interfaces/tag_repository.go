package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListTagsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateTagParams struct {
	ID              model.ULID
	Name            string
	CampaignID      *model.ULID
	ExpectedVersion model.Version
}

type TagRepository interface {
	Create(ctx context.Context, tag model.Tag) (model.Tag, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error)
	List(ctx context.Context, params ListTagsParams) ([]model.Tag, error)
	Update(ctx context.Context, params UpdateTagParams) (model.Tag, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
