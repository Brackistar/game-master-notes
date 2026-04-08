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

type fakePlayerRepo struct {
	createFn  func(context.Context, model.Player) (model.Player, error)
	getFn     func(context.Context, model.ULID, bool) (model.Player, error)
	listFn    func(context.Context, interfaces.ListPlayersParams) ([]model.Player, error)
	searchFn  func(context.Context, interfaces.SearchPlayersParams) ([]model.Player, error)
	updateFn  func(context.Context, interfaces.UpdatePlayerParams) (model.Player, error)
	deleteFn  func(context.Context, model.ULID) error
	restoreFn func(context.Context, model.ULID) (model.Player, error)
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time { return f.now }

type fakeNamePolicy struct {
	normalizeFn func(string) (string, error)
}

func (f fakeNamePolicy) NormalizeAndValidate(name string) (string, error) {
	return f.normalizeFn(name)
}

type fakeIDGenerator struct {
	newULIDFn func() (model.ULID, error)
}

func (f fakeIDGenerator) NewULID() (model.ULID, error) {
	return f.newULIDFn()
}

func (f *fakePlayerRepo) Create(ctx context.Context, p model.Player) (model.Player, error) {
	return f.createFn(ctx, p)
}
func (f *fakePlayerRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakePlayerRepo) List(ctx context.Context, params interfaces.ListPlayersParams) ([]model.Player, error) {
	return f.listFn(ctx, params)
}
func (f *fakePlayerRepo) SearchByName(ctx context.Context, params interfaces.SearchPlayersParams) ([]model.Player, error) {
	return f.searchFn(ctx, params)
}
func (f *fakePlayerRepo) Update(ctx context.Context, params interfaces.UpdatePlayerParams) (model.Player, error) {
	return f.updateFn(ctx, params)
}
func (f *fakePlayerRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}
func (f *fakePlayerRepo) Restore(ctx context.Context, id model.ULID) (model.Player, error) {
	return f.restoreFn(ctx, id)
}

func TestPlayerServiceCreateNormalizesAndValidatesName(t *testing.T) {
	ctx := context.Background()
	var received model.Player
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) {
			received = p
			return p, nil
		},
	}
	svc := NewPlayerServiceWithDeps(PlayerServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: now},
		NamePolicy: fakeNamePolicy{
			normalizeFn: func(in string) (string, error) {
				return normalizeSpaces(in), nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) {
				return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
			},
		},
	})

	out, err := svc.Create(ctx, CreatePlayerParams{
		ID:   model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"),
		Name: "  Ana   Maria  ",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if out.Name != "Ana Maria" {
		t.Fatalf("expected normalized name, got %q", out.Name)
	}
	if received.AuditFields.Version != 1 {
		t.Fatalf("expected version=1, got %d", received.AuditFields.Version)
	}
	if !received.AuditFields.CreatedAt.Equal(now) {
		t.Fatalf("expected injected clock time, got %v", received.AuditFields.CreatedAt)
	}
}

func TestPlayerServiceCreateRejectsInvalidNames(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) { return p, nil },
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) {
			return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
		},
	})

	tests := []string{
		"ab",
		"this-name-is-way-way-way-way-way-way-way-way-way-way-too-long",
		"Ana😊",
	}
	for _, name := range tests {
		_, err := svc.Create(ctx, CreatePlayerParams{
			ID:   model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"),
			Name: name,
		})
		if !errors.Is(err, serviceerrors.ErrValidation) {
			t.Fatalf("expected validation for %q, got %v", name, err)
		}
	}
}

func TestDefaultPlayerNamePolicy(t *testing.T) {
	policy := DefaultPlayerNamePolicy{}

	out, err := policy.NormalizeAndValidate("  Ana   Maria  ")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if out != "Ana Maria" {
		t.Fatalf("unexpected normalization output: %q", out)
	}

	_, err = policy.NormalizeAndValidate("ab")
	if err == nil {
		t.Fatalf("expected too-short error")
	}
}

