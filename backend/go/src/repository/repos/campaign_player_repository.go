package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/postgres/generated"
	"github.com/jackc/pgx/v5"
)

var _ interfaces.CampaignPlayerRepository = (*CampaignPlayerRepository)(nil)

type CampaignPlayerRepository struct {
	queries *generated.Queries
}

func NewCampaignPlayerRepository(db generated.DBTX) *CampaignPlayerRepository {
	return &CampaignPlayerRepository{
		queries: generated.New(db),
	}
}

func (r *CampaignPlayerRepository) Create(ctx context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) {
	row, err := r.queries.CreateCampaignPlayer(ctx, generated.CreateCampaignPlayerParams{
		PCampaignID: string(rel.CampaignID),
		PPlayerID:   string(rel.PlayerID),
	})
	if err != nil {
		return model.CampaignPlayer{}, mapFunctionError(err, "campaign_player.create", "campaign_player",
			map[string]struct{}{
				"GMN_CAMPAIGN_NOT_FOUND": {},
				"GMN_PLAYER_NOT_FOUND":   {},
			},
			map[string]struct{}{
				"GMN_CAMPAIGN_PLAYER_ALREADY_ACTIVE": {},
			},
		)
	}
	return mapCampaignPlayerRow(row), nil
}

func (r *CampaignPlayerRepository) Get(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error) {
	row, err := r.queries.GetCampaignPlayer(ctx, generated.GetCampaignPlayerParams{
		CampaignID: string(campaignID),
		PlayerID:   string(playerID),
		Column3:    includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CampaignPlayer{}, repoerrors.NewNotFound("campaign_player.get", "campaign_player")
		}
		return model.CampaignPlayer{}, repoerrors.WrapUnknown("campaign_player.get", "campaign_player", err)
	}
	return mapCampaignPlayerRow(row), nil
}

func (r *CampaignPlayerRepository) ListByCampaign(ctx context.Context, campaignID model.ULID, params interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
	rows, err := r.queries.ListCampaignPlayersByCampaign(ctx, generated.ListCampaignPlayersByCampaignParams{
		CampaignID: string(campaignID),
		Column2:    params.IncludeDeleted,
		Offset:     params.Offset,
		Limit:      params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("campaign_player.list_by_campaign", "campaign_player", err)
	}
	out := make([]model.CampaignPlayer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCampaignPlayerRow(row))
	}
	return out, nil
}

func (r *CampaignPlayerRepository) ListByPlayer(ctx context.Context, playerID model.ULID, params interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
	rows, err := r.queries.ListCampaignPlayersByPlayer(ctx, generated.ListCampaignPlayersByPlayerParams{
		PlayerID: string(playerID),
		Column2:  params.IncludeDeleted,
		Offset:   params.Offset,
		Limit:    params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("campaign_player.list_by_player", "campaign_player", err)
	}
	out := make([]model.CampaignPlayer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCampaignPlayerRow(row))
	}
	return out, nil
}

func (r *CampaignPlayerRepository) Delete(ctx context.Context, campaignID, playerID model.ULID) error {
	_, err := r.queries.DeleteCampaignPlayer(ctx, generated.DeleteCampaignPlayerParams{
		PCampaignID: string(campaignID),
		PPlayerID:   string(playerID),
	})
	if err != nil {
		return mapFunctionError(err, "campaign_player.delete", "campaign_player",
			map[string]struct{}{
				"GMN_CAMPAIGN_PLAYER_NOT_ACTIVE": {},
			},
			nil,
		)
	}
	return nil
}

func mapCampaignPlayerRow(row generated.CampaignPlayer) model.CampaignPlayer {
	return model.CampaignPlayer{
		CampaignID: model.ULID(fmt.Sprint(row.CampaignID)),
		PlayerID:   model.ULID(fmt.Sprint(row.PlayerID)),
		CreatedAt:  fromPgTimestamptzOrZero(row.CreatedAt),
		UpdatedAt:  fromPgTimestamptzOrZero(row.UpdatedAt),
		DeletedAt:  fromNullablePgTimestamptz(row.DeletedAt),
	}
}
