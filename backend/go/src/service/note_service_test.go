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
)

type fakeNoteRepo struct {
	createFn func(context.Context, model.Note) (model.Note, error)
	getFn    func(context.Context, model.ULID, bool) (model.Note, error)
	listFn   func(context.Context, interfaces.ListNotesParams) ([]model.Note, error)
	updateFn func(context.Context, interfaces.UpdateNoteParams) (model.Note, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeNoteRepo) Create(ctx context.Context, note model.Note) (model.Note, error) {
	return f.createFn(ctx, note)
}
func (f *fakeNoteRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Note, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeNoteRepo) List(ctx context.Context, params interfaces.ListNotesParams) ([]model.Note, error) {
	return f.listFn(ctx, params)
}
func (f *fakeNoteRepo) Update(ctx context.Context, params interfaces.UpdateNoteParams) (model.Note, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeNoteRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeNoteOwnerRepo struct {
	createFn      func(context.Context, model.NoteOwner) (model.NoteOwner, error)
	getFn         func(context.Context, model.ULID, constants.OwnerType, model.ULID, bool) (model.NoteOwner, error)
	listByNoteFn  func(context.Context, model.ULID, interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error)
	listByOwnerFn func(context.Context, constants.OwnerType, model.ULID, interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error)
	deleteFn      func(context.Context, model.ULID, constants.OwnerType, model.ULID) error
}

func (f *fakeNoteOwnerRepo) Create(ctx context.Context, rel model.NoteOwner) (model.NoteOwner, error) {
	return f.createFn(ctx, rel)
}
func (f *fakeNoteOwnerRepo) Get(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, includeDeleted bool) (model.NoteOwner, error) {
	return f.getFn(ctx, noteID, ownerType, ownerID, includeDeleted)
}
func (f *fakeNoteOwnerRepo) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
	return f.listByNoteFn(ctx, noteID, params)
}
func (f *fakeNoteOwnerRepo) ListByOwner(ctx context.Context, ownerType constants.OwnerType, ownerID model.ULID, params interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
	return f.listByOwnerFn(ctx, ownerType, ownerID, params)
}
func (f *fakeNoteOwnerRepo) Delete(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID) error {
	return f.deleteFn(ctx, noteID, ownerType, ownerID)
}

type fakeNoteTagRepo struct {
	createFn     func(context.Context, model.NoteTag) (model.NoteTag, error)
	getFn        func(context.Context, model.ULID, model.ULID, bool) (model.NoteTag, error)
	listByNoteFn func(context.Context, model.ULID, interfaces.ListNoteTagsParams) ([]model.NoteTag, error)
	listByTagFn  func(context.Context, model.ULID, interfaces.ListNoteTagsParams) ([]model.NoteTag, error)
	deleteFn     func(context.Context, model.ULID, model.ULID) error
}

func (f *fakeNoteTagRepo) Create(ctx context.Context, rel model.NoteTag) (model.NoteTag, error) {
	return f.createFn(ctx, rel)
}
func (f *fakeNoteTagRepo) Get(ctx context.Context, noteID, tagID model.ULID, includeDeleted bool) (model.NoteTag, error) {
	return f.getFn(ctx, noteID, tagID, includeDeleted)
}
func (f *fakeNoteTagRepo) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
	return f.listByNoteFn(ctx, noteID, params)
}
func (f *fakeNoteTagRepo) ListByTag(ctx context.Context, tagID model.ULID, params interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
	return f.listByTagFn(ctx, tagID, params)
}
func (f *fakeNoteTagRepo) Delete(ctx context.Context, noteID, tagID model.ULID) error {
	return f.deleteFn(ctx, noteID, tagID)
}

type fakeNoteLinkRepo struct {
	createFn       func(context.Context, model.NoteLink) (model.NoteLink, error)
	getByIDFn      func(context.Context, model.ULID, bool) (model.NoteLink, error)
	listBySourceFn func(context.Context, model.ULID, interfaces.ListNoteLinksParams) ([]model.NoteLink, error)
	listByTargetFn func(context.Context, model.ULID, interfaces.ListNoteLinksParams) ([]model.NoteLink, error)
	updateFn       func(context.Context, interfaces.UpdateNoteLinkParams) (model.NoteLink, error)
	deleteFn       func(context.Context, model.ULID) error
}

