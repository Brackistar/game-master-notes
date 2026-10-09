package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakeNoteService struct {
	now time.Time
}

func (f *fakeNoteService) Create(_ context.Context, params service.CreateNoteParams) (model.Note, error) {
	return model.Note{ID: "01N", Title: params.Title, ContentMD: params.ContentMD, NoteType: params.NoteType, MetadataJSON: params.MetadataJSON, AuditFields: model.AuditFields{CreatedAt: f.now, UpdatedAt: f.now, Version: 1}}, nil
}
func (f *fakeNoteService) GetByID(_ context.Context, id model.ULID, _ bool) (model.Note, error) {
	if id == "missing" {
		return model.Note{}, serviceerrors.NewNotFound("x", "note")
	}
	return model.Note{ID: id, Title: "N", NoteType: constants.General, AuditFields: model.AuditFields{Version: 2}}, nil
}
func (f *fakeNoteService) List(_ context.Context, _ service.ListNotesParams) ([]service.NoteListItem, error) {
	return []service.NoteListItem{{ID: "01N", Title: "N", NoteType: constants.General, Version: 1}}, nil
}
func (f *fakeNoteService) Update(_ context.Context, params service.UpdateNoteParams) (model.Note, error) {
	if params.ID == "conflict" {
		return model.Note{}, serviceerrors.NewConflict("x", "note")
	}
	return model.Note{ID: params.ID, Title: params.Title, NoteType: params.NoteType, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
}
func (f *fakeNoteService) Delete(_ context.Context, _ model.ULID) error { return nil }

func (f *fakeNoteService) AddOwner(_ context.Context, params service.AddNoteOwnerParams) (model.NoteOwner, error) {
	return model.NoteOwner{NoteID: params.NoteID, OwnerType: params.OwnerType, OwnerID: params.OwnerID, CreatedAt: f.now, UpdatedAt: f.now}, nil
}
func (f *fakeNoteService) RemoveOwner(_ context.Context, _ model.ULID, _ constants.OwnerType, _ model.ULID) error {
	return nil
}
func (f *fakeNoteService) GetOwner(_ context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, _ bool) (model.NoteOwner, error) {
	return model.NoteOwner{NoteID: noteID, OwnerType: ownerType, OwnerID: ownerID, CreatedAt: f.now, UpdatedAt: f.now}, nil
}
func (f *fakeNoteService) ListOwnersByNote(_ context.Context, noteID model.ULID, _ service.RelationListParams) ([]model.NoteOwner, error) {
	return []model.NoteOwner{{NoteID: noteID, OwnerType: constants.Player, OwnerID: "01P", CreatedAt: f.now, UpdatedAt: f.now}}, nil
}
func (f *fakeNoteService) ListNotesByOwner(_ context.Context, ownerType constants.OwnerType, ownerID model.ULID, _ service.RelationListParams) ([]model.NoteOwner, error) {
	return []model.NoteOwner{{NoteID: "01N", OwnerType: ownerType, OwnerID: ownerID, CreatedAt: f.now, UpdatedAt: f.now}}, nil
}

func (f *fakeNoteService) AddTag(_ context.Context, params service.AddNoteTagParams) (model.NoteTag, error) {
	return model.NoteTag{NoteID: params.NoteID, TagID: params.TagID, CreatedAt: f.now, UpdatedAt: f.now}, nil
}
func (f *fakeNoteService) RemoveTag(_ context.Context, _ model.ULID, _ model.ULID) error { return nil }
func (f *fakeNoteService) GetTag(_ context.Context, noteID, tagID model.ULID, _ bool) (model.NoteTag, error) {
	return model.NoteTag{NoteID: noteID, TagID: tagID, CreatedAt: f.now, UpdatedAt: f.now}, nil
}
func (f *fakeNoteService) ListTagsByNote(_ context.Context, noteID model.ULID, _ service.RelationListParams) ([]model.NoteTag, error) {
	return []model.NoteTag{{NoteID: noteID, TagID: "01T", CreatedAt: f.now, UpdatedAt: f.now}}, nil
}
func (f *fakeNoteService) ListNotesByTag(_ context.Context, tagID model.ULID, _ service.RelationListParams) ([]model.NoteTag, error) {
	return []model.NoteTag{{NoteID: "01N", TagID: tagID, CreatedAt: f.now, UpdatedAt: f.now}}, nil
}

func (f *fakeNoteService) CreateLink(_ context.Context, params service.CreateNoteLinkParams) (model.NoteLink, error) {
	return model.NoteLink{ID: "01L", SourceNoteID: params.SourceNoteID, TargetNoteID: params.TargetNoteID, LinkType: params.LinkType, AuditFields: model.AuditFields{CreatedAt: f.now, UpdatedAt: f.now, Version: 1}}, nil
}
func (f *fakeNoteService) GetLinkByID(_ context.Context, id model.ULID, _ bool) (model.NoteLink, error) {
	return model.NoteLink{ID: id, SourceNoteID: "01N", TargetNoteID: "01N2", LinkType: constants.Related, AuditFields: model.AuditFields{Version: 2}}, nil
}
func (f *fakeNoteService) ListLinksBySource(_ context.Context, sourceNoteID model.ULID, _ service.RelationListParams) ([]model.NoteLink, error) {
	return []model.NoteLink{{ID: "01L", SourceNoteID: sourceNoteID, TargetNoteID: "01N2", LinkType: constants.Related, AuditFields: model.AuditFields{Version: 1}}}, nil
}
func (f *fakeNoteService) ListLinksByTarget(_ context.Context, targetNoteID model.ULID, _ service.RelationListParams) ([]model.NoteLink, error) {
	return []model.NoteLink{{ID: "01L", SourceNoteID: "01N", TargetNoteID: targetNoteID, LinkType: constants.Related, AuditFields: model.AuditFields{Version: 1}}}, nil
}
func (f *fakeNoteService) UpdateLink(_ context.Context, params service.UpdateNoteLinkParams) (model.NoteLink, error) {
	if params.ID == "conflict" {
		return model.NoteLink{}, serviceerrors.NewConflict("x", "note_link")
	}
	return model.NoteLink{ID: params.ID, SourceNoteID: "01N", TargetNoteID: "01N2", LinkType: params.LinkType, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
}
func (f *fakeNoteService) DeleteLink(_ context.Context, _ model.ULID) error { return nil }

func (f *fakeNoteService) CreateAsset(_ context.Context, params service.CreateNoteAssetParams) (model.NoteAsset, error) {
	return model.NoteAsset{ID: "01A", NoteID: params.NoteID, AssetType: params.AssetType, StoragePath: params.StoragePath, MIMEType: params.MIMEType, AuditFields: model.AuditFields{CreatedAt: f.now, UpdatedAt: f.now, Version: 1}}, nil
}
func (f *fakeNoteService) GetAssetByID(_ context.Context, id model.ULID, _ bool) (model.NoteAsset, error) {
	return model.NoteAsset{ID: id, NoteID: "01N", AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png", AuditFields: model.AuditFields{Version: 2}}, nil
}
func (f *fakeNoteService) ListAssetsByNote(_ context.Context, noteID model.ULID, _ service.RelationListParams) ([]model.NoteAsset, error) {
	return []model.NoteAsset{{ID: "01A", NoteID: noteID, AssetType: constants.Image, StoragePath: "/x", MIMEType: "image/png", AuditFields: model.AuditFields{Version: 1}}}, nil
}
func (f *fakeNoteService) UpdateAsset(_ context.Context, params service.UpdateNoteAssetParams) (model.NoteAsset, error) {
	return model.NoteAsset{ID: params.ID, NoteID: "01N", AssetType: params.AssetType, StoragePath: params.StoragePath, MIMEType: params.MIMEType, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
}
func (f *fakeNoteService) DeleteAsset(_ context.Context, _ model.ULID) error { return nil }

func (f *fakeNoteService) UpsertMapPlacement(_ context.Context, params service.UpsertMapNotePlacementParams) (model.MapNotePlacement, error) {
	return model.MapNotePlacement{ID: "01M", MapNoteID: params.MapNoteID, TargetNoteID: params.TargetNoteID, X: params.X, Y: params.Y, AuditFields: model.AuditFields{CreatedAt: f.now, UpdatedAt: f.now, Version: 1}}, nil
}
func (f *fakeNoteService) GetMapPlacementByID(_ context.Context, id model.ULID, _ bool) (model.MapNotePlacement, error) {
	return model.MapNotePlacement{ID: id, MapNoteID: "01N", TargetNoteID: "01N2", X: 1, Y: 2, AuditFields: model.AuditFields{Version: 2}}, nil
}
func (f *fakeNoteService) ListMapPlacementsByMap(_ context.Context, mapNoteID model.ULID, _ service.RelationListParams) ([]model.MapNotePlacement, error) {
	return []model.MapNotePlacement{{ID: "01M", MapNoteID: mapNoteID, TargetNoteID: "01N2", X: 1, Y: 2, AuditFields: model.AuditFields{Version: 1}}}, nil
}
func (f *fakeNoteService) ListMapPlacementsByTarget(_ context.Context, targetNoteID model.ULID, _ service.RelationListParams) ([]model.MapNotePlacement, error) {
	return []model.MapNotePlacement{{ID: "01M", MapNoteID: "01N", TargetNoteID: targetNoteID, X: 1, Y: 2, AuditFields: model.AuditFields{Version: 1}}}, nil
}
func (f *fakeNoteService) UpdateMapPlacement(_ context.Context, params service.UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
	return model.MapNotePlacement{ID: params.ID, MapNoteID: "01N", TargetNoteID: "01N2", X: params.X, Y: params.Y, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
}
func (f *fakeNoteService) DeleteMapPlacement(_ context.Context, _ model.ULID) error { return nil }

func TestNoteAPIEndpoints(t *testing.T) {
	svc := &fakeNoteService{now: time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)}
	api := NewNoteAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(`{"title":"N","content_md":"C","note_type":"general"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("create invalid note type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBufferString(`{"title":"N","content_md":"C","note_type":"invalid"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("owner endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes/01N/owners", bytes.NewBufferString(`{"owner_type":"player","owner_id":"01P"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/notes/01N/owners", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("tag endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes/01N/tags/01T", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/notes/01N/tags", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("link endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/note-links", bytes.NewBufferString(`{"source_note_id":"01N","target_note_id":"01N2","link_type":"related"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodPatch, "/note-links/conflict", bytes.NewBufferString(`{"link_type":"related"}`))
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("asset invalid type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes/01N/assets", bytes.NewBufferString(`{"asset_type":"video","storage_path":"/x","mime_type":"video/mp4"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("map placement endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/map-note-placements", bytes.NewBufferString(`{"map_note_id":"01N","target_note_id":"01N2","x":10,"y":20}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/notes/01N/map-placements/as-map", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}

func TestNoteAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewNoteAPI(nil)
}
