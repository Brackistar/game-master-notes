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

type fakeSessionService struct {
	createFn func(context.Context, service.CreateSessionParams) (model.Session, error)
	getFn    func(context.Context, model.ULID, bool) (model.Session, error)
	listFn   func(context.Context, service.ListSessionsParams) ([]service.SessionListItem, error)
	updateFn func(context.Context, service.UpdateSessionParams) (model.Session, error)
	deleteFn func(context.Context, model.ULID) error
}

func (f *fakeSessionService) Create(ctx context.Context, params service.CreateSessionParams) (model.Session, error) {
	return f.createFn(ctx, params)
}
func (f *fakeSessionService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakeSessionService) List(ctx context.Context, params service.ListSessionsParams) ([]service.SessionListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakeSessionService) Update(ctx context.Context, params service.UpdateSessionParams) (model.Session, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeSessionService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}

func TestSessionAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	playedOn := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	svc := &fakeSessionService{
		createFn: func(_ context.Context, params service.CreateSessionParams) (model.Session, error) {
			return model.Session{ID: "01S", CampaignID: params.CampaignID, PlayedOn: params.PlayedOn, SummaryMD: params.SummaryMD, AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}}, nil
		},
		getFn: func(_ context.Context, id model.ULID, _ bool) (model.Session, error) {
			if id == "missing" {
				return model.Session{}, serviceerrors.NewNotFound("x", "session")
			}
			return model.Session{ID: id, CampaignID: "01C", PlayedOn: &playedOn, SummaryMD: "S", AuditFields: model.AuditFields{Version: 2}}, nil
		},
		listFn: func(_ context.Context, _ service.ListSessionsParams) ([]service.SessionListItem, error) {
			return []service.SessionListItem{{ID: "01S", CampaignID: "01C", PlayedOn: &playedOn, SummaryMD: "S", Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdateSessionParams) (model.Session, error) {
			if params.ID == "conflict" {
				return model.Session{}, serviceerrors.NewConflict("x", "session")
			}
			return model.Session{ID: params.ID, CampaignID: "01C", PlayedOn: params.PlayedOn, SummaryMD: params.SummaryMD, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, _ model.ULID) error { return nil },
	}

	api := NewSessionAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{"campaign_id":"01C","played_on":"2026-02-02","summary_md":"S"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("create invalid date", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{"campaign_id":"01C","played_on":"bad","summary_md":"S"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sessions/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("update conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/sessions/conflict", bytes.NewBufferString(`{"played_on":"2026-02-03","summary_md":"U"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/sessions/01S", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestSessionAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewSessionAPI(nil)
}
