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

type fakePlaneService struct {
	createFn func(context.Context, service.CreatePlaneParams) (model.Plane, error)
	getFn    func(context.Context, model.ULID, bool) (model.Plane, error)
	listFn   func(context.Context, service.ListPlanesParams) ([]service.PlaneListItem, error)
	updateFn func(context.Context, service.UpdatePlaneParams) (model.Plane, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakePlaneService) Create(ctx context.Context, params service.CreatePlaneParams) (model.Plane, error) {
	return f.createFn(ctx, params)
}
func (f *fakePlaneService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakePlaneService) List(ctx context.Context, params service.ListPlanesParams) ([]service.PlaneListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakePlaneService) Update(ctx context.Context, params service.UpdatePlaneParams) (model.Plane, error) {
	return f.updateFn(ctx, params)
}
func (f *fakePlaneService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

func TestPlaneAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	svc := &fakePlaneService{
		createFn: func(_ context.Context, params service.CreatePlaneParams) (model.Plane, error) {
			return model.Plane{ID: "01PL", WorldID: params.WorldID, Name: params.Name, AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}}, nil
		},
		getFn: func(_ context.Context, id model.ULID, _ bool) (model.Plane, error) {
			if id == "missing" {
				return model.Plane{}, serviceerrors.NewNotFound("x", "plane")
			}
			return model.Plane{ID: id, WorldID: "01W", Name: "P", AuditFields: model.AuditFields{Version: 2}}, nil
		},
		listFn: func(_ context.Context, _ service.ListPlanesParams) ([]service.PlaneListItem, error) {
			return []service.PlaneListItem{{ID: "01PL", WorldID: "01W", Name: "P", Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdatePlaneParams) (model.Plane, error) {
			if params.ID == "conflict" {
				return model.Plane{}, serviceerrors.NewConflict("x", "plane")
			}
			return model.Plane{ID: params.ID, WorldID: "01W", Name: params.Name, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}

	api := NewPlaneAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/planes", bytes.NewBufferString(`{"world_id":"01W","name":"P","description":"D"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/planes/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("list success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/planes", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("update conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/planes/conflict", bytes.NewBufferString(`{"name":"N","description":"D"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/planes/01PL", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestPlaneAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewPlaneAPI(nil)
}