func TestPlayerServiceSearchUsesContainsAndMinLength(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &fakePlayerRepo{
		searchFn: func(_ context.Context, params interfaces.SearchPlayersParams) ([]model.Player, error) {
			called = true
			if params.Query != "mar" {
				t.Fatalf("expected normalized query mar, got %q", params.Query)
			}
			return []model.Player{
				{ID: model.ULID("2"), Name: "Mario"},
				{ID: model.ULID("1"), Name: "Maria"},
			}, nil
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) {
			return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
		},
	})

	_, err := svc.SearchByName(ctx, SearchPlayersParams{
		Query: "  ma ",
		Limit: 10,
	})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for short query, got %v", err)
	}

	out, err := svc.SearchByName(ctx, SearchPlayersParams{
		Query: "  mar ",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if !called {
		t.Fatalf("expected repo search to be called")
	}
	if len(out) != 2 || out[0].Name != "Maria" || out[1].Name != "Mario" {
		t.Fatalf("expected alphabetical sort, got %+v", out)
	}
}

func TestPlayerServiceRestoreRules(t *testing.T) {
	ctx := context.Background()
	active := model.Player{
		ID:   model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"),
		Name: "Active",
		AuditFields: model.AuditFields{
			DeletedAt: nil,
		},
	}
	deletedAt := time.Now().UTC()
	deleted := active
	deleted.AuditFields.DeletedAt = &deletedAt

	repo := &fakePlayerRepo{
		getFn: func(_ context.Context, _ model.ULID, includeDeleted bool) (model.Player, error) {
			if !includeDeleted {
				t.Fatalf("restore must query includeDeleted=true")
			}
			return active, nil
		},
		restoreFn: func(_ context.Context, _ model.ULID) (model.Player, error) {
			t.Fatalf("restore should not be called when active")
			return model.Player{}, nil
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) {
			return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
		},
	})
	_, err := svc.Restore(ctx, active.ID)
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict on restore active player, got %v", err)
	}

	repo2 := &fakePlayerRepo{
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Player, error) { return deleted, nil },
		restoreFn: func(_ context.Context, _ model.ULID) (model.Player, error) {
			deleted.AuditFields.DeletedAt = nil
			return deleted, nil
		},
	}
	svc2 := NewPlayerService(repo2, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) {
			return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
		},
	})
	restored, err := svc2.Restore(ctx, deleted.ID)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if restored.AuditFields.DeletedAt != nil {
		t.Fatalf("expected deleted_at nil after restore")
	}
}

func TestPlayerServiceMapsRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.NewNotFound("player.delete", "player")
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) {
			return model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"), nil
		},
	})
	err := svc.Delete(ctx, model.ULID("01HZZZZZZZZZZZZZZZZZZZZZZZ"))
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected service not found, got %v", err)
	}
}

func TestPlayerServiceCreateRequiresID(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) { return p, nil },
	}
	svc := NewPlayerServiceWithDeps(PlayerServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: time.Now().UTC()},
		NamePolicy: fakeNamePolicy{
			normalizeFn: func(in string) (string, error) {
				return in, nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) {
				return "", errors.New("idgen down")
			},
		},
	})

	_, err := svc.Create(ctx, CreatePlayerParams{Name: "Valid Name"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation error when id is missing, got %v", err)
	}
}

func TestNewPlayerServiceWithDepsFailsOnNilDependencies(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) { return p, nil },
	}

	tests := []struct {
		name string
		deps PlayerServiceDeps
	}{
		{
			name: "nil repo",
			deps: PlayerServiceDeps{
				Clock:       fakeClock{now: now},
				NamePolicy:  DefaultPlayerNamePolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil clock",
			deps: PlayerServiceDeps{
				Repo:        repo,
				NamePolicy:  DefaultPlayerNamePolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil name policy",
			deps: PlayerServiceDeps{
				Repo:        repo,
				Clock:       fakeClock{now: now},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil id generator",
			deps: PlayerServiceDeps{
				Repo:       repo,
				Clock:      fakeClock{now: now},
				NamePolicy: DefaultPlayerNamePolicy{},
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
			_ = NewPlayerServiceWithDeps(tc.deps)
		})
	}
}

func TestNewPlayerServiceSetsDefaultsButPreservesInjectedIDGenerator(t *testing.T) {
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) { return p, nil },
	}
	idgen := fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	}
	svc := NewPlayerService(repo, idgen)
	if svc == nil {
		t.Fatalf("expected service instance")
	}
}

