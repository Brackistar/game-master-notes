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

type fakeTagRepo struct {
	createFn func(context.Context, model.Tag) (model.Tag, error)
	getFn    func(context.Context, model.ULID, bool) (model.Tag, error)
	listFn   func(context.Context, interfaces.ListTagsParams) ([]model.Tag, error)
	updateFn func(context.Context, interfaces.UpdateTagParams) (model.Tag, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeTagRepo) Create(ctx context.Context, tag model.Tag) (model.Tag, error) {
	return f.createFn(ctx, tag)
}
func (f *fakeTagRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeTagRepo) List(ctx context.Context, params interfaces.ListTagsParams) ([]model.Tag, error) {
	return f.listFn(ctx, params)
}
func (f *fakeTagRepo) Update(ctx context.Context, params interfaces.UpdateTagParams) (model.Tag, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeTagRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeTagPolicy struct {
	validateFn func(name string) (string, error)
}

func (f fakeTagPolicy) NormalizeAndValidate(name string) (string, error) {
	return f.validateFn(name)
}

func TestTagServiceFlows(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	repo := &fakeTagRepo{
		createFn: func(_ context.Context, tag model.Tag) (model.Tag, error) { return tag, nil },
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Tag, error) {
			return model.Tag{}, repoerrors.NewNotFound("tag.get_by_id", "tag")
		},
		listFn: func(_ context.Context, _ interfaces.ListTagsParams) ([]model.Tag, error) {
			return nil, repoerrors.NewConflict("tag.list", "tag")
		},
		updateFn: func(_ context.Context, _ interfaces.UpdateTagParams) (model.Tag, error) {
			return model.Tag{}, repoerrors.NewConflict("tag.update", "tag")
		},
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("tag.delete", "tag", errors.New("db"))
		},
	}
	svc := NewTagServiceWithDeps(TagServiceDeps{
		Repo:  repo,
		Clock: fakeClock{now: now},
		NamePolicy: fakeTagPolicy{
			validateFn: func(name string) (string, error) {
				if name == "bad" {
					return "", errors.New("invalid")
				}
				return "tag", nil
			},
		},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01H", nil }},
	})

	emptyID := model.ULID("")
	_, err := svc.Create(ctx, CreateTagParams{Name: "bad"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected name validation")
	}
	_, err = svc.Create(ctx, CreateTagParams{Name: "ok", CampaignID: &emptyID})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected campaign empty validation")
	}
	created, err := svc.Create(ctx, CreateTagParams{Name: "ok"})
	if err != nil || created.ID == "" {
		t.Fatalf("expected create success")
	}

	_, err = svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.GetByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping")
	}
	_, err = svc.List(ctx, ListTagsParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected offset validation")
	}
	_, err = svc.List(ctx, ListTagsParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected limit validation")
	}
	_, err = svc.List(ctx, ListTagsParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping")
	}
	_, err = svc.Update(ctx, UpdateTagParams{ID: "", Name: "ok", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.Update(ctx, UpdateTagParams{ID: "01", Name: "ok", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected version validation")
	}
	_, err = svc.Update(ctx, UpdateTagParams{ID: "01", Name: "bad", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected name validation")
	}
	_, err = svc.Update(ctx, UpdateTagParams{ID: "01", Name: "ok", ExpectedVersion: 1, CampaignID: &emptyID})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected campaign empty validation")
	}
	_, err = svc.Update(ctx, UpdateTagParams{ID: "01", Name: "ok", ExpectedVersion: 1})
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

func TestTagServiceNilDependenciesAndHelpers(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeTagRepo{createFn: func(_ context.Context, t model.Tag) (model.Tag, error) { return t, nil }}
	tests := []TagServiceDeps{
		{Clock: fakeClock{now: now}, NamePolicy: DefaultTagNamePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, NamePolicy: DefaultTagNamePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: repo, Clock: fakeClock{now: now}, NamePolicy: DefaultTagNamePolicy{}},
	}
	for _, tc := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = NewTagServiceWithDeps(tc)
		}()
	}

	items := toTagListItems([]model.Tag{{ID: "1", Name: "a"}})
	if len(items) != 1 || items[0].Name != "a" {
		t.Fatalf("unexpected mapping")
	}
}
