package service

import (
	"context"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
)

func TestDefaultPoliciesAndConstructors(t *testing.T) {
	if _, err := (DefaultCampaignNamePolicy{}).NormalizeAndValidate(""); err == nil {
		t.Fatalf("expected campaign name validation")
	}
	if out, err := (DefaultCampaignNamePolicy{}).NormalizeAndValidate("  Alpha  "); err != nil || out != "Alpha" {
		t.Fatalf("expected campaign name normalization")
	}
	if _, err := (DefaultTagNamePolicy{}).NormalizeAndValidate(""); err == nil {
		t.Fatalf("expected tag name validation")
	}
	if out, err := (DefaultTagNamePolicy{}).NormalizeAndValidate("  Tag  "); err != nil || out != "Tag" {
		t.Fatalf("expected tag name normalization")
	}
	if _, _, err := (DefaultPlanePolicy{}).NormalizeAndValidate("", ""); err == nil {
		t.Fatalf("expected plane name validation")
	}

	worldRepo := &fakeWorldRepo{
		createFn: func(_ context.Context, w model.World) (model.World, error) { return w, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.World, error) { return model.World{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListWorldsParams) ([]model.World, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdateWorldParams) (model.World, error) {
			return model.World{}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}
	if NewWorldService(worldRepo, fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}) == nil {
		t.Fatalf("expected world constructor service")
	}

	campaignRepo := &fakeCampaignRepo{
		createFn: func(_ context.Context, c model.Campaign) (model.Campaign, error) { return c, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Campaign, error) { return model.Campaign{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListCampaignsParams) ([]model.Campaign, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdateCampaignParams) (model.Campaign, error) {
			return model.Campaign{}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
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
	if NewCampaignService(campaignRepo, campaignPlayerRepo, fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}) == nil {
		t.Fatalf("expected campaign constructor service")
	}

	planeRepo := &fakePlaneRepo{
		createFn: func(_ context.Context, p model.Plane) (model.Plane, error) { return p, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Plane, error) { return model.Plane{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListPlanesParams) ([]model.Plane, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdatePlaneParams) (model.Plane, error) {
			return model.Plane{}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}
	if NewPlaneService(planeRepo, fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}) == nil {
		t.Fatalf("expected plane constructor service")
	}

	sessionRepo := &fakeSessionRepo{
		createFn: func(_ context.Context, s model.Session) (model.Session, error) { return s, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Session, error) { return model.Session{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListSessionsParams) ([]model.Session, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdateSessionParams) (model.Session, error) {
			return model.Session{}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}
	if NewSessionService(sessionRepo, fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}) == nil {
		t.Fatalf("expected session constructor service")
	}

	tagRepo := &fakeTagRepo{
		createFn: func(_ context.Context, tag model.Tag) (model.Tag, error) { return tag, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Tag, error) { return model.Tag{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListTagsParams) ([]model.Tag, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdateTagParams) (model.Tag, error) { return model.Tag{}, nil },
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}
	if NewTagService(tagRepo, fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}) == nil {
		t.Fatalf("expected tag constructor service")
	}

	noteSvc := NewNoteService(
		&fakeNoteRepo{
			createFn: func(_ context.Context, n model.Note) (model.Note, error) { return n, nil },
			getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Note, error) { return model.Note{}, nil },
			listFn:   func(_ context.Context, _ interfaces.ListNotesParams) ([]model.Note, error) { return nil, nil },
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteParams) (model.Note, error) { return model.Note{}, nil },
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		&fakeNoteOwnerRepo{
			createFn: func(_ context.Context, rel model.NoteOwner) (model.NoteOwner, error) { return rel, nil },
			getFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID, _ bool) (model.NoteOwner, error) {
				return model.NoteOwner{}, nil
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, nil
			},
			listByOwnerFn: func(_ context.Context, _ constants.OwnerType, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID) error { return nil },
		},
		&fakeNoteTagRepo{
			createFn: func(_ context.Context, rel model.NoteTag) (model.NoteTag, error) { return rel, nil },
			getFn:    func(_ context.Context, _, _ model.ULID, _ bool) (model.NoteTag, error) { return model.NoteTag{}, nil },
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, nil
			},
			listByTagFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, nil
			},
			deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
		},
		&fakeNoteLinkRepo{
			createFn:  func(_ context.Context, link model.NoteLink) (model.NoteLink, error) { return link, nil },
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteLink, error) { return model.NoteLink{}, nil },
			listBySourceFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, nil
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, nil
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
				return model.NoteLink{}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		&fakeNoteAssetRepo{
			createFn:  func(_ context.Context, asset model.NoteAsset) (model.NoteAsset, error) { return asset, nil },
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteAsset, error) { return model.NoteAsset{}, nil },
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
				return nil, nil
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
				return model.NoteAsset{}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		&fakeMapPlacementRepo{
			createFn: func(_ context.Context, p model.MapNotePlacement) (model.MapNotePlacement, error) { return p, nil },
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, nil
			},
			listByMapFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, nil
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, nil
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }},
	)
	if noteSvc == nil {
		t.Fatalf("expected note constructor service")
	}

	planeItems := toPlaneListItems([]model.Plane{{ID: "1"}})
	if len(planeItems) != 1 || planeItems[0].ID != "1" {
		t.Fatalf("unexpected plane list mapping")
	}
}

