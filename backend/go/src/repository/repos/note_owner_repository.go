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

var _ interfaces.NoteOwnerRepository = (*NoteOwnerRepository)(nil)

type NoteOwnerRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewNoteOwnerRepository(db generated.DBTX) *NoteOwnerRepository {
	return &NoteOwnerRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *NoteOwnerRepository) Create(ctx context.Context, rel model.NoteOwner) (model.NoteOwner, error) {
	ownerType, err := toDBOwnerType(rel.OwnerType)
	if err != nil {
		return model.NoteOwner{}, repoerrors.WrapValidation("note_owner.create", "note_owner", err)
	}
	row, err := r.queries.CreateNoteOwner(ctx, generated.CreateNoteOwnerParams{
		NoteID:    string(rel.NoteID),
		OwnerType: ownerType,
		OwnerID:   string(rel.OwnerID),
		CreatedAt: toPgTimestamptz(rel.CreatedAt),
		UpdatedAt: toPgTimestamptz(rel.UpdatedAt),
		DeletedAt: toNullablePgTimestamptz(rel.DeletedAt),
	})
	if err != nil {
		return model.NoteOwner{}, repoerrors.WrapUnknown("note_owner.create", "note_owner", err)
	}
	return mapNoteOwnerRow(row)
}

func (r *NoteOwnerRepository) Get(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, includeDeleted bool) (model.NoteOwner, error) {
	dbOwnerType, err := toDBOwnerType(ownerType)
	if err != nil {
		return model.NoteOwner{}, repoerrors.WrapValidation("note_owner.get", "note_owner", err)
	}
	row, err := r.queries.GetNoteOwner(ctx, generated.GetNoteOwnerParams{
		NoteID:    string(noteID),
		OwnerType: dbOwnerType,
		OwnerID:   string(ownerID),
		Column4:   includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteOwner{}, repoerrors.NewNotFound("note_owner.get", "note_owner")
		}
		return model.NoteOwner{}, repoerrors.WrapUnknown("note_owner.get", "note_owner", err)
	}
	return mapNoteOwnerRow(row)
}

func (r *NoteOwnerRepository) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
	rows, err := r.queries.ListNoteOwnersByNote(ctx, generated.ListNoteOwnersByNoteParams{
		NoteID:  string(noteID),
		Column2: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_owner.list_by_note", "note_owner", err)
	}
	out := make([]model.NoteOwner, 0, len(rows))
	for _, row := range rows {
		rel, mapErr := mapNoteOwnerRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note_owner.list_by_note", "note_owner", mapErr)
		}
		out = append(out, rel)
	}
	return out, nil
}

func (r *NoteOwnerRepository) ListByOwner(ctx context.Context, ownerType constants.OwnerType, ownerID model.ULID, params interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
	dbOwnerType, err := toDBOwnerType(ownerType)
	if err != nil {
		return nil, repoerrors.WrapValidation("note_owner.list_by_owner", "note_owner", err)
	}
	rows, err := r.queries.ListNoteOwnersByOwner(ctx, generated.ListNoteOwnersByOwnerParams{
		OwnerType: dbOwnerType,
		OwnerID:   string(ownerID),
		Column3:   params.IncludeDeleted,
		Offset:    params.Offset,
		Limit:     params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_owner.list_by_owner", "note_owner", err)
	}
	out := make([]model.NoteOwner, 0, len(rows))
	for _, row := range rows {
		rel, mapErr := mapNoteOwnerRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note_owner.list_by_owner", "note_owner", mapErr)
		}
		out = append(out, rel)
	}
	return out, nil
}

func (r *NoteOwnerRepository) Delete(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID) error {
	dbOwnerType, err := toDBOwnerType(ownerType)
	if err != nil {
		return repoerrors.WrapValidation("note_owner.delete", "note_owner", err)
	}
	affected, err := r.queries.DeleteNoteOwner(ctx, generated.DeleteNoteOwnerParams{
		NoteID:    string(noteID),
		OwnerType: dbOwnerType,
		OwnerID:   string(ownerID),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("note_owner.delete", "note_owner", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("note_owner.delete", "note_owner")
	}
	return nil
}

func mapNoteOwnerRow(row generated.NoteOwner) (model.NoteOwner, error) {
	ownerType, err := fromDBOwnerType(row.OwnerType)
	if err != nil {
		return model.NoteOwner{}, err
	}
	return model.NoteOwner{
		NoteID:    model.ULID(fmt.Sprint(row.NoteID)),
		OwnerType: ownerType,
		OwnerID:   model.ULID(fmt.Sprint(row.OwnerID)),
		CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
		UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
		DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
	}, nil
}
