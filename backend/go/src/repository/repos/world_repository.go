package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/postgres/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ interfaces.WorldRepository = (*WorldRepository)(nil)

type WorldRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewWorldRepository(db generated.DBTX) *WorldRepository {
	return &WorldRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *WorldRepository) Create(ctx context.Context, world model.World) (model.World, error) {
	status, err := toDBWorldStatus(world.Status)
	if err != nil {
		return model.World{}, repoerrors.WrapValidation("world.create", "world", err)
	}

	row, err := r.queries.CreateWorld(ctx, generated.CreateWorldParams{
		ID:          string(world.ID),
		Name:        world.Name,
		Description: world.Description,
		Status:      status,
		CreatedAt:   toPgTimestamptz(world.AuditFields.CreatedAt),
		UpdatedAt:   toPgTimestamptz(world.AuditFields.UpdatedAt),
		DeletedAt:   toNullablePgTimestamptz(world.AuditFields.DeletedAt),
		Version:     int32(world.AuditFields.Version),
	})
	if err != nil {
		return model.World{}, repoerrors.WrapUnknown("world.create", "world", err)
	}

	out, err := mapWorldRow(row)
	if err != nil {
		return model.World{}, repoerrors.WrapValidation("world.create", "world", err)
	}
	return out, nil
}

func (r *WorldRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error) {
	row, err := r.queries.GetWorldByID(ctx, generated.GetWorldByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.World{}, repoerrors.NewNotFound("world.get_by_id", "world")
		}
		return model.World{}, repoerrors.WrapUnknown("world.get_by_id", "world", err)
	}

	out, err := mapWorldRow(row)
	if err != nil {
		return model.World{}, repoerrors.WrapValidation("world.get_by_id", "world", err)
	}
	return out, nil
}

func (r *WorldRepository) List(ctx context.Context, params interfaces.ListWorldsParams) ([]model.World, error) {
	rows, err := r.queries.ListWorlds(ctx, generated.ListWorldsParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("world.list", "world", err)
	}

	worlds := make([]model.World, 0, len(rows))
	for _, row := range rows {
		w, mapErr := mapWorldRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("world.list", "world", mapErr)
		}
		worlds = append(worlds, w)
	}
	return worlds, nil
}

func (r *WorldRepository) Update(ctx context.Context, params interfaces.UpdateWorldParams) (model.World, error) {
	status, err := toDBWorldStatus(params.Status)
	if err != nil {
		return model.World{}, repoerrors.WrapValidation("world.update", "world", err)
	}

	row, err := r.queries.UpdateWorld(ctx, generated.UpdateWorldParams{
		ID:          string(params.ID),
		Name:        params.Name,
		Description: params.Description,
		Status:      status,
		UpdatedAt:   toPgTimestamptz(r.nowFn()),
		Version:     int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.World{}, repoerrors.NewConflict("world.update", "world")
		}
		return model.World{}, repoerrors.WrapUnknown("world.update", "world", err)
	}

	out, err := mapWorldRow(row)
	if err != nil {
		return model.World{}, repoerrors.WrapValidation("world.update", "world", err)
	}
	return out, nil
}

func (r *WorldRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeleteWorld(ctx, generated.DeleteWorldParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("world.delete", "world", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("world.delete", "world")
	}
	return nil
}

func mapWorldRow(row generated.World) (model.World, error) {
	status, err := fromDBWorldStatus(row.Status)
	if err != nil {
		return model.World{}, err
	}

	return model.World{
		ID:          model.ULID(fmt.Sprint(row.ID)),
		Name:        row.Name,
		Description: row.Description,
		Status:      status,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}

func toDBWorldStatus(status constants.WorldStatus) (generated.WorldStatus, error) {
	switch status {
	case constants.Draft:
		return generated.WorldStatusDraft, nil
	case constants.Active:
		return generated.WorldStatusActive, nil
	case constants.Archived:
		return generated.WorldStatusArchived, nil
	default:
		return "", fmt.Errorf("invalid world status enum value: %d", status)
	}
}

func fromDBWorldStatus(status generated.WorldStatus) (constants.WorldStatus, error) {
	switch status {
	case generated.WorldStatusDraft:
		return constants.Draft, nil
	case generated.WorldStatusActive:
		return constants.Active, nil
	case generated.WorldStatusArchived:
		return constants.Archived, nil
	default:
		return constants.Draft, fmt.Errorf("invalid world status db value: %q", status)
	}
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t.UTC(),
		Valid: true,
	}
}

func toNullablePgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{
		Time:  t.UTC(),
		Valid: true,
	}
}

func fromPgTimestamptzOrZero(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time.UTC()
}

func fromNullablePgTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	value := t.Time.UTC()
	return &value
}
