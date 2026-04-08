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

var _ interfaces.PlaneRepository = (*PlaneRepository)(nil)

type PlaneRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewPlaneRepository(db generated.DBTX) *PlaneRepository {
	return &PlaneRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *PlaneRepository) Create(ctx context.Context, plane model.Plane) (model.Plane, error) {
	row, err := r.queries.CreatePlane(ctx, generated.CreatePlaneParams{
		ID:          string(plane.ID),
		WorldID:     string(plane.WorldID),
		Name:        plane.Name,
		Description: plane.Description,
		CreatedAt:   toPgTimestamptz(plane.AuditFields.CreatedAt),
		UpdatedAt:   toPgTimestamptz(plane.AuditFields.UpdatedAt),
		DeletedAt:   toNullablePgTimestamptz(plane.AuditFields.DeletedAt),
		Version:     int32(plane.AuditFields.Version),
	})
	if err != nil {
		return model.Plane{}, repoerrors.WrapUnknown("plane.create", "plane", err)
	}

	return mapPlaneRow(row), nil
}

func (r *PlaneRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error) {
	row, err := r.queries.GetPlaneByID(ctx, generated.GetPlaneByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Plane{}, repoerrors.NewNotFound("plane.get_by_id", "plane")
		}
		return model.Plane{}, repoerrors.WrapUnknown("plane.get_by_id", "plane", err)
	}

	return mapPlaneRow(row), nil
}

func (r *PlaneRepository) List(ctx context.Context, params interfaces.ListPlanesParams) ([]model.Plane, error) {
	rows, err := r.queries.ListPlanes(ctx, generated.ListPlanesParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("plane.list", "plane", err)
	}

	out := make([]model.Plane, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapPlaneRow(row))
	}
	return out, nil
}

func (r *PlaneRepository) Update(ctx context.Context, params interfaces.UpdatePlaneParams) (model.Plane, error) {
	row, err := r.queries.UpdatePlane(ctx, generated.UpdatePlaneParams{
		ID:          string(params.ID),
		Name:        params.Name,
		Description: params.Description,
		UpdatedAt:   toPgTimestamptz(r.nowFn()),
		Version:     int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Plane{}, repoerrors.NewConflict("plane.update", "plane")
		}
		return model.Plane{}, repoerrors.WrapUnknown("plane.update", "plane", err)
	}
	return mapPlaneRow(row), nil
}

func (r *PlaneRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeletePlane(ctx, generated.DeletePlaneParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("plane.delete", "plane", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("plane.delete", "plane")
	}
	return nil
}

func mapPlaneRow(row generated.Plane) model.Plane {
	return model.Plane{
		ID:          model.ULID(fmt.Sprint(row.ID)),
		WorldID:     model.ULID(fmt.Sprint(row.WorldID)),
		Name:        row.Name,
		Description: row.Description,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}
}
