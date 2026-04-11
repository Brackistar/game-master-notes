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
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/repository/repos/shared"
	"github.com/jackc/pgx/v5"
)

var _ interfaces.NoteRepository = (*NoteRepository)(nil)

type NoteRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewNoteRepository(db generated.DBTX) *NoteRepository {
	return &NoteRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *NoteRepository) Create(ctx context.Context, note model.Note) (model.Note, error) {
	defer helpers.LogRepositoryCall()()
	noteType, err := toDBNoteType(note.NoteType)
	if err != nil {
		return model.Note{}, repoerrors.WrapValidation("note.create", "note", err)
	}

	row, err := r.queries.CreateNote(ctx, generated.CreateNoteParams{
		ID:           string(note.ID),
		Title:        note.Title,
		ContentMd:    note.ContentMD,
		NoteType:     noteType,
		MetadataJson: note.MetadataJSON,
		CreatedAt:    toPgTimestamptz(note.AuditFields.CreatedAt),
		UpdatedAt:    toPgTimestamptz(note.AuditFields.UpdatedAt),
		DeletedAt:    toNullablePgTimestamptz(note.AuditFields.DeletedAt),
		Version:      int32(note.AuditFields.Version),
	})
	if err != nil {
		return model.Note{}, repoerrors.WrapUnknown("note.create", "note", err)
	}

	out, err := mapNoteRow(row)
	if err != nil {
		return model.Note{}, repoerrors.WrapValidation("note.create", "note", err)
	}
	return out, nil
}

func (r *NoteRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Note, error) {
	defer helpers.LogRepositoryCall()()
	row, err := r.queries.GetNoteByID(ctx, generated.GetNoteByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Note{}, repoerrors.NewNotFound("note.get_by_id", "note")
		}
		return model.Note{}, repoerrors.WrapUnknown("note.get_by_id", "note", err)
	}

	out, err := mapNoteRow(row)
	if err != nil {
		return model.Note{}, repoerrors.WrapValidation("note.get_by_id", "note", err)
	}
	return out, nil
}

func (r *NoteRepository) List(ctx context.Context, params interfaces.ListNotesParams) ([]model.Note, error) {
	defer helpers.LogRepositoryCall()()
	rows, err := r.queries.ListNotes(ctx, generated.ListNotesParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note.list", "note", err)
	}

	out := make([]model.Note, 0, len(rows))
	for _, row := range rows {
		n, mapErr := mapNoteRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note.list", "note", mapErr)
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *NoteRepository) Update(ctx context.Context, params interfaces.UpdateNoteParams) (model.Note, error) {
	defer helpers.LogRepositoryCall()()
	noteType, err := toDBNoteType(params.NoteType)
	if err != nil {
		return model.Note{}, repoerrors.WrapValidation("note.update", "note", err)
	}

	row, err := r.queries.UpdateNote(ctx, generated.UpdateNoteParams{
		ID:           string(params.ID),
		Title:        params.Title,
		ContentMd:    params.ContentMD,
		NoteType:     noteType,
		MetadataJson: params.MetadataJSON,
		UpdatedAt:    toPgTimestamptz(r.nowFn()),
		Version:      int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Note{}, repoerrors.NewConflict("note.update", "note")
		}
		return model.Note{}, repoerrors.WrapUnknown("note.update", "note", err)
	}

	out, err := mapNoteRow(row)
	if err != nil {
		return model.Note{}, repoerrors.WrapValidation("note.update", "note", err)
	}
	return out, nil
}

func (r *NoteRepository) Delete(ctx context.Context, id model.ULID) error {
	defer helpers.LogRepositoryCall()()
	affected, err := r.queries.DeleteNote(ctx, generated.DeleteNoteParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("note.delete", "note", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("note.delete", "note")
	}
	return nil
}

func mapNoteRow(row generated.Note) (model.Note, error) {
	noteType, err := fromDBNoteType(row.NoteType)
	if err != nil {
		return model.Note{}, err
	}

	return model.Note{
		ID:           model.ULID(fmt.Sprint(row.ID)),
		Title:        row.Title,
		ContentMD:    row.ContentMd,
		NoteType:     noteType,
		MetadataJSON: row.MetadataJson,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}

func toDBNoteType(noteType constants.NoteType) (generated.NoteType, error) {
	switch noteType {
	case constants.General:
		return generated.NoteTypeGeneral, nil
	case constants.SummaryNote:
		return generated.NoteTypeSummaryNote, nil
	case constants.Map:
		return generated.NoteTypeMap, nil
	case constants.Character:
		return generated.NoteTypeCharacter, nil
	case constants.Location:
		return generated.NoteTypeLocation, nil
	default:
		return "", fmt.Errorf("invalid note type enum value: %d", noteType)
	}
}

func fromDBNoteType(noteType generated.NoteType) (constants.NoteType, error) {
	switch noteType {
	case generated.NoteTypeGeneral:
		return constants.General, nil
	case generated.NoteTypeSummaryNote:
		return constants.SummaryNote, nil
	case generated.NoteTypeMap:
		return constants.Map, nil
	case generated.NoteTypeCharacter:
		return constants.Character, nil
	case generated.NoteTypeLocation:
		return constants.Location, nil
	default:
		return constants.General, fmt.Errorf("invalid note type db value: %q", noteType)
	}
}
