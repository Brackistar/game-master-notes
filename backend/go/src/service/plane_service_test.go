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

type fakePlaneRepo struct {
	createFn func(context.Context, model.Plane) (model.Plane, error)
	getFn    func(context.Context, model.ULID, bool) (model.Plane, error)
	listFn   func(context.Context, interfaces.ListPlanesParams) ([]model.Plane, error)
	updateFn func(context.Context, interfaces.UpdatePlaneParams) (model.Plane, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakePlaneRepo) Create(ctx context.Context, plane model.Plane) (model.Plane, error) {
	return f.createFn(ctx, plane)
}
func (f *fakePlaneRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakePlaneRepo) List(ctx context.Context, params interfaces.ListPlanesParams) ([]model.Plane, error) {
	return f.listFn(ctx, params)
}
func (f *fakePlaneRepo) Update(ctx context.Context, params interfaces.UpdatePlaneParams) (model.Plane, error) {
	return f.updateFn(ctx, params)
}
func (f *fakePlaneRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakePlanePolicy struct {
	validateFn func(name, description string) (string, string, error)
}

func (f fakePlanePolicy) NormalizeAndValidate(name, description string) (string, string, error) {
	return f.validateFn(name, description)
}

func TestPlaneServiceNilDependenciesPanic(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakePlaneRepo{createFn: func(_ context.Context, p model.Plane) (model.Plane, error) { return p, nil }}
	tests := []PlaneServiceDeps{
		{Clock: fakeClock{now: now}, Policy: DefaultPlanePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Policy: DefaultPlanePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, Policy: DefaultPlanePolicy{}},
	}
	for _, deps := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = NewPlaneServiceWithDeps(deps)
		}()
	}
}

func TestPlaneServiceCreateAndMappings(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var created model.Plane
	repo := &fakePlaneRepo{
		createFn: func(_ context.Context, plane model.Plane) (model.Plane, error) { created = plane; return plane, nil },
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Plane, error) {
			return model.Plane{}, repoerrors.NewNotFound("plane.get_by_id", "plane")
		},
		listFn: func(_ context.Context, _ interfaces.ListPlanesParams) ([]model.Plane, error) {
			return nil, repoerrors.NewConflict("plane.list", "plane")
		},
		updateFn: func(_ context.Context, _ interfaces.UpdatePlaneParams) (model.Plane, error) {
			return model.Plane{}, repoerrors.NewConflict("plane.update", "plane")
		},
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("plane.delete", "plane", errors.New("db"))
		},
	}
	svc := NewPlaneServiceWithDeps(PlaneServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: now},
		Policy: fakePlanePolicy{
			validateFn: func(name, description string) (string, string, error) {
				if name == "bad" {
					return "", "", errors.New("invalid")
				}
				return "Name", "Desc", nil
			},
		},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
	})

	plane, err := svc.Create(ctx, CreatePlaneParams{Name: "ok"})
	if err != nil || plane.ID == "" {
		t.Fatalf("expected create success, err=%v", err)
	}
	if created.Name != "Name" || !created.AuditFields.CreatedAt.Equal(now) {
		t.Fatalf("expected normalized create payload")
	}

	_, err = svc.Create(ctx, CreatePlaneParams{Name: "bad"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation from policy")
	}
	svcIDErr := NewPlaneServiceWithDeps(PlaneServiceDeps{
		Repo:        repo,
		Clock:       fakeClock{now: now},
		Policy:      DefaultPlanePolicy{},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "", errors.New("idgen") }},
	})
	_, err = svcIDErr.Create(ctx, CreatePlaneParams{Name: "name"})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown on idgen failure")
	}

	_, err = svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.GetByID(ctx, "01H", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping")
	}

	_, err = svc.List(ctx, ListPlanesParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected offset validation")
	}
	_, err = svc.List(ctx, ListPlanesParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected limit validation")
	}
	_, err = svc.List(ctx, ListPlanesParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping")
	}

	_, err = svc.Update(ctx, UpdatePlaneParams{ID: "", Name: "x", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.Update(ctx, UpdatePlaneParams{ID: "01H", Name: "x", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected expected_version validation")
	}
	_, err = svc.Update(ctx, UpdatePlaneParams{ID: "01H", Name: "bad", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected policy validation")
	}
	_, err = svc.Update(ctx, UpdatePlaneParams{ID: "01H", Name: "good", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping")
	}

	err = svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected delete id validation")
	}
	err = svc.Delete(ctx, "01H")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping")
	}
}
