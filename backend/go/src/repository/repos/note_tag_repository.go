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

var _ interfaces.NoteTagRepository = (*NoteTagRepository)(nil)

type NoteTagRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewNoteTagRepository(db generated.DBTX) *NoteTagRepository {
	return &NoteTagRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *NoteTagRepository) Create(ctx context.Context, rel model.NoteTag) (model.NoteTag, error) {
	row, err := r.queries.CreateNoteTag(ctx, generated.CreateNoteTagParams{
		NoteID:    string(rel.NoteID),
		TagID:     string(rel.TagID),
		CreatedAt: toPgTimestamptz(rel.CreatedAt),
		UpdatedAt: toPgTimestamptz(rel.UpdatedAt),
		DeletedAt: toNullablePgTimestamptz(rel.DeletedAt),
	})
	if err != nil {
		return model.NoteTag{}, repoerrors.WrapUnknown("note_tag.create", "note_tag", err)
	}
	return mapNoteTagRow(row), nil
}

func (r *NoteTagRepository) Get(ctx context.Context, noteID, tagID model.ULID, includeDeleted bool) (model.NoteTag, error) {
	row, err := r.queries.GetNoteTag(ctx, generated.GetNoteTagParams{
		NoteID:  string(noteID),
		TagID:   string(tagID),
		Column3: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteTag{}, repoerrors.NewNotFound("note_tag.get", "note_tag")
		}
		return model.NoteTag{}, repoerrors.WrapUnknown("note_tag.get", "note_tag", err)
	}
	return mapNoteTagRow(row), nil
}

func (r *NoteTagRepository) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
	rows, err := r.queries.ListNoteTagsByNote(ctx, generated.ListNoteTagsByNoteParams{
		NoteID:  string(noteID),
		Column2: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_tag.list_by_note", "note_tag", err)
	}
	out := make([]model.NoteTag, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapNoteTagRow(row))
	}
	return out, nil
}

func (r *NoteTagRepository) ListByTag(ctx context.Context, tagID model.ULID, params interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
	rows, err := r.queries.ListNoteTagsByTag(ctx, generated.ListNoteTagsByTagParams{
		TagID:   string(tagID),
		Column2: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_tag.list_by_tag", "note_tag", err)
	}
	out := make([]model.NoteTag, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapNoteTagRow(row))
	}
	return out, nil
}

func (r *NoteTagRepository) Delete(ctx context.Context, noteID, tagID model.ULID) error {
	affected, err := r.queries.DeleteNoteTag(ctx, generated.DeleteNoteTagParams{
		NoteID:    string(noteID),
		TagID:     string(tagID),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("note_tag.delete", "note_tag", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("note_tag.delete", "note_tag")
	}
	return nil
}

func mapNoteTagRow(row generated.NoteTag) model.NoteTag {
	return model.NoteTag{
		NoteID:    model.ULID(fmt.Sprint(row.NoteID)),
		TagID:     model.ULID(fmt.Sprint(row.TagID)),
		CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
		UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
		DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
	}
}
