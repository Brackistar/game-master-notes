package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/postgres/generated"
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/repository/repos/shared"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ interfaces.CampaignRepository = (*CampaignRepository)(nil)

type CampaignRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewCampaignRepository(db generated.DBTX) *CampaignRepository {
	return &CampaignRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *CampaignRepository) Create(ctx context.Context, campaign model.Campaign) (model.Campaign, error) {
	row, err := r.queries.CreateCampaign(ctx, generated.CreateCampaignParams{
		ID:        string(campaign.ID),
		WorldID:   string(campaign.WorldID),
		Name:      campaign.Name,
		StartDate: toNullablePgDate(campaign.StartDate),
		EndDate:   toNullablePgDate(campaign.EndDate),
		CreatedAt: toPgTimestamptz(campaign.AuditFields.CreatedAt),
		UpdatedAt: toPgTimestamptz(campaign.AuditFields.UpdatedAt),
		DeletedAt: toNullablePgTimestamptz(campaign.AuditFields.DeletedAt),
		Version:   int32(campaign.AuditFields.Version),
	})
	if err != nil {
		return model.Campaign{}, repoerrors.WrapUnknown("campaign.create", "campaign", err)
	}

	out, err := mapCampaignRow(row)
	if err != nil {
		return model.Campaign{}, repoerrors.WrapValidation("campaign.create", "campaign", err)
	}
	return out, nil
}

func (r *CampaignRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error) {
	row, err := r.queries.GetCampaignByID(ctx, generated.GetCampaignByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Campaign{}, repoerrors.NewNotFound("campaign.get_by_id", "campaign")
		}
		return model.Campaign{}, repoerrors.WrapUnknown("campaign.get_by_id", "campaign", err)
	}

	out, err := mapCampaignRow(row)
	if err != nil {
		return model.Campaign{}, repoerrors.WrapValidation("campaign.get_by_id", "campaign", err)
	}
	return out, nil
}

func (r *CampaignRepository) List(ctx context.Context, params interfaces.ListCampaignsParams) ([]model.Campaign, error) {
	rows, err := r.queries.ListCampaigns(ctx, generated.ListCampaignsParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("campaign.list", "campaign", err)
	}

	campaigns := make([]model.Campaign, 0, len(rows))
	for _, row := range rows {
		c, mapErr := mapCampaignRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("campaign.list", "campaign", mapErr)
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

func (r *CampaignRepository) Update(ctx context.Context, params interfaces.UpdateCampaignParams) (model.Campaign, error) {
	row, err := r.queries.UpdateCampaign(ctx, generated.UpdateCampaignParams{
		ID:        string(params.ID),
		Name:      params.Name,
		StartDate: toNullablePgDate(params.StartDate),
		EndDate:   toNullablePgDate(params.EndDate),
		UpdatedAt: toPgTimestamptz(r.nowFn()),
		Version:   int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Campaign{}, repoerrors.NewConflict("campaign.update", "campaign")
		}
		return model.Campaign{}, repoerrors.WrapUnknown("campaign.update", "campaign", err)
	}

	out, err := mapCampaignRow(row)
	if err != nil {
		return model.Campaign{}, repoerrors.WrapValidation("campaign.update", "campaign", err)
	}
	return out, nil
}

func (r *CampaignRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeleteCampaign(ctx, generated.DeleteCampaignParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("campaign.delete", "campaign", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("campaign.delete", "campaign")
	}
	return nil
}

func mapCampaignRow(row generated.Campaign) (model.Campaign, error) {
	return model.Campaign{
		ID:        model.ULID(fmt.Sprint(row.ID)),
		WorldID:   model.ULID(fmt.Sprint(row.WorldID)),
		Name:      row.Name,
		StartDate: fromNullablePgDate(row.StartDate),
		EndDate:   fromNullablePgDate(row.EndDate),
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}

func toNullablePgDate(d *time.Time) pgtype.Date {
	if d == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{
		Time:  d.UTC(),
		Valid: true,
	}
}

func fromNullablePgDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	value := d.Time.UTC()
	return &value
}
