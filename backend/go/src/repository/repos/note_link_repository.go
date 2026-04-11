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

var _ interfaces.NoteLinkRepository = (*NoteLinkRepository)(nil)

type NoteLinkRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewNoteLinkRepository(db generated.DBTX) *NoteLinkRepository {
	return &NoteLinkRepository{
		queries: generated.New(db),
		nowFn:   func() time.Time { return time.Now().UTC() },
	}
}

func (r *NoteLinkRepository) Create(ctx context.Context, link model.NoteLink) (model.NoteLink, error) {
	defer helpers.LogRepositoryCall()()
	linkType, err := toDBNoteLinkType(link.LinkType)
	if err != nil {
		return model.NoteLink{}, repoerrors.WrapValidation("note_link.create", "note_link", err)
	}
	row, err := r.queries.CreateNoteLink(ctx, generated.CreateNoteLinkParams{
		PID:           string(link.ID),
		PSourceNoteID: string(link.SourceNoteID),
		PTargetNoteID: string(link.TargetNoteID),
		PLinkType:     linkType,
	})
	if err != nil {
		return model.NoteLink{}, mapFunctionError(err, "note_link.create", "note_link",
			map[string]struct{}{
				repoerrors.GMNSourceNoteNotFound: {},
				repoerrors.GMNTargetNoteNotFound: {},
			},
			map[string]struct{}{
				repoerrors.GMNNoteLinkAlreadyOpen: {},
			},
		)
	}
	return mapNoteLinkRow(row)
}

func (r *NoteLinkRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteLink, error) {
	defer helpers.LogRepositoryCall()()
	row, err := r.queries.GetNoteLinkByID(ctx, generated.GetNoteLinkByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteLink{}, repoerrors.NewNotFound("note_link.get_by_id", "note_link")
		}
		return model.NoteLink{}, repoerrors.WrapUnknown("note_link.get_by_id", "note_link", err)
	}
	return mapNoteLinkRow(row)
}

func (r *NoteLinkRepository) ListBySource(ctx context.Context, sourceNoteID model.ULID, params interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
	defer helpers.LogRepositoryCall()()
	rows, err := r.queries.ListNoteLinksBySource(ctx, generated.ListNoteLinksBySourceParams{
		SourceNoteID: string(sourceNoteID),
		Column2:      params.IncludeDeleted,
		Offset:       params.Offset,
		Limit:        params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_link.list_by_source", "note_link", err)
	}
	out := make([]model.NoteLink, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapNoteLinkRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note_link.list_by_source", "note_link", mapErr)
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *NoteLinkRepository) ListByTarget(ctx context.Context, targetNoteID model.ULID, params interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
	defer helpers.LogRepositoryCall()()
	rows, err := r.queries.ListNoteLinksByTarget(ctx, generated.ListNoteLinksByTargetParams{
		TargetNoteID: string(targetNoteID),
		Column2:      params.IncludeDeleted,
		Offset:       params.Offset,
		Limit:        params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("note_link.list_by_target", "note_link", err)
	}
	out := make([]model.NoteLink, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapNoteLinkRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("note_link.list_by_target", "note_link", mapErr)
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *NoteLinkRepository) Update(ctx context.Context, params interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
	defer helpers.LogRepositoryCall()()
	linkType, err := toDBNoteLinkType(params.LinkType)
	if err != nil {
		return model.NoteLink{}, repoerrors.WrapValidation("note_link.update", "note_link", err)
	}
	row, err := r.queries.UpdateNoteLink(ctx, generated.UpdateNoteLinkParams{
		ID:        string(params.ID),
		LinkType:  linkType,
		UpdatedAt: toPgTimestamptz(r.nowFn()),
		Version:   int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NoteLink{}, repoerrors.NewConflict("note_link.update", "note_link")
		}
		return model.NoteLink{}, repoerrors.WrapUnknown("note_link.update", "note_link", err)
	}
	return mapNoteLinkRow(row)
}

func (r *NoteLinkRepository) Delete(ctx context.Context, id model.ULID) error {
	defer helpers.LogRepositoryCall()()
	current, err := r.GetByID(ctx, id, false)
	if err != nil {
		return err
	}

	dbLinkType, err := toDBNoteLinkType(current.LinkType)
	if err != nil {
		return repoerrors.WrapValidation("note_link.delete", "note_link", err)
	}

	_, err = r.queries.DeleteNoteLink(ctx, generated.DeleteNoteLinkParams{
		PSourceNoteID: string(current.SourceNoteID),
		PTargetNoteID: string(current.TargetNoteID),
		PLinkType:     dbLinkType,
	})
	if err != nil {
		return mapFunctionError(err, "note_link.delete", "note_link",
			map[string]struct{}{
				repoerrors.GMNNoteLinkNotActive: {},
			},
			nil,
		)
	}
	return nil
}

func mapNoteLinkRow(row generated.NoteLink) (model.NoteLink, error) {
	linkType, err := fromDBNoteLinkType(row.LinkType)
	if err != nil {
		return model.NoteLink{}, err
	}
	return model.NoteLink{
		ID:           model.ULID(fmt.Sprint(row.ID)),
		SourceNoteID: model.ULID(fmt.Sprint(row.SourceNoteID)),
		TargetNoteID: model.ULID(fmt.Sprint(row.TargetNoteID)),
		LinkType:     linkType,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}

var _ constants.NoteLinkType = constants.Related
