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

type fakeCampaignRepo struct {
	createFn func(context.Context, model.Campaign) (model.Campaign, error)
	getFn    func(context.Context, model.ULID, bool) (model.Campaign, error)
	listFn   func(context.Context, interfaces.ListCampaignsParams) ([]model.Campaign, error)
	updateFn func(context.Context, interfaces.UpdateCampaignParams) (model.Campaign, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeCampaignRepo) Create(ctx context.Context, campaign model.Campaign) (model.Campaign, error) {
	return f.createFn(ctx, campaign)
}
func (f *fakeCampaignRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeCampaignRepo) List(ctx context.Context, params interfaces.ListCampaignsParams) ([]model.Campaign, error) {
	return f.listFn(ctx, params)
}
func (f *fakeCampaignRepo) Update(ctx context.Context, params interfaces.UpdateCampaignParams) (model.Campaign, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeCampaignRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeCampaignPlayerRepo struct {
	createFn         func(context.Context, model.CampaignPlayer) (model.CampaignPlayer, error)
	getFn            func(context.Context, model.ULID, model.ULID, bool) (model.CampaignPlayer, error)
	listByCampaignFn func(context.Context, model.ULID, interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error)
	listByPlayerFn   func(context.Context, model.ULID, interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error)
	deleteFn         func(context.Context, model.ULID, model.ULID) error
}

func (f *fakeCampaignPlayerRepo) Create(ctx context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) {
	return f.createFn(ctx, rel)
}
func (f *fakeCampaignPlayerRepo) Get(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error) {
	return f.getFn(ctx, campaignID, playerID, includeDeleted)
}
func (f *fakeCampaignPlayerRepo) ListByCampaign(ctx context.Context, campaignID model.ULID, params interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
	return f.listByCampaignFn(ctx, campaignID, params)
}
func (f *fakeCampaignPlayerRepo) ListByPlayer(ctx context.Context, playerID model.ULID, params interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
	return f.listByPlayerFn(ctx, playerID, params)
}
func (f *fakeCampaignPlayerRepo) Delete(ctx context.Context, campaignID, playerID model.ULID) error {
	return f.deleteFn(ctx, campaignID, playerID)
}

type fakeCampaignNamePolicy struct {
	validateFn func(name string) (string, error)
}

func (f fakeCampaignNamePolicy) NormalizeAndValidate(name string) (string, error) {
	return f.validateFn(name)
}

func TestNewCampaignServiceWithDepsFailsOnNilDependencies(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeCampaignRepo{
		createFn: func(_ context.Context, campaign model.Campaign) (model.Campaign, error) { return campaign, nil },
	}
	campaignPlayerRepo := &fakeCampaignPlayerRepo{
		createFn: func(_ context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) { return rel, nil },
		getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{}, nil
		},
		listByCampaignFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		listByPlayerFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
	}

	tests := []struct {
		name string
		deps CampaignServiceDeps
	}{
		{
			name: "nil repo",
			deps: CampaignServiceDeps{
				Clock:       fakeClock{now: now},
				NamePolicy:  DefaultCampaignNamePolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil campaign player repo",
			deps: CampaignServiceDeps{
				Repo:        repo,
				Clock:       fakeClock{now: now},
				NamePolicy:  DefaultCampaignNamePolicy{},
				IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil clock",
			deps: CampaignServiceDeps{
				Repo:               repo,
				CampaignPlayerRepo: campaignPlayerRepo,
				NamePolicy:         DefaultCampaignNamePolicy{},
				IDGenerator:        fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil name policy",
			deps: CampaignServiceDeps{
				Repo:               repo,
				CampaignPlayerRepo: campaignPlayerRepo,
				Clock:              fakeClock{now: now},
				IDGenerator:        fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
			},
		},
		{
			name: "nil id generator",
			deps: CampaignServiceDeps{
				Repo:               repo,
				CampaignPlayerRepo: campaignPlayerRepo,
				Clock:              fakeClock{now: now},
				NamePolicy:         DefaultCampaignNamePolicy{},
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
			_ = NewCampaignServiceWithDeps(tc.deps)
		})
	}
}

func TestCampaignServiceCreateUsesPolicyAndGeneratedID(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	var got model.Campaign

	repo := &fakeCampaignRepo{
		createFn: func(_ context.Context, campaign model.Campaign) (model.Campaign, error) {
			got = campaign
			return campaign, nil
		},
	}
	campaignPlayerRepo := &fakeCampaignPlayerRepo{
		createFn: func(_ context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) { return rel, nil },
		getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{}, nil
		},
		listByCampaignFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		listByPlayerFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
	}
	svc := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: now},
		NamePolicy: fakeCampaignNamePolicy{
			validateFn: func(name string) (string, error) { return "Campaign Name", nil },
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})

	out, err := svc.Create(ctx, CreateCampaignParams{
		WorldID:   "01HWWWWWWWWWWWWWWWWWWWWWWW",
		Name:      "ignored",
		StartDate: &start,
		EndDate:   &end,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if out.ID != "01HZZZZZZZZZZZZZZZZZZZZZZZ" {
		t.Fatalf("unexpected generated id: %s", out.ID)
	}
	if got.Name != "Campaign Name" {
		t.Fatalf("expected normalized name, got %q", got.Name)
	}
	if !got.AuditFields.CreatedAt.Equal(now) {
		t.Fatalf("expected injected clock timestamp")
	}
}

func TestCampaignServiceCreateValidationAndIdGeneratorErrors(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	repo := &fakeCampaignRepo{
		createFn: func(_ context.Context, campaign model.Campaign) (model.Campaign, error) { return campaign, nil },
	}
	campaignPlayerRepo := &fakeCampaignPlayerRepo{
		createFn: func(_ context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) { return rel, nil },
		getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{}, nil
		},
		listByCampaignFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		listByPlayerFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, nil
		},
		deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
	}

	svc := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: time.Now().UTC()},
		NamePolicy: fakeCampaignNamePolicy{
			validateFn: func(name string) (string, error) { return "", errors.New("invalid name") },
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})

	_, err := svc.Create(ctx, CreateCampaignParams{WorldID: "", Name: "x"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for missing world id, got %v", err)
	}

	_, err = svc.Create(ctx, CreateCampaignParams{WorldID: "01HWWWWWWWWWWWWWWWWWWWWWWW", Name: "x"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation from name policy, got %v", err)
	}

	svcValidName := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: time.Now().UTC()},
		NamePolicy: fakeCampaignNamePolicy{
			validateFn: func(name string) (string, error) { return name, nil },
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})
	_, err = svcValidName.Create(ctx, CreateCampaignParams{
		WorldID:   "01HWWWWWWWWWWWWWWWWWWWWWWW",
		Name:      "name",
		StartDate: &start,
		EndDate:   &end,
	})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for invalid date range, got %v", err)
	}

	svcIDErr := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: time.Now().UTC()},
		NamePolicy: fakeCampaignNamePolicy{
			validateFn: func(name string) (string, error) { return name, nil },
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "", errors.New("idgen down") },
		},
	})
	_, err = svcIDErr.Create(ctx, CreateCampaignParams{
		WorldID: "01HWWWWWWWWWWWWWWWWWWWWWWW",
		Name:    "name",
	})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown on idgen failure, got %v", err)
	}
}

