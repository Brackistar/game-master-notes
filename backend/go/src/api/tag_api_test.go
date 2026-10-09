package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakeTagService struct {
	createFn func(context.Context, service.CreateTagParams) (model.Tag, error)
	getFn    func(context.Context, model.ULID, bool) (model.Tag, error)
	listFn   func(context.Context, service.ListTagsParams) ([]service.TagListItem, error)
	updateFn func(context.Context, service.UpdateTagParams) (model.Tag, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeTagService) Create(ctx context.Context, params service.CreateTagParams) (model.Tag, error) {
	return f.createFn(ctx, params)
}
func (f *fakeTagService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeTagService) List(ctx context.Context, params service.ListTagsParams) ([]service.TagListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakeTagService) Update(ctx context.Context, params service.UpdateTagParams) (model.Tag, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeTagService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

func TestTagAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	campaignID := model.ULID("01C")
	svc := &fakeTagService{
		createFn: func(_ context.Context, params service.CreateTagParams) (model.Tag, error) {
			return model.Tag{ID: "01T", Name: params.Name, CampaignID: params.CampaignID, AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}}, nil
		},
		getFn: func(_ context.Context, id model.ULID, _ bool) (model.Tag, error) {
			if id == "missing" {
				return model.Tag{}, serviceerrors.NewNotFound("x", "tag")
			}
			return model.Tag{ID: id, Name: "Tag", CampaignID: &campaignID, AuditFields: model.AuditFields{Version: 2}}, nil
		},
		listFn: func(_ context.Context, _ service.ListTagsParams) ([]service.TagListItem, error) {
			return []service.TagListItem{{ID: "01T", Name: "Tag", CampaignID: &campaignID, Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdateTagParams) (model.Tag, error) {
			if params.ID == "conflict" {
				return model.Tag{}, serviceerrors.NewConflict("x", "tag")
			}
			return model.Tag{ID: params.ID, Name: params.Name, CampaignID: params.CampaignID, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}

	api := NewTagAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewBufferString(`{"name":"Tag","campaign_id":"01C"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tags/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("list success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tags", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("update conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/tags/conflict", bytes.NewBufferString(`{"name":"Tag2"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/tags/01T", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestTagAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewTagAPI(nil)
}