func TestCampaignRelationHappyPaths(t *testing.T) {
	ctx := context.Background()
	repo := &fakeCampaignRepo{
		createFn: func(_ context.Context, c model.Campaign) (model.Campaign, error) { return c, nil },
		getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Campaign, error) { return model.Campaign{}, nil },
		listFn:   func(_ context.Context, _ interfaces.ListCampaignsParams) ([]model.Campaign, error) { return nil, nil },
		updateFn: func(_ context.Context, _ interfaces.UpdateCampaignParams) (model.Campaign, error) {
			return model.Campaign{}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}
	campaignPlayerRepo := &fakeCampaignPlayerRepo{
		createFn: func(_ context.Context, rel model.CampaignPlayer) (model.CampaignPlayer, error) { return rel, nil },
		getFn: func(_ context.Context, campaignID, playerID model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{CampaignID: campaignID, PlayerID: playerID}, nil
		},
		listByCampaignFn: func(_ context.Context, campaignID model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return []model.CampaignPlayer{{CampaignID: campaignID}}, nil
		},
		listByPlayerFn: func(_ context.Context, playerID model.ULID, _ interfaces.ListCampaignPlayersParams) ([]model.CampaignPlayer, error) {
			return []model.CampaignPlayer{{PlayerID: playerID}}, nil
		},
		deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
	}
	svc := NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              fakeClock{now: time.Now().UTC()},
		NamePolicy:         DefaultCampaignNamePolicy{},
		IDGenerator:        fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }},
	})

	if _, err := svc.AddPlayer(ctx, "01C", "01P"); err != nil {
		t.Fatalf("expected add player success: %v", err)
	}
	if err := svc.RemovePlayer(ctx, "01C", "01P"); err != nil {
		t.Fatalf("expected remove player success: %v", err)
	}
	if _, err := svc.GetPlayerRelation(ctx, "01C", "01P", false); err != nil {
		t.Fatalf("expected get relation success: %v", err)
	}
	if _, err := svc.ListPlayers(ctx, "01C", ListCampaignsParams{Offset: 0, Limit: 10}); err != nil {
		t.Fatalf("expected list players success: %v", err)
	}
	if _, err := svc.ListCampaignsForPlayer(ctx, "01P", ListCampaignsParams{Offset: 0, Limit: 10}); err != nil {
		t.Fatalf("expected list campaigns by player success: %v", err)
	}

	if err := svc.RemovePlayer(ctx, "", "01P"); err == nil {
		t.Fatalf("expected campaign id validation")
	}
	if err := svc.RemovePlayer(ctx, "01C", ""); err == nil {
		t.Fatalf("expected player id validation")
	}
	if _, err := svc.GetPlayerRelation(ctx, "", "01P", false); err == nil {
		t.Fatalf("expected campaign id validation")
	}
	if _, err := svc.GetPlayerRelation(ctx, "01C", "", false); err == nil {
		t.Fatalf("expected player id validation")
	}
	if _, err := svc.ListPlayers(ctx, "01C", ListCampaignsParams{Offset: -1, Limit: 10}); err == nil {
		t.Fatalf("expected offset validation")
	}
	if _, err := svc.ListPlayers(ctx, "01C", ListCampaignsParams{Offset: 0, Limit: 0}); err == nil {
		t.Fatalf("expected limit validation")
	}
	if _, err := svc.ListCampaignsForPlayer(ctx, "01P", ListCampaignsParams{Offset: -1, Limit: 10}); err == nil {
		t.Fatalf("expected offset validation")
	}
	if _, err := svc.ListCampaignsForPlayer(ctx, "01P", ListCampaignsParams{Offset: 0, Limit: 0}); err == nil {
		t.Fatalf("expected limit validation")
	}
}
