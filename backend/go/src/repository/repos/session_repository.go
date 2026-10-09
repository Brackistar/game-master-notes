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

var _ interfaces.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewSessionRepository(db generated.DBTX) *SessionRepository {
	return &SessionRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session model.Session) (model.Session, error) {
	defer helpers.LogRepositoryCall()()
	row, err := r.queries.CreateSession(ctx, generated.CreateSessionParams{
		ID:         string(session.ID),
		CampaignID: string(session.CampaignID),
		PlayedOn:   toNullablePgDate(session.PlayedOn),
		SummaryMd:  session.SummaryMD,
		CreatedAt:  toPgTimestamptz(session.AuditFields.CreatedAt),
		UpdatedAt:  toPgTimestamptz(session.AuditFields.UpdatedAt),
		DeletedAt:  toNullablePgTimestamptz(session.AuditFields.DeletedAt),
		Version:    int32(session.AuditFields.Version),
	})
	if err != nil {
		return model.Session{}, repoerrors.WrapUnknown("session.create", "session", err)
	}
	return mapSessionRow(row), nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error) {
	defer helpers.LogRepositoryCall()()
	row, err := r.queries.GetSessionByID(ctx, generated.GetSessionByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Session{}, repoerrors.NewNotFound("session.get_by_id", "session")
		}
		return model.Session{}, repoerrors.WrapUnknown("session.get_by_id", "session", err)
	}
	return mapSessionRow(row), nil
}

func (r *SessionRepository) List(ctx context.Context, params interfaces.ListSessionsParams) ([]model.Session, error) {
	defer helpers.LogRepositoryCall()()
	rows, err := r.queries.ListSessions(ctx, generated.ListSessionsParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("session.list", "session", err)
	}
	out := make([]model.Session, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionRow(row))
	}
	return out, nil
}

func (r *SessionRepository) Update(ctx context.Context, params interfaces.UpdateSessionParams) (model.Session, error) {
	defer helpers.LogRepositoryCall()()
	row, err := r.queries.UpdateSession(ctx, generated.UpdateSessionParams{
		ID:        string(params.ID),
		PlayedOn:  toNullablePgDate(params.PlayedOn),
		SummaryMd: params.SummaryMD,
		UpdatedAt: toPgTimestamptz(r.nowFn()),
		Version:   int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Session{}, repoerrors.NewConflict("session.update", "session")
		}
		return model.Session{}, repoerrors.WrapUnknown("session.update", "session", err)
	}
	return mapSessionRow(row), nil
}

func (r *SessionRepository) Delete(ctx context.Context, id model.ULID) error {
	defer helpers.LogRepositoryCall()()
	affected, err := r.queries.DeleteSession(ctx, generated.DeleteSessionParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("session.delete", "session", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("session.delete", "session")
	}
	return nil
}

func mapSessionRow(row generated.Session) model.Session {
	return model.Session{
		ID:         model.ULID(fmt.Sprint(row.ID)),
		CampaignID: model.ULID(fmt.Sprint(row.CampaignID)),
		PlayedOn:   fromNullablePgDate(row.PlayedOn),
		SummaryMD:  row.SummaryMd,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}
}
