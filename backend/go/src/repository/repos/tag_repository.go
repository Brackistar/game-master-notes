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
)

var _ interfaces.TagRepository = (*TagRepository)(nil)

type TagRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewTagRepository(db generated.DBTX) *TagRepository {
	return &TagRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *TagRepository) Create(ctx context.Context, tag model.Tag) (model.Tag, error) {
	row, err := r.queries.CreateTag(ctx, generated.CreateTagParams{
		ID:         string(tag.ID),
		Name:       tag.Name,
		CampaignID: nullableULIDToAny(tag.CampaignID),
		CreatedAt:  toPgTimestamptz(tag.AuditFields.CreatedAt),
		UpdatedAt:  toPgTimestamptz(tag.AuditFields.UpdatedAt),
		DeletedAt:  toNullablePgTimestamptz(tag.AuditFields.DeletedAt),
		Version:    int32(tag.AuditFields.Version),
	})
	if err != nil {
		return model.Tag{}, repoerrors.WrapUnknown("tag.create", "tag", err)
	}
	return mapTagRow(row), nil
}

func (r *TagRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error) {
	row, err := r.queries.GetTagByID(ctx, generated.GetTagByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tag{}, repoerrors.NewNotFound("tag.get_by_id", "tag")
		}
		return model.Tag{}, repoerrors.WrapUnknown("tag.get_by_id", "tag", err)
	}
	return mapTagRow(row), nil
}

func (r *TagRepository) List(ctx context.Context, params interfaces.ListTagsParams) ([]model.Tag, error) {
	rows, err := r.queries.ListTags(ctx, generated.ListTagsParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("tag.list", "tag", err)
	}
	out := make([]model.Tag, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTagRow(row))
	}
	return out, nil
}

func (r *TagRepository) Update(ctx context.Context, params interfaces.UpdateTagParams) (model.Tag, error) {
	row, err := r.queries.UpdateTag(ctx, generated.UpdateTagParams{
		ID:         string(params.ID),
		Name:       params.Name,
		CampaignID: nullableULIDToAny(params.CampaignID),
		UpdatedAt:  toPgTimestamptz(r.nowFn()),
		Version:    int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Tag{}, repoerrors.NewConflict("tag.update", "tag")
		}
		return model.Tag{}, repoerrors.WrapUnknown("tag.update", "tag", err)
	}
	return mapTagRow(row), nil
}

func (r *TagRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeleteTag(ctx, generated.DeleteTagParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("tag.delete", "tag", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("tag.delete", "tag")
	}
	return nil
}

func mapTagRow(row generated.Tag) model.Tag {
	return model.Tag{
		ID:         model.ULID(fmt.Sprint(row.ID)),
		Name:       row.Name,
		CampaignID: anyToNullableULID(row.CampaignID),
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}
}

func nullableULIDToAny(id *model.ULID) interface{} {
	if id == nil {
		return nil
	}
	return string(*id)
}

func anyToNullableULID(value interface{}) *model.ULID {
	if value == nil {
		return nil
	}
	id := model.ULID(fmt.Sprint(value))
	return &id
}
