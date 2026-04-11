package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	"github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const sessionServiceName string = "session"

type SessionPolicy interface {
	NormalizeAndValidate(summaryMD string) (string, error)
}

type DefaultSessionPolicy struct{}

func (DefaultSessionPolicy) NormalizeAndValidate(summaryMD string) (string, error) {
	return strings.TrimSpace(summaryMD), nil
}

type CreateSessionParams struct {
	CampaignID model.ULID
	PlayedOn   *time.Time
	SummaryMD  string
}

type UpdateSessionParams struct {
	ID              model.ULID
	PlayedOn        *time.Time
	SummaryMD       string
	ExpectedVersion model.Version
}

type ListSessionsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type SessionListItem struct {
	ID         model.ULID
	CampaignID model.ULID
	PlayedOn   *time.Time
	SummaryMD  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Version    model.Version
}

type SessionService struct {
	repo        repo.SessionRepository
	clock       shared.Clock
	policy      SessionPolicy
	idGenerator shared.IDGenerator
}

type SessionServiceDeps struct {
	Repo        repo.SessionRepository
	Clock       shared.Clock
	Policy      SessionPolicy
	IDGenerator shared.IDGenerator
}

func NewSessionService(repo repo.SessionRepository, idGenerator shared.IDGenerator) *SessionService {
	return NewSessionServiceWithDeps(SessionServiceDeps{
		Repo:        repo,
		Clock:       shared.SystemClock{},
		Policy:      DefaultSessionPolicy{},
		IDGenerator: idGenerator,
	})
}

func NewSessionServiceWithDeps(deps SessionServiceDeps) *SessionService {
	shared.PanicIfNilDependency(sessionServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(sessionServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(sessionServiceName, "Policy", deps.Policy)
	shared.PanicIfNilDependency(sessionServiceName, "IDGenerator", deps.IDGenerator)
	return &SessionService{
		repo:        deps.Repo,
		clock:       deps.Clock,
		policy:      deps.Policy,
		idGenerator: deps.IDGenerator,
	}
}

func (s *SessionService) Create(ctx context.Context, params CreateSessionParams) (model.Session, error) {
	op := "session_service.create"
	if strings.TrimSpace(string(params.CampaignID)) == "" {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "campaign_id"))
	}
	summary, err := s.policy.NormalizeAndValidate(params.SummaryMD)
	if err != nil {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, err)
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Session{}, serviceerrors.WrapUnknown(op, sessionServiceName, err)
	}
	now := s.clock.Now()
	session, repoErr := s.repo.Create(ctx, model.Session{
		ID:         id,
		CampaignID: params.CampaignID,
		PlayedOn:   params.PlayedOn,
		SummaryMD:  summary,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Session{}, shared.MapRepositoryError(repoErr, op, sessionServiceName)
	}
	return session, nil
}

func (s *SessionService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error) {
	op := "session_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	session, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Session{}, shared.MapRepositoryError(err, op, sessionServiceName)
	}
	return session, nil
}

func (s *SessionService) List(ctx context.Context, params ListSessionsParams) ([]SessionListItem, error) {
	op := "session_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, sessionServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, sessionServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	rows, err := s.repo.List(ctx, repo.ListSessionsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, sessionServiceName)
	}
	return toSessionListItems(rows), nil
}

func (s *SessionService) Update(ctx context.Context, params UpdateSessionParams) (model.Session, error) {
	op := "session_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	summary, err := s.policy.NormalizeAndValidate(params.SummaryMD)
	if err != nil {
		return model.Session{}, serviceerrors.WrapValidation(op, sessionServiceName, err)
	}
	session, repoErr := s.repo.Update(ctx, repo.UpdateSessionParams{
		ID:              params.ID,
		PlayedOn:        params.PlayedOn,
		SummaryMD:       summary,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Session{}, shared.MapRepositoryError(repoErr, op, sessionServiceName)
	}
	return session, nil
}

func (s *SessionService) Delete(ctx context.Context, id model.ULID) error {
	op := "session_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, sessionServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, sessionServiceName)
	}
	return nil
}

func toSessionListItems(rows []model.Session) []SessionListItem {
	out := make([]SessionListItem, 0, len(rows))
	for _, session := range rows {
		out = append(out, SessionListItem{
			ID:         session.ID,
			CampaignID: session.CampaignID,
			PlayedOn:   session.PlayedOn,
			SummaryMD:  session.SummaryMD,
			CreatedAt:  session.AuditFields.CreatedAt,
			UpdatedAt:  session.AuditFields.UpdatedAt,
			DeletedAt:  session.AuditFields.DeletedAt,
			Version:    session.AuditFields.Version,
		})
	}
	return out
}
