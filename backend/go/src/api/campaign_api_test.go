package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	helpers "github.com/Brackistar/game-master-notes/backend/go/src/api/shared"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

type fakeCampaignService struct {
	createFn                 func(context.Context, service.CreateCampaignParams) (model.Campaign, error)
	getByIDFn                func(context.Context, model.ULID, bool) (model.Campaign, error)
	listFn                   func(context.Context, service.ListCampaignsParams) ([]service.CampaignListItem, error)
	updateFn                 func(context.Context, service.UpdateCampaignParams) (model.Campaign, error)
	deleteFn                 func(context.Context, model.ULID) error
	addPlayerFn              func(context.Context, model.ULID, model.ULID) (model.CampaignPlayer, error)
	removePlayerFn           func(context.Context, model.ULID, model.ULID) error
	getPlayerRelationFn      func(context.Context, model.ULID, model.ULID, bool) (model.CampaignPlayer, error)
	listPlayersFn            func(context.Context, model.ULID, service.ListCampaignsParams) ([]model.CampaignPlayer, error)
	listCampaignsForPlayerFn func(context.Context, model.ULID, service.ListCampaignsParams) ([]model.CampaignPlayer, error)
}

func (f *fakeCampaignService) Create(ctx context.Context, params service.CreateCampaignParams) (model.Campaign, error) {
	return f.createFn(ctx, params)
}
func (f *fakeCampaignService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error) {
	return f.getByIDFn(ctx, id, includeDeleted)
}
func (f *fakeCampaignService) List(ctx context.Context, params service.ListCampaignsParams) ([]service.CampaignListItem, error) {
	return f.listFn(ctx, params)
}
func (f *fakeCampaignService) Update(ctx context.Context, params service.UpdateCampaignParams) (model.Campaign, error) {
	return f.updateFn(ctx, params)
}
func (f *fakeCampaignService) Delete(ctx context.Context, id model.ULID) error {
	return f.deleteFn(ctx, id)
}
func (f *fakeCampaignService) AddPlayer(ctx context.Context, campaignID, playerID model.ULID) (model.CampaignPlayer, error) {
	return f.addPlayerFn(ctx, campaignID, playerID)
}
func (f *fakeCampaignService) RemovePlayer(ctx context.Context, campaignID, playerID model.ULID) error {
	return f.removePlayerFn(ctx, campaignID, playerID)
}
func (f *fakeCampaignService) GetPlayerRelation(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error) {
	return f.getPlayerRelationFn(ctx, campaignID, playerID, includeDeleted)
}
func (f *fakeCampaignService) ListPlayers(ctx context.Context, campaignID model.ULID, params service.ListCampaignsParams) ([]model.CampaignPlayer, error) {
	return f.listPlayersFn(ctx, campaignID, params)
}
func (f *fakeCampaignService) ListCampaignsForPlayer(ctx context.Context, playerID model.ULID, params service.ListCampaignsParams) ([]model.CampaignPlayer, error) {
	return f.listCampaignsForPlayerFn(ctx, playerID, params)
}

