package interfaces

import (
	"context"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
)

type ListNoteAssetsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type UpdateNoteAssetParams struct {
	ID              model.ULID
	AssetType       constants.AssetType
	StoragePath     string
	MIMEType        string
	ExpectedVersion model.Version
}

type NoteAssetRepository interface {
	Create(ctx context.Context, asset model.NoteAsset) (model.NoteAsset, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteAsset, error)
	ListByNote(ctx context.Context, noteID model.ULID, params ListNoteAssetsParams) ([]model.NoteAsset, error)
	Update(ctx context.Context, params UpdateNoteAssetParams) (model.NoteAsset, error)
	Delete(ctx context.Context, id model.ULID) error
}
