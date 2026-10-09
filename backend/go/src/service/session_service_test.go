package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakeSessionRepo struct {
	createFn func(context.Context, model.Session) (model.Session, error)
	getFn    func(context.Context, model.ULID, bool) (model.Session, error)
	listFn   func(context.Context, interfaces.ListSessionsParams) ([]model.Session, error)
	updateFn func(context.Context, interfaces.UpdateSessionParams) (model.Session, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeSessionRepo) Create(ctx context.Context, session model.Session) (model.Session, error) {
	return f.createFn(ctx, session)
}
func (f *fakeSessionRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeSessionRepo) List(ctx context.Context, params interfaces.ListSessionsParams) ([]model.Session, error) {
	return f.listFn(ctx, params)
}
func (f *fakeSessionRepo) Update(ctx context.Context, params interfaces.UpdateSessionParams) (model.Session, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeSessionRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeSessionPolicy struct {
	validateFn func(summary string) (string, error)
}

func (f fakeSessionPolicy) NormalizeAndValidate(summary string) (string, error) {
	return f.validateFn(summary)
}

func TestSessionServiceCreateValidationAndMappings(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakeSessionRepo{
		createFn: func(_ context.Context, session model.Session) (model.Session, error) { return session, nil },
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Session, error) {
			return model.Session{}, repoerrors.NewNotFound("session.get_by_id", "session")
		},
		listFn: func(_ context.Context, _ interfaces.ListSessionsParams) ([]model.Session, error) {
			return nil, repoerrors.NewConflict("session.list", "session")
		},
		updateFn: func(_ context.Context, _ interfaces.UpdateSessionParams) (model.Session, error) {
			return model.Session{}, repoerrors.NewConflict("session.update", "session")
		},
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("session.delete", "session", errors.New("db"))
		},
	}
	svc := NewSessionServiceWithDeps(SessionServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: now},
		Policy: fakeSessionPolicy{
			validateFn: func(summary string) (string, error) {
				if summary == "bad" {
					return "", errors.New("invalid")
				}
				return "summary", nil
			},
		},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01H", nil }},
	})

	_, err := svc.Create(ctx, CreateSessionParams{CampaignID: "", SummaryMD: "ok"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected campaign id validation")
	}
	_, err = svc.Create(ctx, CreateSessionParams{CampaignID: "01C", SummaryMD: "bad"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected policy validation")
	}

	created, err := svc.Create(ctx, CreateSessionParams{CampaignID: "01C", SummaryMD: "ok"})
	if err != nil || created.ID == "" || !created.AuditFields.CreatedAt.Equal(now) {
		t.Fatalf("expected create success")
	}
	svcIDErr := NewSessionServiceWithDeps(SessionServiceDeps{
		Repo:        repo,
		Clock:       fakeClock{now: now},
		Policy:      DefaultSessionPolicy{},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "", errors.New("idgen") }},
	})
	_, err = svcIDErr.Create(ctx, CreateSessionParams{CampaignID: "01C"})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown idgen error")
	}

	_, err = svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.GetByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping")
	}
	_, err = svc.List(ctx, ListSessionsParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected offset validation")
	}
	_, err = svc.List(ctx, ListSessionsParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected limit validation")
	}
	_, err = svc.List(ctx, ListSessionsParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping")
	}
	_, err = svc.Update(ctx, UpdateSessionParams{ID: "", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.Update(ctx, UpdateSessionParams{ID: "01", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected expected_version validation")
	}
	_, err = svc.Update(ctx, UpdateSessionParams{ID: "01", ExpectedVersion: 1, SummaryMD: "bad"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected policy validation")
	}
	_, err = svc.Update(ctx, UpdateSessionParams{ID: "01", ExpectedVersion: 1, SummaryMD: "ok"})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping")
	}
	err = svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	err = svc.Delete(ctx, "01")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping")
	}
}

func TestSessionServiceNilDependenciesPanicAndHelpers(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeSessionRepo{createFn: func(_ context.Context, s model.Session) (model.Session, error) { return s, nil }}
	deps := []SessionServiceDeps{
		{Clock: fakeClock{now: now}, Policy: DefaultSessionPolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Policy: DefaultSessionPolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, Policy: DefaultSessionPolicy{}},
	}
	for _, dep := range deps {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = NewSessionServiceWithDeps(dep)
		}()
	}

	items := toSessionListItems([]model.Session{{ID: "1", CampaignID: "2"}})
	if len(items) != 1 || items[0].CampaignID != "2" {
		t.Fatalf("unexpected mapping")
	}
}