func TestCampaignAPIEndpoints(t *testing.T) {
	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	svc := &fakeCampaignService{
		createFn: func(_ context.Context, params service.CreateCampaignParams) (model.Campaign, error) {
			if params.Name == "bad" {
				return model.Campaign{}, serviceerrors.WrapValidation("x", "campaign", errors.New("bad"))
			}
			return model.Campaign{
				ID:        "01C",
				WorldID:   params.WorldID,
				Name:      params.Name,
				StartDate: params.StartDate,
				EndDate:   params.EndDate,
				AuditFields: model.AuditFields{
					CreatedAt: now,
					UpdatedAt: now,
					Version:   1,
				},
			}, nil
		},
		getByIDFn: func(_ context.Context, id model.ULID, _ bool) (model.Campaign, error) {
			if id == "missing" {
				return model.Campaign{}, serviceerrors.NewNotFound("x", "campaign")
			}
			return model.Campaign{ID: id, WorldID: "01W", Name: "Camp", AuditFields: model.AuditFields{Version: 4}}, nil
		},
		listFn: func(_ context.Context, params service.ListCampaignsParams) ([]service.CampaignListItem, error) {
			return []service.CampaignListItem{{ID: "01C", WorldID: "01W", Name: "Camp", StartDate: &start, EndDate: &end, Version: 1}}, nil
		},
		updateFn: func(_ context.Context, params service.UpdateCampaignParams) (model.Campaign, error) {
			if params.ID == "conflict" {
				return model.Campaign{}, serviceerrors.NewConflict("x", "campaign")
			}
			return model.Campaign{ID: params.ID, WorldID: "01W", Name: params.Name, StartDate: params.StartDate, EndDate: params.EndDate, AuditFields: model.AuditFields{Version: params.ExpectedVersion + 1}}, nil
		},
		deleteFn: func(_ context.Context, id model.ULID) error {
			if id == "missing" {
				return serviceerrors.NewNotFound("x", "campaign")
			}
			return nil
		},
		addPlayerFn: func(_ context.Context, campaignID, playerID model.ULID) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{CampaignID: campaignID, PlayerID: playerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		removePlayerFn: func(_ context.Context, campaignID, playerID model.ULID) error {
			if playerID == "missing" {
				return serviceerrors.NewNotFound("x", "campaign_player")
			}
			return nil
		},
		getPlayerRelationFn: func(_ context.Context, campaignID, playerID model.ULID, _ bool) (model.CampaignPlayer, error) {
			return model.CampaignPlayer{CampaignID: campaignID, PlayerID: playerID, CreatedAt: now, UpdatedAt: now}, nil
		},
		listPlayersFn: func(_ context.Context, campaignID model.ULID, _ service.ListCampaignsParams) ([]model.CampaignPlayer, error) {
			return []model.CampaignPlayer{{CampaignID: campaignID, PlayerID: "01P", CreatedAt: now, UpdatedAt: now}}, nil
		},
		listCampaignsForPlayerFn: func(_ context.Context, playerID model.ULID, _ service.ListCampaignsParams) ([]model.CampaignPlayer, error) {
			return []model.CampaignPlayer{{CampaignID: "01C", PlayerID: playerID, CreatedAt: now, UpdatedAt: now}}, nil
		},
	}

	api := NewCampaignAPI(svc)
	mux := http.NewServeMux()
	api.Register(mux)

	t.Run("create success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewBufferString(`{"world_id":"01W","name":"Camp","start_date":"2026-01-10","end_date":"2026-01-20"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("create invalid date", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewBufferString(`{"world_id":"01W","name":"Camp","start_date":"bad"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/campaigns/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
		assertMessage(t, rec.Body.Bytes(), "resource not found")
	})

	t.Run("list success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("update conflict", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/campaigns/conflict", bytes.NewBufferString(`{"name":"New"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/campaigns/01C", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("campaign player add and list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/campaigns/01C/players/01P", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/campaigns/01C/players", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("campaign player remove not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/campaigns/01C/players/missing", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("campaign-player relation and list campaigns by player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/campaigns/01C/players/01P", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/players/01P/campaigns", nil)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}

func TestCampaignAPINilServicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = NewCampaignAPI(nil)
}

func TestCampaignDateHelpers(t *testing.T) {
	date := "2026-04-01"
	parsed, err := helpers.ParseDatePointer(&date, "start_date")
	if err != nil || parsed == nil {
		t.Fatalf("expected parse success")
	}
	if out := helpers.FormatDatePointer(parsed); out == nil || *out != date {
		t.Fatalf("expected round-trip date")
	}
	blank := " "
	nilDate, err := helpers.ParseDatePointer(&blank, "end_date")
	if err != nil || nilDate != nil {
		t.Fatalf("expected blank date to map nil")
	}
	invalid := "x"
	_, err = helpers.ParseDatePointer(&invalid, "end_date")
	if err == nil {
		t.Fatalf("expected invalid date error")
	}
	if out := helpers.FormatDatePointer(nil); out != nil {
		t.Fatalf("expected nil formatter")
	}
}