func TestCampaignServiceGetListUpdateDeleteValidationAndMappings(t *testing.T) {
	ctx := context.Background()
	repo := &fakeCampaignRepo{
		getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Campaign, error) {
			return model.Campaign{}, repoerrors.NewNotFound("campaign.get_by_id", "campaign")
		},
		listFn: func(_ context.Context, _ interfaces.ListCampaignsParams) ([]model.Campaign, error) {
			return nil, repoerrors.NewConflict("campaign.list", "campaign")
		},
		updateFn: func(_ context.Context, _ interfaces.UpdateCampaignParams) (model.Campaign, error) {
			return model.Campaign{}, repoerrors.NewConflict("campaign.update", "campaign")
		},
		deleteFn: func(_ context.Context, _ model.ULID) error {
			return repoerrors.WrapUnknown("campaign.delete", "campaign", errors.New("db"))
		},
	}
	campaignPlayerRepo := &fakeCampaignPlayerRepo{
		createFn: func(_ context.Context, _ model.CampaignPlayer) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{}, repoerrors.NewConflict("campaign_player.create", "campaign_player")
		},
		getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{}, repoerrors.NewNotFound("campaign_player.get", "campaign_player")
		},
		listByCampaignFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, repoerrors.WrapUnknown("campaign_player.list_by_campaign", "campaign_player", errors.New("db"))
		},
		listByPlayerFn: func(_ context.Context, _ model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return nil, repoerrors.WrapUnknown("campaign_player.list_by_player", "campaign_player", errors.New("db"))
		},
		deleteFn: func(_ context.Context, _, _ model.ULID) error {
			return repoerrors.NewConflict("campaign_player.delete", "campaign_player")
		},
	}
	svc := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: time.Now().UTC()},
		NamePolicy: fakeCampaignNamePolicy{
			validateFn: func(name string) (string, error) {
				if name == "bad" {
					return "", errors.New("invalid")
				}
				return name, nil
			},
		},
		IDGenerator: fakeIDGenerator{
			newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil },
		},
	})

	_, err := svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id, got %v", err)
	}
	_, err = svc.GetByID(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}

	_, err = svc.List(ctx, ListCampaignsParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for negative offset, got %v", err)
	}
	_, err = svc.List(ctx, ListCampaignsParams{Offset: 0, Limit: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for limit, got %v", err)
	}
	_, err = svc.List(ctx, ListCampaignsParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping, got %v", err)
	}

	_, err = svc.Update(ctx, UpdateCampaignParams{ID: "", Name: "name", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for id, got %v", err)
	}
	_, err = svc.Update(ctx, UpdateCampaignParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "name", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for expected_version, got %v", err)
	}
	_, err = svc.Update(ctx, UpdateCampaignParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "bad", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation from name policy, got %v", err)
	}
	_, err = svc.Update(ctx, UpdateCampaignParams{ID: "01HZZZZZZZZZZZZZZZZZZZZZZZ", Name: "name", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping, got %v", err)
	}

	err = svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty id, got %v", err)
	}
	err = svc.Delete(ctx, "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping, got %v", err)
	}

	_, err = svc.AddPlayer(ctx, "", "01HPPPPPPPPPPPPPPPPPPPPPPPP")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected validation for empty campaign_id, got %v", err)
	}
	_, err = svc.AddPlayer(ctx, "01HCCCCCCCCCCCCCCCCCCCCCCCC", "01HPPPPPPPPPPPPPPPPPPPPPPPP")
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping on add player, got %v", err)
	}
	err = svc.RemovePlayer(ctx, "01HCCCCCCCCCCCCCCCCCCCCCCCC", "01HPPPPPPPPPPPPPPPPPPPPPPPP")
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected conflict mapping on remove player, got %v", err)
	}
	_, err = svc.GetPlayerRelation(ctx, "01HCCCCCCCCCCCCCCCCCCCCCCCC", "01HPPPPPPPPPPPPPPPPPPPPPPPP", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected not found mapping on get player relation, got %v", err)
	}
	_, err = svc.ListPlayers(ctx, "01HCCCCCCCCCCCCCCCCCCCCCCCC", ListCampaignsParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping on list players, got %v", err)
	}
	_, err = svc.ListCampaignsForPlayer(ctx, "01HPPPPPPPPPPPPPPPPPPPPPPPP", ListCampaignsParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected unknown mapping on list campaigns by player, got %v", err)
	}
}

func TestCampaignServiceHelpers(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if err := validateCampaignDateRange(&start, &end); err != nil {
		t.Fatalf("expected valid range, got %v", err)
	}
	if err := validateCampaignDateRange(nil, &end); err != nil {
		t.Fatalf("expected nil start to be valid, got %v", err)
	}
	if err := validateCampaignDateRange(&end, &start); err == nil {
		t.Fatalf("expected invalid range error")
	}

	campaigns := []model.Campaign{
		{
			ID:      "01HZZZZZZZZZZZZZZZZZZZZZZZ",
			WorldID: "01HWWWWWWWWWWWWWWWWWWWWWWW",
			Name:    "Test",
			AuditFields: model.AuditFields{
				Version: 2,
			},
		},
	}
	items := toCampaignListItems(campaigns)
	if len(items) != 1 {
		t.Fatalf("expected one item")
	}
	if items[0].WorldID != campaigns[0].WorldID || items[0].Version != campaigns[0].AuditFields.Version {
		t.Fatalf("unexpected mapping: %+v", items[0])
	}
}
