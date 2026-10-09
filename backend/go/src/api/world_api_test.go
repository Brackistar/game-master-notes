package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakeWorldService struct {
	createFn func(context.Context, service.CreateWorldParams) (model.World, error)
	getFn    func(context.Context, model.ULID, bool) (model.World, error)
	listFn   func(context.Context, service.ListWorldsParams) ([]service.WorldListItem, error)
	updateFn func(context.Context, service.UpdateWorldParams) (model.World, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeWorldService) Create(ctx context.Context, params service.CreateWorldParams) (model.World, error) {
	return f.createFn(ctx, params)
}
func (f *fakeWorldService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeWorldService) List(ctx context.Context, params service.ListWorldsParams) ([]service.WorldListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakeWorldService) Update(ctx context.Context, params service.UpdateWorldParams) (model.World, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeWorldService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

func TestWorldAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	svc := &fakeWorldService{
		createFn: func(_ context.Context, params service.CreateWorldParams) (model.World, error) {
			return model.World{ID: "01W", PlaneID: params.PlaneID, Name: params.Name, Description: params.Description, Status: params.Status, AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}}, nil
		},
		getFn: func(_ context.Context, id model.ULID, _ bool) (model.World, error) {
			if id == "missing" {
				return model.World{}, serviceerrors.NewNotFound("x", "world")
			}
			return model.World{ID: id, PlaneID: "01P", Name: "W", Status: constants.Active, AuditFields: model.AuditFields{Version: 3}}, nil
		},
		listFn: func(_ context.Context, _ service.ListWorldsParams) ([]service.WorldListItem, error) {
			return []service.WorldListItem{{ID: "01W", PlaneID: "01P", Name: "W", Status: constants.Draft, Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdateWorldParams) (model.World, error) {
			if params.ID == "conflict" {
				return model.World{}, serviceerrors.NewConflict("x", "world")
			}
			return model.World{ID: params.ID, PlaneID: "01P", Name: params.Name, Status: params.Status, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, id model.ULID) error {
			if id == "boom" {
				return serviceerrors.WrapUnknown("x", "world", errors.New("sql exploded"))
			}
			return nil
		},
	}

	api := NewWorldAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/worlds", bytes.NewBufferString(`{"plane_id":"01P","name":"A","description":"D","status":"active"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("create invalid status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/worlds", bytes.NewBufferString(`{"plane_id":"01P","name":"A","description":"D","status":"bad"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/worlds/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("list invalid query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/worlds?offset=x", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("update conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/worlds/conflict", bytes.NewBufferString(`{"name":"B","description":"D","status":"active"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("delete unknown no leak", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/worlds/boom", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		assertNoTechLeak(t, rec.Body.Bytes())
	})
}

func TestWorldAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewWorldAPI(nil)
}
