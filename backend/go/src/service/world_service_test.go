package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	shared "github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

type fakeWorldRepo struct {
	createFn func(context.Context, model.World) (model.World, error)
	getFn    func(context.Context, model.ULID, bool) (model.World, error)
	listFn   func(context.Context, interfaces.ListWorldsParams) ([]model.World, error)
	updateFn func(context.Context, interfaces.UpdateWorldParams) (model.World, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeWorldRepo) Create(ctx context.Context, world model.World) (model.World, error) {
	return f.createFn(ctx, world)
}
func (f *fakeWorldRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeWorldRepo) List(ctx context.Context, params interfaces.ListWorldsParams) ([]model.World, error) {
	return f.listFn(ctx, params)
}
func (f *fakeWorldRepo) Update(ctx context.Context, params interfaces.UpdateWorldParams) (model.World, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeWorldRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeWorldPolicy struct {
	validateFn func(name, description string, status constants.WorldStatus) (string, string, error)
}

func (f fakeWorldPolicy) NormalizeAndValidate(name, description string, status constants.WorldStatus) (string, string, error) {
	return f.validateFn(name, description, status)
}

func TestNewWorldServiceWithDepsFailsOnNilDependencies(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeWorldRepo{
		createFn: func(_ context.Context, world model.World) (model.World, error) { return world, nil },
	}

	tests := []struct {
		name string
		deps WorldServiceDeps
	}{
		{
			name: "nil repo",
			deps: WorldServiceDeps{
				Clock:       fakeClock{now: now},
				Policy:      DefaultWorldPolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil clock",
			deps: WorldServiceDeps{
				Repo:        repo,
				Policy:      DefaultWorldPolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil policy",
			deps: WorldServiceDeps{
				Repo:        repo,
				Clock:       fakeClock{now: now},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil id generator",
			deps: WorldServiceDeps{
				Repo:   repo,
				Clock:  fakeClock{now: now},
				Policy: DefaultWorldPolicy{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = NewWorldServiceWithDeps(tc.deps)
		})
	}
}

func TestWorldServiceCreateUsesPolicyAndGeneratedID(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var got model.World

	repo := &fakeWorldRepo{
		createFn: func(_ context.Context, world model.World) (model.World, error) {
			got = world
			return world, nil
		},
	}
	svc := NewWorldServiceWithDeps(WorldServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: now},
		Policy: fakeWorldPolicy{
			validateFn: func(name, description string, status constants.WorldStatus) (string, string, error) {
				return "Normalized Name", "desc", nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})

	out, err := svc.Create(ctx, CreateWorldParams{
		Name:        "ignored",
		Description: "ignored",
		Status:      constants.Draft,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if out.ID != "01HZZZZZZZZZZZZZZZZZZZZZZZ" {
		t.Fatalf("unexpected generated id: %s", out.ID)
	}
	if got.Name != "Normalized Name" {
		t.Fatalf("expected normalized name, got %q", got.Name)
	}
	if !got.AuditFields.CreatedAt.Equal(now) {
		t.Fatalf("expected injected clock timestamp")
	}
}

func TestWorldServiceCreateValidationAndIdGeneratorErrors(t *testing.T) {
	ctx := context.Background()
	repo := &fakeWorldRepo{
		createFn: func(_ context.Context, world model.World) (model.World, error) { return world, nil },
	}

	svcValidation := NewWorldServiceWithDeps(WorldServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: time.Now().UTC()},
		Policy: fakeWorldPolicy{
			validateFn: func(name, description string, status constants.WorldStatus) (string, string, error) {
				return "", "", errors.New("invalid world")
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})
	_, err := svcValidation.Create(ctx, CreateWorldParams{Name: "x", Status: constants.Draft})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	svcIDErr := NewWorldServiceWithDeps(WorldServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: time.Now().UTC()},
		Policy: fakeWorldPolicy{
			validateFn: func(name, description string, status constants.WorldStatus) (string, string, error) {
				return name, description, nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "", errors.New("idgen down") },
		},
	})
	_, err = svcIDErr.Create(ctx, CreateWorldParams{Name: "x", Status: constants.Draft})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown on idgen failure, got %v", err)
	}
}

func TestWorldServiceGetListUpdateDeleteValidationAndMappings(t *testing.T) {
	ctx := context.Background()
	repo := &fakeWorldRepo{
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.World, error) {
			return model.World{}, repoerrors.NewNotFound("world.get_by_id", "world")
		},
		listFn: func(_ context.Context, _ interfaces.ListWorldsParams) ([]model.World, error) {
			return nil, repoerrors.NewConflict("world.list", "world")
		},
		updateFn: func(_ context.Context, _ interfaces.UpdateWorldParams) (model.World, error) {
			return model.World{}, repoerrors.NewConflict("world.update", "world")
		},
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("world.delete", "world", errors.New("db"))
		},
	}
	svc := NewWorldServiceWithDeps(WorldServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: time.Now().UTC()},
		Policy: fakeWorldPolicy{
			validateFn: func(name, description string, status constants.WorldStatus) (string, string, error) {
				if name == "bad" {
					return "", "", errors.New("invalid")
				}
				return name, description, nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})

	_, err := svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id")
	}
	_, err = svc.GetByID(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping")
	}

	_, err = svc.List(ctx, ListWorldsParams{Offset: -1, Limit: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for offset")
	}
	_, err = svc.List(ctx, ListWorldsParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for limit")
	}
	_, err = svc.List(ctx, ListWorldsParams{Offset: 0, Limit: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping on list")
	}

	_, err = svc.Update(ctx, UpdateWorldParams{ID: "", Name: "ok", Status: constants.Draft, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for missing id")
	}
	_, err = svc.Update(ctx, UpdateWorldParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "ok", Status: constants.Draft, ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for expected version")
	}
	_, err = svc.Update(ctx, UpdateWorldParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "bad", Status: constants.Draft, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation from policy")
	}
	_, err = svc.Update(ctx, UpdateWorldParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "ok", Status: constants.Draft, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping from repo")
	}

	err = svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for delete id")
	}
	err = svc.Delete(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping for delete")
	}
}

func TestDefaultWorldPolicyAndHelpers(t *testing.T) {
	policy := DefaultWorldPolicy{}
	name, description, err := policy.NormalizeAndValidate("  My   World  ", "  Desc  ", constants.Active)
	if err != nil {
		t.Fatalf("unexpected validation failure: %v", err)
	}
	if name != "My World" || description != "Desc" {
		t.Fatalf("unexpected normalization: %q %q", name, description)
	}

	_, _, err = policy.NormalizeAndValidate("   ", "d", constants.Active)
	if err == nil {
		t.Fatalf("expected error for empty name")
	}
	_, _, err = policy.NormalizeAndValidate("world", "d", constants.WorldStatus(99))
	if err == nil {
		t.Fatalf("expected error for invalid status")
	}

	if !isValidWorldStatus(constants.Draft) || !isValidWorldStatus(constants.Active) || !isValidWorldStatus(constants.Archived) {
		t.Fatalf("expected valid known statuses")
	}
	if isValidWorldStatus(constants.WorldStatus(255)) {
		t.Fatalf("expected invalid status to fail")
	}
}

func TestMapWorldRepositoryErrorDefaultPathAndListItems(t *testing.T) {
	err := shared.MapRepositoryError(errors.New("plain"), "world_service.test", worldServiceName)
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected default unknown mapping")
	}

	deletedAt := time.Now().UTC()
	items := toWorldListItems([]model.World{
		{
			ID:          "01HZZZZZZZZZZZZZZZZZZZZZZZ",
			Name:        "A",
			Description: "B",
			Status:      constants.Draft,
			AuditFields: model.AuditFields{
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				DeletedAt: &deletedAt,
				Version:   2,
			},
		},
	})
	if len(items) != 1 || items[0].Name != "A" || items[0].DeletedAt == nil {
		t.Fatalf("unexpected world list item mapping: %+v", items)
	}
}
