package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakePlayerService struct {
	createFn  func(context.Context, service.CreatePlayerParams) (model.Player, error)
	getFn     func(context.Context, model.ULID, bool) (model.Player, error)
	listFn    func(context.Context, service.ListPlayersParams) ([]service.PlayerListItem, error)
	searchFn  func(context.Context, service.SearchPlayersParams) ([]service.PlayerListItem, error)
	updateFn  func(context.Context, service.UpdatePlayerParams) (model.Player, error)
	deleteFn  func(context.Context, model.ULID) error
	restoreFn func(context.Context, model.ULID) (model.Player, error)
}

func (f *fakePlayerService) Create(ctx context.Context, params service.CreatePlayerParams) (model.Player, error) {
	return f.createFn(ctx, params)
}
func (f *fakePlayerService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error) {
	return f.getFn(ctx, id, includeDeleted)
}
func (f *fakePlayerService) List(ctx context.Context, params service.ListPlayersParams) ([]service.PlayerListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakePlayerService) SearchByName(ctx context.Context, params service.SearchPlayersParams) ([]service.PlayerListItem, error) {
	return f.searchFn(ctx, params)
}
func (f *fakePlayerService) Update(ctx context.Context, params service.UpdatePlayerParams) (model.Player, error) {
	return f.updateFn(ctx, params)
}
func (f *fakePlayerService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}
func (f *fakePlayerService) Restore(ctx context.Context, id model.ULID) (model.Player, error) {
	return f.restoreFn(ctx, id)
}

func TestPlayerAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := &fakePlayerService{
		createFn: func(_ context.Context, params service.CreatePlayerParams) (model.Player, error) {
			if params.Name == "bad" {
				return model.Player{}, serviceerrors.WrapValidation("x", "player", errors.New("bad input"))
			}
			return model.Player{
				ID:   "01H",
				Name: params.Name,
				AuditFields: model.AuditFields{
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
			}, nil
		},
		getFn: func(_ context.Context, id model.ULID, _ bool) (model.Player, error) {
			if id == "missing" {
				return model.Player{}, serviceerrors.NewNotFound("x", "player")
			}
			return model.Player{ID: id, Name: "Name", AuditFields: model.AuditFields{Version: 1}}, nil
		},
		listFn: func(_ context.Context, params service.ListPlayersParams) ([]service.PlayerListItem, error) {
			if params.Offset < 0 {
				return nil, serviceerrors.WrapValidation("x", "player", errors.New("offset"))
			}
			return []service.PlayerListItem{{ID: "1", Name: "A", Version: 1}}, nil
		},
		searchFn: func(_ context.Context, params service.SearchPlayersParams) ([]service.PlayerListItem, error) {
			if params.Query == "none" {
				return nil, serviceerrors.NewNotFound("x", "player")
			}
			return []service.PlayerListItem{{ID: "1", Name: "A", Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdatePlayerParams) (model.Player, error) {
			if params.ID == "conflict" {
				return model.Player{}, serviceerrors.NewConflict("x", "player")
			}
			return model.Player{ID: params.ID, Name: params.Name, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, id model.ULID) error {
			if id == "missing" {
				return serviceerrors.NewNotFound("x", "player")
			}
			return nil
		},
		restoreFn: func(_ context.Context, id model.ULID) (model.Player, error) {
			if id == "boom" {
				return model.Player{}, serviceerrors.WrapUnknown("x", "player", errors.New("sql failed"))
			}
			return model.Player{ID: id, Name: "R"}, nil
		},
	}

	api := NewPlayerAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBufferString(`{"name":"Ana"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("create null body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBufferString(`null`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
		assertNoTechLeak(t, rec.Body.Bytes())
	})

	t.Run("create validation mapping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBufferString(`{"name":"bad"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
		assertMessage(t, rec.Body.Bytes(), "invalid request")
	})

	t.Run("get not found mapping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/players/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
		assertMessage(t, rec.Body.Bytes(), "resource not found")
	})

	t.Run("list invalid offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/players?offset=x", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("search query required", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/players/search", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("update conflict mapping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/players/conflict", bytes.NewBufferString(`{"name":"A"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
		assertMessage(t, rec.Body.Bytes(), "request conflict")
	})

	t.Run("delete no content", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/players/01H", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("restore unknown mapping no leak", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/players/boom/restore", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		assertMessage(t, rec.Body.Bytes(), "internal server error")
		assertNoTechLeak(t, rec.Body.Bytes())
	})
}

func assertMessage(t *testing.T, body []byte, expected string) {
	t.Helper()
	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if payload["message"] != expected {
		t.Fatalf("expected message %q, got %q", expected, payload["message"])
	}
}

func assertNoTechLeak(t *testing.T, body []byte) {
	t.Helper()
	var payload map[string]string
	_ = json.Unmarshal(body, &payload)
	text := payload["message"]
	if text == "" {
		text = string(body)
	}
	if bytes.Contains(bytes.ToLower([]byte(text)), []byte("sql")) ||
		bytes.Contains(bytes.ToLower([]byte(text)), []byte("postgres")) ||
		bytes.Contains(bytes.ToLower([]byte(text)), []byte("service:")) {
		t.Fatalf("error message leaked implementation detail: %q", text)
	}
}
