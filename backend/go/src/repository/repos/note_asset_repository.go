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
)

var _ interfaces.NoteAssetRepository = (*NoteAssetRepository)(nil)

type NoteAssetRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewNoteAssetRepository(db generated.DBTX) *NoteAssetRepository {
	return &NoteAssetRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *NoteAssetRepository) Create(ctx context.Context, asset model.NoteAsset) (model.NoteAsset, error) {
	assetType, err := toDBAssetType(asset.AssetType)
	if err != nil {
		return model.NoteAsset{}, repoerrors.WrapValidation("note_asset.create", "note_asset", err)
	}
	row, err := r.queries.CreateNoteAsset(ctx, generated.CreateNoteAssetParams{
		ID:          string(asset.ID),
		NoteID:      string(asset.NoteID),
		AssetType:   assetType,
		StoragePath: asset.StoragePath,
		MimeType:    asset.MIMEType,
		CreatedAt:   toPgTimestamptz(asset.AuditFields.CreatedAt),
		UpdatedAt:   toPgTimestamptz(asset.AuditFields.UpdatedAt),
		DeletedAt:   toNullablePgTimestamptz(asset.AuditFields.DeletedAt),
		Version:     int32(asset.AuditFields.Version),
	})
	if err != nil {
		return model.NoteAsset{}, repoerrors.WrapUnknown("note_asset.create", "note_asset", err)
	}
	return mapNoteAssetRow(row)
}

func (r *NoteAssetRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteAsset, error) {
	row, err := r.queries.GetNoteAssetByID(ctx, generated.GetNoteAssetByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteAsset{}, repoerrors.NewNotFound("note_asset.get_by_id", "note_asset")
		}
		return model.NoteAsset{}, repoerrors.WrapUnknown("note_asset.get_by_id", "note_asset", err)
	}
	return mapNoteAssetRow(row)
}

func (r *NoteAssetRepository) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
	rows, err := r.queries.ListNoteAssetsByNote(ctx, generated.ListNoteAssetsByNoteParams{
		NoteID:  string(noteID),
		Column2: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_asset.list_by_note", "note_asset", err)
	}
	out := make([]model.NoteAsset, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapNoteAssetRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note_asset.list_by_note", "note_asset", mapErr)
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *NoteAssetRepository) Update(ctx context.Context, params interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
	assetType, err := toDBAssetType(params.AssetType)
	if err != nil {
		return model.NoteAsset{}, repoerrors.WrapValidation("note_asset.update", "note_asset", err)
	}
	row, err := r.queries.UpdateNoteAsset(ctx, generated.UpdateNoteAssetParams{
		ID:          string(params.ID),
		AssetType:   assetType,
		StoragePath: params.StoragePath,
		MimeType:    params.MIMEType,
		UpdatedAt:   toPgTimestamptz(r.nowFn()),
		Version:     int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteAsset{}, repoerrors.NewConflict("note_asset.update", "note_asset")
		}
		return model.NoteAsset{}, repoerrors.WrapUnknown("note_asset.update", "note_asset", err)
	}
	return mapNoteAssetRow(row)
}

func (r *NoteAssetRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeleteNoteAsset(ctx, generated.DeleteNoteAssetParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("note_asset.delete", "note_asset", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("note_asset.delete", "note_asset")
	}
	return nil
}

func mapNoteAssetRow(row generated.NoteAsset) (model.NoteAsset, error) {
	assetType, err := fromDBAssetType(row.AssetType)
	if err != nil {
		return model.NoteAsset{}, err
	}
	return model.NoteAsset{
		ID:          model.ULID(fmt.Sprint(row.ID)),
		NoteID:      model.ULID(fmt.Sprint(row.NoteID)),
		AssetType:   assetType,
		StoragePath: row.StoragePath,
		MIMEType:    row.MimeType,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}

var _ constants.AssetType = constants.Image