func (f *fakeNoteLinkRepo) Create(ctx context.Context, link model.NoteLink) (model.NoteLink, error) {
	return f.createFn(ctx, link)
}
func (f *fakeNoteLinkRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteLink, error) {
	return f.getByIDFn(ctx, id, includeDeleted)
}
func (f *fakeNoteLinkRepo) ListBySource(ctx context.Context, sourceNoteID model.ULID, params interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
	return f.listBySourceFn(ctx, sourceNoteID, params)
}
func (f *fakeNoteLinkRepo) ListByTarget(ctx context.Context, targetNoteID model.ULID, params interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
	return f.listByTargetFn(ctx, targetNoteID, params)
}
func (f *fakeNoteLinkRepo) Update(ctx context.Context, params interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeNoteLinkRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeNoteAssetRepo struct {
	createFn     func(context.Context, model.NoteAsset) (model.NoteAsset, error)
	getByIDFn    func(context.Context, model.ULID, bool) (model.NoteAsset, error)
	listByNoteFn func(context.Context, model.ULID, interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error)
	updateFn     func(context.Context, interfaces.UpdateNoteAssetParams) (model.NoteAsset, error)
	deleteFn     func(context.Context, model.ULID) error
}

func (f *fakeNoteAssetRepo) Create(ctx context.Context, asset model.NoteAsset) (model.NoteAsset, error) {
	return f.createFn(ctx, asset)
}
func (f *fakeNoteAssetRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteAsset, error) {
	return f.getByIDFn(ctx, id, includeDeleted)
}
func (f *fakeNoteAssetRepo) ListByNote(ctx context.Context, noteID model.ULID, params interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
	return f.listByNoteFn(ctx, noteID, params)
}
func (f *fakeNoteAssetRepo) Update(ctx context.Context, params interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeNoteAssetRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeMapPlacementRepo struct {
	createFn       func(context.Context, model.MapNotePlacement) (model.MapNotePlacement, error)
	getByIDFn      func(context.Context, model.ULID, bool) (model.MapNotePlacement, error)
	listByMapFn    func(context.Context, model.ULID, interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error)
	listByTargetFn func(context.Context, model.ULID, interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error)
	updateFn       func(context.Context, interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error)
	deleteFn       func(context.Context, model.ULID) error
}

func (f *fakeMapPlacementRepo) Create(ctx context.Context, placement model.MapNotePlacement) (model.MapNotePlacement, error) {
	return f.createFn(ctx, placement)
}
func (f *fakeMapPlacementRepo) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.MapNotePlacement, error) {
	return f.getByIDFn(ctx, id, includeDeleted)
}
func (f *fakeMapPlacementRepo) ListByMapNote(ctx context.Context, mapNoteID model.ULID, params interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
	return f.listByMapFn(ctx, mapNoteID, params)
}
func (f *fakeMapPlacementRepo) ListByTargetNote(ctx context.Context, targetNoteID model.ULID, params interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
	return f.listByTargetFn(ctx, targetNoteID, params)
}
func (f *fakeMapPlacementRepo) Update(ctx context.Context, params interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeMapPlacementRepo) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

type fakeNotePolicy struct {
	validateFn func(title, content string, noteType constants.NoteType, metadata []byte) (string, string, []byte, error)
}

func (f fakeNotePolicy) NormalizeAndValidate(title, content string, noteType constants.NoteType, metadata []byte) (string, string, []byte, error) {
	return f.validateFn(title, content, noteType, metadata)
}

func newNoteServiceForTests() *NoteService {
	return NewNoteServiceWithDeps(NoteServiceDeps{
		Repo: &fakeNoteRepo{
			createFn: func(_ context.Context, n model.Note) (model.Note, error) { return n, nil },
			getFn:    func(_ context.Context, id model.ULID, _ bool) (model.Note, error) { return model.Note{ID: id}, nil },
			listFn: func(_ context.Context, _ interfaces.ListNotesParams) ([]model.Note, error) {
				return []model.Note{{ID: "1"}}, nil
			},
			updateFn: func(_ context.Context, p interfaces.UpdateNoteParams) (model.Note, error) {
				return model.Note{ID: p.ID}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		NoteOwnerRepo: &fakeNoteOwnerRepo{
			createFn: func(_ context.Context, rel model.NoteOwner) (model.NoteOwner, error) { return rel, nil },
			getFn: func(_ context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, _ bool) (model.NoteOwner, error) {
				return model.NoteOwner{NoteID: noteID, OwnerType: ownerType, OwnerID: ownerID}, nil
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return []model.NoteOwner{{NoteID: "1"}}, nil
			},
			listByOwnerFn: func(_ context.Context, _ constants.OwnerType, ownerID model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return []model.NoteOwner{{OwnerID: ownerID}}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID) error { return nil },
		},
		NoteTagRepo: &fakeNoteTagRepo{
			createFn: func(_ context.Context, rel model.NoteTag) (model.NoteTag, error) { return rel, nil },
			getFn: func(_ context.Context, noteID, tagID model.ULID, _ bool) (model.NoteTag, error) {
				return model.NoteTag{NoteID: noteID, TagID: tagID}, nil
			},
			listByNoteFn: func(_ context.Context, noteID model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return []model.NoteTag{{NoteID: noteID}}, nil
			},
			listByTagFn: func(_ context.Context, tagID model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return []model.NoteTag{{TagID: tagID}}, nil
			},
			deleteFn: func(_ context.Context, _, _ model.ULID) error { return nil },
		},
		NoteLinkRepo: &fakeNoteLinkRepo{
			createFn: func(_ context.Context, link model.NoteLink) (model.NoteLink, error) { return link, nil },
			getByIDFn: func(_ context.Context, id model.ULID, _ bool) (model.NoteLink, error) {
				return model.NoteLink{ID: id}, nil
			},
			listBySourceFn: func(_ context.Context, source model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return []model.NoteLink{{SourceNoteID: source}}, nil
			},
			listByTargetFn: func(_ context.Context, target model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return []model.NoteLink{{TargetNoteID: target}}, nil
			},
			updateFn: func(_ context.Context, p interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
				return model.NoteLink{ID: p.ID, LinkType: p.LinkType}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		NoteAssetRepo: &fakeNoteAssetRepo{
			createFn: func(_ context.Context, asset model.NoteAsset) (model.NoteAsset, error) { return asset, nil },
			getByIDFn: func(_ context.Context, id model.ULID, _ bool) (model.NoteAsset, error) {
				return model.NoteAsset{ID: id}, nil
			},
			listByNoteFn: func(_ context.Context, noteID model.ULID, _ interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
				return []model.NoteAsset{{NoteID: noteID}}, nil
			},
			updateFn: func(_ context.Context, p interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
				return model.NoteAsset{ID: p.ID, AssetType: p.AssetType}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		MapNotePlacementRepo: &fakeMapPlacementRepo{
			createFn: func(_ context.Context, p model.MapNotePlacement) (model.MapNotePlacement, error) { return p, nil },
			getByIDFn: func(_ context.Context, id model.ULID, _ bool) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{ID: id}, nil
			},
			listByMapFn: func(_ context.Context, mapID model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return []model.MapNotePlacement{{MapNoteID: mapID}}, nil
			},
			listByTargetFn: func(_ context.Context, targetID model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return []model.MapNotePlacement{{TargetNoteID: targetID}}, nil
			},
			updateFn: func(_ context.Context, p interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{ID: p.ID, X: p.X, Y: p.Y}, nil
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		Clock:       fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		Policy:      DefaultNotePolicy{},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01HZZZZZZZZZZZZZZZZZZZZZZZ", nil }},
	})
}

func TestNoteServiceNilDepsPanic(t *testing.T) {
	now := time.Now().UTC()
	tests := []NoteServiceDeps{
		{Clock: fakeClock{now: now}, Policy: DefaultNotePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
		{Repo: &fakeNoteRepo{}, Clock: fakeClock{now: now}, Policy: DefaultNotePolicy{}, IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01", nil }}},
	}
	for _, tc := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic")
				}
			}()
			_ = NewNoteServiceWithDeps(tc)
		}()
	}
}

func TestNoteServicePolicyAndHelpers(t *testing.T) {
	p := DefaultNotePolicy{}
	_, _, _, err := p.NormalizeAndValidate("", "", constants.General, []byte("{}"))
	if err == nil {
		t.Fatalf("expected title validation")
	}
	_, _, _, err = p.NormalizeAndValidate("Title", "", constants.NoteType(99), []byte("{}"))
	if err == nil {
		t.Fatalf("expected note type validation")
	}
	_, _, _, err = p.NormalizeAndValidate("Title", "", constants.General, []byte("{"))
	if err == nil {
		t.Fatalf("expected metadata validation")
	}
	_, _, metadata, err := p.NormalizeAndValidate("Title", " content ", constants.General, nil)
	if err != nil || string(metadata) != "{}" {
		t.Fatalf("expected normalized metadata")
	}

	if err := validateOffsetLimit(-1, 1); err == nil {
		t.Fatalf("expected offset validation")
	}
	if err := validateOffsetLimit(0, 0); err == nil {
		t.Fatalf("expected limit validation")
	}
	if err := validateTwoIDs("", "a", "1", "b"); err == nil {
		t.Fatalf("expected left id validation")
	}
	if err := validateTwoIDs("1", "a", "", "b"); err == nil {
		t.Fatalf("expected right id validation")
	}

	if !isValidNoteType(constants.General) || isValidNoteType(constants.NoteType(99)) {
		t.Fatalf("unexpected note type validity")
	}
	if !isValidOwnerType(constants.World) || isValidOwnerType(constants.OwnerType(99)) {
		t.Fatalf("unexpected owner type validity")
	}
	if !isValidNoteLinkType(constants.Related) || isValidNoteLinkType(constants.NoteLinkType(99)) {
		t.Fatalf("unexpected link type validity")
	}
	if !isValidAssetType(constants.Image) || isValidAssetType(constants.AssetType(99)) {
		t.Fatalf("unexpected asset type validity")
	}

	items := toNoteListItems([]model.Note{{ID: "1", Title: "T"}})
	if len(items) != 1 || items[0].Title != "T" {
		t.Fatalf("unexpected note list mapping")
	}
}

func TestNoteServiceCRUDAndOwnerTagFlows(t *testing.T) {
	ctx := context.Background()
	svc := newNoteServiceForTests()

	_, err := svc.Create(ctx, CreateNoteParams{Title: ""})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected create validation")
	}
	n, err := svc.Create(ctx, CreateNoteParams{Title: "title", NoteType: constants.General})
	if err != nil || n.ID == "" {
		t.Fatalf("expected note create success")
	}
	_, err = svc.GetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected get id validation")
	}
	_, err = svc.GetByID(ctx, "01N", false)
	if err != nil {
		t.Fatalf("expected get success")
	}
	_, err = svc.List(ctx, ListNotesParams{Offset: -1, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected list offset validation")
	}
	_, err = svc.List(ctx, ListNotesParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list success")
	}
	_, err = svc.Update(ctx, UpdateNoteParams{ID: "", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected update id validation")
	}
	_, err = svc.Update(ctx, UpdateNoteParams{ID: "01N", ExpectedVersion: 0})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected update version validation")
	}
	_, err = svc.Update(ctx, UpdateNoteParams{ID: "01N", ExpectedVersion: 1, Title: "t", NoteType: constants.General})
	if err != nil {
		t.Fatalf("expected update success")
	}
	err = svc.Delete(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected delete id validation")
	}
	if err = svc.Delete(ctx, "01N"); err != nil {
		t.Fatalf("expected delete success")
	}

	_, err = svc.AddOwner(ctx, AddNoteOwnerParams{NoteID: "", OwnerID: "01O", OwnerType: constants.World})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected owner id validation")
	}
	_, err = svc.AddOwner(ctx, AddNoteOwnerParams{NoteID: "01N", OwnerID: "01O", OwnerType: constants.OwnerType(99)})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected owner type validation")
	}
	_, err = svc.AddOwner(ctx, AddNoteOwnerParams{NoteID: "01N", OwnerID: "01O", OwnerType: constants.World})
	if err != nil {
		t.Fatalf("expected add owner success")
	}
	err = svc.RemoveOwner(ctx, "01N", constants.World, "01O")
	if err != nil {
		t.Fatalf("expected remove owner success")
	}
	_, err = svc.GetOwner(ctx, "01N", constants.World, "01O", false)
	if err != nil {
		t.Fatalf("expected get owner success")
	}
	_, err = svc.ListOwnersByNote(ctx, "01N", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list by note success")
	}
	_, err = svc.ListNotesByOwner(ctx, constants.World, "01O", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list by owner success")
	}

	_, err = svc.AddTag(ctx, AddNoteTagParams{NoteID: "", TagID: "01T"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected add tag validation")
	}
	_, err = svc.AddTag(ctx, AddNoteTagParams{NoteID: "01N", TagID: "01T"})
	if err != nil {
		t.Fatalf("expected add tag success")
	}
	if err = svc.RemoveTag(ctx, "01N", "01T"); err != nil {
		t.Fatalf("expected remove tag success")
	}
	_, err = svc.GetTag(ctx, "01N", "01T", false)
	if err != nil {
		t.Fatalf("expected get tag success")
	}
	_, err = svc.ListTagsByNote(ctx, "01N", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list tags by note success")
	}
	_, err = svc.ListNotesByTag(ctx, "01T", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list notes by tag success")
	}
}

func TestNoteServiceLinkAssetMapAndErrorMapping(t *testing.T) {
	ctx := context.Background()
	svc := newNoteServiceForTests()

	_, err := svc.CreateLink(ctx, CreateNoteLinkParams{SourceNoteID: "", TargetNoteID: "01", LinkType: constants.Related})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected link id validation")
	}
	_, err = svc.CreateLink(ctx, CreateNoteLinkParams{SourceNoteID: "01", TargetNoteID: "01", LinkType: constants.Related})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected self link validation")
	}
	_, err = svc.CreateLink(ctx, CreateNoteLinkParams{SourceNoteID: "01", TargetNoteID: "02", LinkType: constants.Related})
	if err != nil {
		t.Fatalf("expected create link success")
	}
	_, err = svc.GetLinkByID(ctx, "01L", false)
	if err != nil {
		t.Fatalf("expected get link success")
	}
	_, err = svc.ListLinksBySource(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list source success")
	}
	_, err = svc.ListLinksByTarget(ctx, "02", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list target success")
	}
	_, err = svc.UpdateLink(ctx, UpdateNoteLinkParams{ID: "01L", LinkType: constants.Related, ExpectedVersion: 1})
	if err != nil {
		t.Fatalf("expected update link success")
	}
	if err = svc.DeleteLink(ctx, "01L"); err != nil {
		t.Fatalf("expected delete link success")
	}

	_, err = svc.CreateAsset(ctx, CreateNoteAssetParams{NoteID: "01N", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png"})
	if err != nil {
		t.Fatalf("expected create asset success")
	}
	_, err = svc.GetAssetByID(ctx, "01A", false)
	if err != nil {
		t.Fatalf("expected get asset success")
	}
	_, err = svc.ListAssetsByNote(ctx, "01N", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list assets success")
	}
	_, err = svc.UpdateAsset(ctx, UpdateNoteAssetParams{ID: "01A", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png", ExpectedVersion: 1})
	if err != nil {
		t.Fatalf("expected update asset success")
	}
	if err = svc.DeleteAsset(ctx, "01A"); err != nil {
		t.Fatalf("expected delete asset success")
	}

	_, err = svc.UpsertMapPlacement(ctx, UpsertMapNotePlacementParams{MapNoteID: "01M", TargetNoteID: "01T", X: 10, Y: 20})
	if err != nil {
		t.Fatalf("expected upsert map success")
	}
	_, err = svc.GetMapPlacementByID(ctx, "01P", false)
	if err != nil {
		t.Fatalf("expected get map placement success")
	}
	_, err = svc.ListMapPlacementsByMap(ctx, "01M", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list map placements by map success")
	}
	_, err = svc.ListMapPlacementsByTarget(ctx, "01T", RelationListParams{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("expected list map placements by target success")
	}
	_, err = svc.UpdateMapPlacement(ctx, UpdateMapNotePlacementParams{ID: "01P", X: 1, Y: 2, ExpectedVersion: 1})
	if err != nil {
		t.Fatalf("expected update map placement success")
	}
	if err = svc.DeleteMapPlacement(ctx, "01P"); err != nil {
		t.Fatalf("expected delete map placement success")
	}

	svcErr := NewNoteServiceWithDeps(NoteServiceDeps{
		Repo: &fakeNoteRepo{
			createFn: func(_ context.Context, _ model.Note) (model.Note, error) {
				return model.Note{}, repoerrors.NewConflict("note.create", "note")
			},
			getFn: func(_ context.Context, _ model.ULID, _ bool) (model.Note, error) {
				return model.Note{}, repoerrors.NewNotFound("note.get_by_id", "note")
			},
			listFn: func(_ context.Context, _ interfaces.ListNotesParams) ([]model.Note, error) {
				return nil, repoerrors.WrapUnknown("note.list", "note", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteParams) (model.Note, error) {
				return model.Note{}, repoerrors.NewConflict("note.update", "note")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error { return repoerrors.NewNotFound("note.delete", "note") },
		},
		NoteOwnerRepo: &fakeNoteOwnerRepo{
			createFn: func(_ context.Context, _ model.NoteOwner) (model.NoteOwner, error) {
				return model.NoteOwner{}, repoerrors.NewConflict("note_owner.create", "note_owner")
			},
			getFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID, _ bool) (model.NoteOwner, error) {
				return model.NoteOwner{}, repoerrors.NewNotFound("note_owner.get", "note_owner")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, repoerrors.WrapUnknown("note_owner.list_by_note", "note_owner", errors.New("db"))
			},
			listByOwnerFn: func(_ context.Context, _ constants.OwnerType, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, repoerrors.WrapUnknown("note_owner.list_by_owner", "note_owner", errors.New("db"))
			},
			deleteFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID) error {
				return repoerrors.NewConflict("note_owner.delete", "note_owner")
			},
		},
		NoteTagRepo: &fakeNoteTagRepo{
			createFn: func(_ context.Context, _ model.NoteTag) (model.NoteTag, error) {
				return model.NoteTag{}, repoerrors.NewConflict("note_tag.create", "note_tag")
			},
			getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.NoteTag, error) {
				return model.NoteTag{}, repoerrors.NewNotFound("note_tag.get", "note_tag")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, repoerrors.WrapUnknown("note_tag.list_by_note", "note_tag", errors.New("db"))
			},
			listByTagFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, repoerrors.WrapUnknown("note_tag.list_by_tag", "note_tag", errors.New("db"))
			},
			deleteFn: func(_ context.Context, _, _ model.ULID) error {
				return repoerrors.NewConflict("note_tag.delete", "note_tag")
			},
		},
		NoteLinkRepo: &fakeNoteLinkRepo{
			createFn: func(_ context.Context, _ model.NoteLink) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewConflict("note_link.create", "note_link")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewNotFound("note_link.get_by_id", "note_link")
			},
			listBySourceFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, repoerrors.WrapUnknown("note_link.list_by_source", "note_link", errors.New("db"))
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, repoerrors.WrapUnknown("note_link.list_by_target", "note_link", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewConflict("note_link.update", "note_link")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("note_link.delete", "note_link")
			},
		},
		NoteAssetRepo: &fakeNoteAssetRepo{
			createFn: func(_ context.Context, _ model.NoteAsset) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewConflict("note_asset.create", "note_asset")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewNotFound("note_asset.get_by_id", "note_asset")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
				return nil, repoerrors.WrapUnknown("note_asset.list_by_note", "note_asset", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewConflict("note_asset.update", "note_asset")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("note_asset.delete", "note_asset")
			},
		},
		MapNotePlacementRepo: &fakeMapPlacementRepo{
			createFn: func(_ context.Context, _ model.MapNotePlacement) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewConflict("map_note_placement.create", "map_note_placement")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewNotFound("map_note_placement.get_by_id", "map_note_placement")
			},
			listByMapFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, repoerrors.WrapUnknown("map_note_placement.list_by_map", "map_note_placement", errors.New("db"))
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, repoerrors.WrapUnknown("map_note_placement.list_by_target", "map_note_placement", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewConflict("map_note_placement.update", "map_note_placement")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("map_note_placement.delete", "map_note_placement")
			},
		},
		Clock: fakeClock{now: time.Now().UTC()},
		Policy: fakeNotePolicy{validateFn: func(t, c string, nt constants.NoteType, m []byte) (string, string, []byte, error) {
			return t, c, m, nil
		}},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01X", nil }},
	})

	_, err = svcErr.Create(ctx, CreateNoteParams{Title: "x", NoteType: constants.General})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected note create conflict mapping")
	}
	_, err = svcErr.GetByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected note get not found mapping")
	}
	_, err = svcErr.List(ctx, ListNotesParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected note list unknown mapping")
	}
}

func TestNoteServiceBroadValidationAndMappingCoverage(t *testing.T) {
	ctx := context.Background()
	svc := newNoteServiceForTests()

	_, err := svc.GetOwner(ctx, "01N", constants.OwnerType(99), "01O", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected invalid owner type validation")
	}
	_, err = svc.ListOwnersByNote(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected note id validation")
	}
	_, err = svc.ListNotesByOwner(ctx, constants.OwnerType(99), "01O", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected owner type validation")
	}
	_, err = svc.ListTagsByNote(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected note id validation")
	}
	_, err = svc.ListNotesByTag(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected tag id validation")
	}
	_, err = svc.GetLinkByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.ListLinksBySource(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected source id validation")
	}
	_, err = svc.ListLinksByTarget(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected target id validation")
	}
	_, err = svc.UpdateLink(ctx, UpdateNoteLinkParams{ID: "01", LinkType: constants.NoteLinkType(99), ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected link type validation")
	}
	err = svc.DeleteLink(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.CreateAsset(ctx, CreateNoteAssetParams{NoteID: "", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected note id validation")
	}
	_, err = svc.CreateAsset(ctx, CreateNoteAssetParams{NoteID: "01N", AssetType: constants.AssetType(99), StoragePath: "/x", MIMEType: "image/png"})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected asset type validation")
	}
	_, err = svc.GetAssetByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.ListAssetsByNote(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected note id validation")
	}
	_, err = svc.UpdateAsset(ctx, UpdateNoteAssetParams{ID: "01A", AssetType: constants.Image, StoragePath: "", MIMEType: "image/png", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected storage path validation")
	}
	err = svc.DeleteAsset(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.UpsertMapPlacement(ctx, UpsertMapNotePlacementParams{MapNoteID: "", TargetNoteID: "01T", X: 1, Y: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected map id validation")
	}
	_, err = svc.UpsertMapPlacement(ctx, UpsertMapNotePlacementParams{MapNoteID: "01M", TargetNoteID: "01T", X: 101, Y: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected x range validation")
	}
	_, err = svc.GetMapPlacementByID(ctx, "", false)
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}
	_, err = svc.ListMapPlacementsByMap(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected map note id validation")
	}
	_, err = svc.ListMapPlacementsByTarget(ctx, "", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected target note id validation")
	}
	_, err = svc.UpdateMapPlacement(ctx, UpdateMapNotePlacementParams{ID: "01P", X: 101, Y: 0, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected range validation")
	}
	err = svc.DeleteMapPlacement(ctx, "")
	if !errors.Is(err, serviceerrors.ErrValidation) {
		t.Fatalf("expected id validation")
	}

	svcErr := NewNoteServiceWithDeps(NoteServiceDeps{
		Repo: &fakeNoteRepo{
			createFn: func(_ context.Context, _ model.Note) (model.Note, error) { return model.Note{}, nil },
			getFn:    func(_ context.Context, _ model.ULID, _ bool) (model.Note, error) { return model.Note{}, nil },
			listFn:   func(_ context.Context, _ interfaces.ListNotesParams) ([]model.Note, error) { return nil, nil },
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteParams) (model.Note, error) { return model.Note{}, nil },
			deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
		},
		NoteOwnerRepo: &fakeNoteOwnerRepo{
			createFn: func(_ context.Context, _ model.NoteOwner) (model.NoteOwner, error) {
				return model.NoteOwner{}, repoerrors.NewConflict("note_owner.create", "note_owner")
			},
			getFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID, _ bool) (model.NoteOwner, error) {
				return model.NoteOwner{}, repoerrors.NewNotFound("note_owner.get", "note_owner")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, repoerrors.WrapUnknown("note_owner.list_by_note", "note_owner", errors.New("db"))
			},
			listByOwnerFn: func(_ context.Context, _ constants.OwnerType, _ model.ULID, _ interfaces.ListNoteOwnersParams) ([]model.NoteOwner, error) {
				return nil, repoerrors.WrapUnknown("note_owner.list_by_owner", "note_owner", errors.New("db"))
			},
			deleteFn: func(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID) error {
				return repoerrors.NewConflict("note_owner.delete", "note_owner")
			},
		},
		NoteTagRepo: &fakeNoteTagRepo{
			createFn: func(_ context.Context, _ model.NoteTag) (model.NoteTag, error) {
				return model.NoteTag{}, repoerrors.NewConflict("note_tag.create", "note_tag")
			},
			getFn: func(_ context.Context, _, _ model.ULID, _ bool) (model.NoteTag, error) {
				return model.NoteTag{}, repoerrors.NewNotFound("note_tag.get", "note_tag")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, repoerrors.WrapUnknown("note_tag.list_by_note", "note_tag", errors.New("db"))
			},
			listByTagFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteTagsParams) ([]model.NoteTag, error) {
				return nil, repoerrors.WrapUnknown("note_tag.list_by_tag", "note_tag", errors.New("db"))
			},
			deleteFn: func(_ context.Context, _, _ model.ULID) error {
				return repoerrors.NewConflict("note_tag.delete", "note_tag")
			},
		},
		NoteLinkRepo: &fakeNoteLinkRepo{
			createFn: func(_ context.Context, _ model.NoteLink) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewConflict("note_link.create", "note_link")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewNotFound("note_link.get_by_id", "note_link")
			},
			listBySourceFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, repoerrors.WrapUnknown("note_link.list_by_source", "note_link", errors.New("db"))
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteLinksParams) ([]model.NoteLink, error) {
				return nil, repoerrors.WrapUnknown("note_link.list_by_target", "note_link", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteLinkParams) (model.NoteLink, error) {
				return model.NoteLink{}, repoerrors.NewConflict("note_link.update", "note_link")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("note_link.delete", "note_link")
			},
		},
		NoteAssetRepo: &fakeNoteAssetRepo{
			createFn: func(_ context.Context, _ model.NoteAsset) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewConflict("note_asset.create", "note_asset")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewNotFound("note_asset.get_by_id", "note_asset")
			},
			listByNoteFn: func(_ context.Context, _ model.ULID, _ interfaces.ListNoteAssetsParams) ([]model.NoteAsset, error) {
				return nil, repoerrors.WrapUnknown("note_asset.list_by_note", "note_asset", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateNoteAssetParams) (model.NoteAsset, error) {
				return model.NoteAsset{}, repoerrors.NewConflict("note_asset.update", "note_asset")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("note_asset.delete", "note_asset")
			},
		},
		MapNotePlacementRepo: &fakeMapPlacementRepo{
			createFn: func(_ context.Context, _ model.MapNotePlacement) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewConflict("map_note_placement.create", "map_note_placement")
			},
			getByIDFn: func(_ context.Context, _ model.ULID, _ bool) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewNotFound("map_note_placement.get_by_id", "map_note_placement")
			},
			listByMapFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, repoerrors.WrapUnknown("map_note_placement.list_by_map", "map_note_placement", errors.New("db"))
			},
			listByTargetFn: func(_ context.Context, _ model.ULID, _ interfaces.ListMapNotePlacementsParams) ([]model.MapNotePlacement, error) {
				return nil, repoerrors.WrapUnknown("map_note_placement.list_by_target", "map_note_placement", errors.New("db"))
			},
			updateFn: func(_ context.Context, _ interfaces.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
				return model.MapNotePlacement{}, repoerrors.NewConflict("map_note_placement.update", "map_note_placement")
			},
			deleteFn: func(_ context.Context, _ model.ULID) error {
				return repoerrors.NewNotFound("map_note_placement.delete", "map_note_placement")
			},
		},
		Clock: fakeClock{now: time.Now().UTC()},
		Policy: fakeNotePolicy{validateFn: func(t, c string, nt constants.NoteType, m []byte) (string, string, []byte, error) {
			return t, c, []byte("{}"), nil
		}},
		IDGenerator: fakeIDGenerator{newULIDFn: func() (model.ULID, error) { return "01X", nil }},
	})

	_, err = svcErr.AddOwner(ctx, AddNoteOwnerParams{NoteID: "01", OwnerID: "02", OwnerType: constants.World})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected add owner conflict")
	}
	err = svcErr.RemoveOwner(ctx, "01", constants.World, "02")
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected remove owner conflict")
	}
	_, err = svcErr.GetOwner(ctx, "01", constants.World, "02", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected get owner not found")
	}
	_, err = svcErr.ListOwnersByNote(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list owners unknown")
	}
	_, err = svcErr.ListNotesByOwner(ctx, constants.World, "02", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list by owner unknown")
	}
	_, err = svcErr.AddTag(ctx, AddNoteTagParams{NoteID: "01", TagID: "02"})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected add tag conflict")
	}
	err = svcErr.RemoveTag(ctx, "01", "02")
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected remove tag conflict")
	}
	_, err = svcErr.GetTag(ctx, "01", "02", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected get tag not found")
	}
	_, err = svcErr.ListTagsByNote(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list tags unknown")
	}
	_, err = svcErr.ListNotesByTag(ctx, "02", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list notes by tag unknown")
	}
	_, err = svcErr.CreateLink(ctx, CreateNoteLinkParams{SourceNoteID: "01", TargetNoteID: "02", LinkType: constants.Related})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected create link conflict")
	}
	_, err = svcErr.GetLinkByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected get link not found")
	}
	_, err = svcErr.ListLinksBySource(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list links source unknown")
	}
	_, err = svcErr.ListLinksByTarget(ctx, "02", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list links target unknown")
	}
	_, err = svcErr.UpdateLink(ctx, UpdateNoteLinkParams{ID: "01", LinkType: constants.Related, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected update link conflict")
	}
	err = svcErr.DeleteLink(ctx, "01")
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected delete link not found")
	}
	_, err = svcErr.CreateAsset(ctx, CreateNoteAssetParams{NoteID: "01", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png"})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected create asset conflict")
	}
	_, err = svcErr.GetAssetByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected get asset not found")
	}
	_, err = svcErr.ListAssetsByNote(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list assets unknown")
	}
	_, err = svcErr.UpdateAsset(ctx, UpdateNoteAssetParams{ID: "01", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png", ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected update asset conflict")
	}
	err = svcErr.DeleteAsset(ctx, "01")
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected delete asset not found")
	}
	_, err = svcErr.UpsertMapPlacement(ctx, UpsertMapNotePlacementParams{MapNoteID: "01", TargetNoteID: "02", X: 1, Y: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected create placement conflict")
	}
	_, err = svcErr.GetMapPlacementByID(ctx, "01", false)
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected get placement not found")
	}
	_, err = svcErr.ListMapPlacementsByMap(ctx, "01", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list placement by map unknown")
	}
	_, err = svcErr.ListMapPlacementsByTarget(ctx, "02", RelationListParams{Offset: 0, Limit: 10})
	if !errors.Is(err, serviceerrors.ErrUnknown) {
		t.Fatalf("expected list placement by target unknown")
	}
	_, err = svcErr.UpdateMapPlacement(ctx, UpdateMapNotePlacementParams{ID: "01", X: 1, Y: 1, ExpectedVersion: 1})
	if !errors.Is(err, serviceerrors.ErrConflict) {
		t.Fatalf("expected update placement conflict")
	}
	err = svcErr.DeleteMapPlacement(ctx, "01")
	if !errors.Is(err, serviceerrors.ErrNotFound) {
		t.Fatalf("expected delete placement not found")
	}
}
