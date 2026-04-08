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
	"github.com/jackc/pgx/v5"
)

var _ interfaces.MapNotePlacementRepository = (*MapNotePlacementRepository)(nil)

type MapNotePlacementRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewMapNotePlacementRepository(db generated.DBTX) *MapNotePlacementRepository {
	return &MapNotePlacementRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *MapNotePlacementRepository) Create(ctx context.Context, placement model.MapNotePlacement) (model.MapNotePlacement, error) {
	row, err := r.queries.CreateMapNotePlacement(ctx, generated.CreateMapNotePlacementParams{
		ID:           string(placement.ID),
		MapNoteID:    string(placement.MapNoteID),
		TargetNoteID: string(placement.TargetNoteID),
		X:            int16(placement.X),
		Y:            int16(placement.Y),
		CreatedAt:    toPgTimestamptz(placement.AuditFields.CreatedAt),
		UpdatedAt:    toPgTimestamptz(placement.AuditFields.UpdatedAt),
		DeletedAt:    toNullablePgTimestamptz(placement.AuditFields.DeletedAt),
		Version:      int32(placement.AuditFields.Version),
	})
	if err != nil {
		return model.MapNotePlacement{}, repoerrors.WrapUnknown("map_note_placement.create", "map_note_placement", err)
	}
	return mapMapNotePlacementRow(row), nil
}

func (r *MapNotePlacementRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.MapNotePlacement, error) {
	row, err := r.queries.GetMapNotePlacementByID(ctx, generated.GetMapNotePlacementByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.MapNotePlacement{}, repoerrors.NewNotFound("map_note_placement.get_by_id", "map_note_placement")
		}
		return model.MapNotePlacement{}, repoerrors.WrapUnknown("map_note_placement.get_by_id", "map_note_placement", err)
	}
	return mapMapNotePlacementRow(row), nil
}

func (r *MapNotePlacementRepository) ListByMapNote(ctx context.Context, mapNoteID model.ULID, params interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
	rows, err := r.queries.ListMapNotePlacementsByMap(ctx, generated.ListMapNotePlacementsByMapParams{
		MapNoteID: string(mapNoteID),
		Column2:   params.IncludeDeleted,
		Offset:    params.Offset,
		Limit:     params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("map_note_placement.list_by_map_note", "map_note_placement", err)
	}
	out := make([]model.MapNotePlacement, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapMapNotePlacementRow(row))
	}
	return out, nil
}

func (r *MapNotePlacementRepository) ListByTargetNote(ctx context.Context, targetNoteID model.ULID, params interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
	rows, err := r.queries.ListMapNotePlacementsByTarget(ctx, generated.ListMapNotePlacementsByTargetParams{
		TargetNoteID: string(targetNoteID),
		Column2:      params.IncludeDeleted,
		Offset:       params.Offset,
		Limit:        params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("map_note_placement.list_by_target_note", "map_note_placement", err)
	}
	out := make([]model.MapNotePlacement, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapMapNotePlacementRow(row))
	}
	return out, nil
}

func (r *MapNotePlacementRepository) Update(ctx context.Context, params interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
	row, err := r.queries.UpdateMapNotePlacement(ctx, generated.UpdateMapNotePlacementParams{
		ID:        string(params.ID),
		X:         int16(params.X),
		Y:         int16(params.Y),
		UpdatedAt: toPgTimestamptz(r.nowFn()),
		Version:   int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.MapNotePlacement{}, repoerrors.NewConflict("map_note_placement.update", "map_note_placement")
		}
		return model.MapNotePlacement{}, repoerrors.WrapUnknown("map_note_placement.update", "map_note_placement", err)
	}
	return mapMapNotePlacementRow(row), nil
}

func (r *MapNotePlacementRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeleteMapNotePlacement(ctx, generated.DeleteMapNotePlacementParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("map_note_placement.delete", "map_note_placement", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("map_note_placement.delete", "map_note_placement")
	}
	return nil
}

func mapMapNotePlacementRow(row generated.MapNotePlacement) model.MapNotePlacement {
	return model.MapNotePlacement{
		ID:           model.ULID(fmt.Sprint(row.ID)),
		MapNoteID:    model.ULID(fmt.Sprint(row.MapNoteID)),
		TargetNoteID: model.ULID(fmt.Sprint(row.TargetNoteID)),
		X:            uint8(row.X),
		Y:            uint8(row.Y),
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}
}
