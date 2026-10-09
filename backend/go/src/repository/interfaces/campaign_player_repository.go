package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
)

type ListCampaignPlayersParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type CampaignPlayerRepository interface {
	Create(ctx context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error)
	Get(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error)
	ListByCampaign(ctx context.Context, campaignID model.ULID, params ListCampaignPlayersParams) ([]model.CampaignPlayer, error)
	ListByPlayer(ctx context.Context, playerID model.ULID, params ListCampaignPlayersParams) ([]model.CampaignPlayer, error)
	Delete(ctx context.Context, campaignID, playerID model.ULID) error
}
