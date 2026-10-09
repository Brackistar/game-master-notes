package interfaces

import (
	"context"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListCampaignsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateCampaignParams struct {
	ID              model.ULID
	Name            string
	StartDate       *time.Time
	EndDate         *time.Time
	ExpectedVersion model.Version
}

type CampaignRepository interface {
	Create(ctx context.Context, campaign model.Campaign) (model.Campaign, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error)
	List(ctx context.Context, params ListCampaignsParams) ([]model.Campaign, error)
	Update(ctx context.Context, params UpdateCampaignParams) (model.Campaign, error)
	Delete(ctx context.Context, id model.ULID) error // soft delete
}