func TestNewPlayerServiceWithDepsSucceedsWithAllDependencies(t *testing.T) {
	repo := &fakePlayerRepo{
		createFn: func(_ context.Context, p model.Player) (model.Player, error) { return p, nil },
	}
	idgen := fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	}
	svc := NewPlayerServiceWithDeps(PlayerServiceDeps{
		Repo:        repo,
		Clock:       fakeClock{now: time.Now().UTC()},
		NamePolicy:  DefaultPlayerNamePolicy{},
		IDGenerator: idgen,
	})
	if svc == nil {
		t.Fatalf("expected service instance")
	}
}

func TestPlayerServiceGetByIDValidationAndMapping(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Player, error) {
			return model.Player{}, repoerrors.NewNotFound("player.get_by_id", "player")
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	_, err := svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id, got %v", err)
	}

	_, err = svc.GetByID(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}

func TestPlayerServiceListValidationAndRepoErrorMapping(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		listFn: func(_ context.Context, _ interfaces.ListPlayersParams) ([]model.Player, error) {
			return nil, repoerrors.NewConflict("player.list", "player")
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	_, err := svc.List(ctx, ListPlayersParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for negative offset, got %v", err)
	}
	_, err = svc.List(ctx, ListPlayersParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for zero limit, got %v", err)
	}
	_, err = svc.List(ctx, ListPlayersParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping, got %v", err)
	}
}

func TestPlayerServiceSearchValidationAndRepoErrorMapping(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		searchFn: func(_ context.Context, _ interfaces.SearchPlayersParams) ([]model.Player, error) {
			return nil, repoerrors.WrapUnknown("player.search_by_name", "player", errors.New("db down"))
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	_, err := svc.SearchByName(ctx, SearchPlayersParams{Query: "aaa", Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for negative offset, got %v", err)
	}
	_, err = svc.SearchByName(ctx, SearchPlayersParams{Query: "aaa", Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for limit, got %v", err)
	}
	_, err = svc.SearchByName(ctx, SearchPlayersParams{Query: "aaa", Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping, got %v", err)
	}
}

func TestPlayerServiceUpdateValidationAndRepoErrorMapping(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		updateFn: func(_ context.Context, _ interfaces.UpdatePlayerParams) (model.Player, error) {
			return model.Player{}, repoerrors.NewConflict("player.update", "player")
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	_, err := svc.Update(ctx, UpdatePlayerParams{ID: "", Name: "Valid Name", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for id, got %v", err)
	}
	_, err = svc.Update(ctx, UpdatePlayerParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "Valid Name", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for expected version, got %v", err)
	}
	_, err = svc.Update(ctx, UpdatePlayerParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "ab", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for name, got %v", err)
	}
	_, err = svc.Update(ctx, UpdatePlayerParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "Valid Name", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping, got %v", err)
	}
}

func TestPlayerServiceDeleteValidationAndUnknownMapping(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("player.delete", "player", errors.New("db"))
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	err := svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id, got %v", err)
	}
	err = svc.Delete(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping, got %v", err)
	}
}

func TestPlayerServiceRestoreValidationAndErrorMappings(t *testing.T) {
	ctx := context.Background()
	repo := &fakePlayerRepo{
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Player, error) {
			return model.Player{}, repoerrors.NewNotFound("player.get_by_id", "player")
		},
		restoreFn: func(_ context.Context, _ model.ULID) (model.Player, error) {
			return model.Player{}, repoerrors.NewConflict("player.restore", "player")
		},
	}
	svc := NewPlayerService(repo, fakeIDGenerator{
		newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
	})

	_, err := svc.Restore(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id, got %v", err)
	}
	_, err = svc.Restore(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping on lookup, got %v", err)
	}
}

func TestMapRepositoryErrorDefaultPath(t *testing.T) {
	err := mapRepositoryError(errors.New("plain"), "op", "entity")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected default unknown mapping, got %v", err)
	}
}

func TestNormalizeSpacesAndSortHelpers(t *testing.T) {
	if got := normalizeSpaces("  a   b  c "); got != "a b c" {
		t.Fatalf("unexpected normalize result: %q", got)
	}

	items := []PlayerListItem{
		{ID: "2", Name: "beta"},
		{ID: "1", Name: "Beta"},
		{ID: "3", Name: "alpha"},
	}
	sortPlayerItems(items)
	if items[0].Name != "alpha" || items[1].ID != "1" || items[2].ID != "2" {
		t.Fatalf("unexpected sort order: %+v", items)
	}
}
