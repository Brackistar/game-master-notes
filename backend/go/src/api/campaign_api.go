package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/api/shared"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
)

type CampaignService interface {
	Create(ctx context.Context, params service.CreateCampaignParams) (model.Campaign, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error)
	List(ctx context.Context, params service.ListCampaignsParams) ([]service.CampaignListItem, error)
	Update(ctx context.Context, params service.UpdateCampaignParams) (model.Campaign, error)
	Delete(ctx context.Context, id model.ULID) error

	AddPlayer(ctx context.Context, campaignID, playerID model.ULID) (model.CampaignPlayer, error)
	RemovePlayer(ctx context.Context, campaignID, playerID model.ULID) error
	GetPlayerRelation(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error)
	ListPlayers(ctx context.Context, campaignID model.ULID, params service.ListCampaignsParams) ([]model.CampaignPlayer, error)
	ListCampaignsForPlayer(ctx context.Context, playerID model.ULID, params service.ListCampaignsParams) ([]model.CampaignPlayer, error)
}

type CampaignAPI struct {
	service CampaignService
}

func NewCampaignAPI(service CampaignService) *CampaignAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "campaign", "service"))
	}
	return &CampaignAPI{service: service}
}

func (a *CampaignAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /campaigns", a.create)
	mux.HandleFunc("GET /campaigns", a.list)
	mux.HandleFunc("GET /campaigns/{id}", a.getByID)
	mux.HandleFunc("PATCH /campaigns/{id}", a.update)
	mux.HandleFunc("DELETE /campaigns/{id}", a.delete)

	mux.HandleFunc("POST /campaigns/{id}/players/{player_id}", a.addPlayer)
	mux.HandleFunc("DELETE /campaigns/{id}/players/{player_id}", a.removePlayer)
	mux.HandleFunc("GET /campaigns/{id}/players/{player_id}", a.getPlayerRelation)
	mux.HandleFunc("GET /campaigns/{id}/players", a.listPlayersByCampaign)
	mux.HandleFunc("GET /players/{id}/campaigns", a.listCampaignsByPlayer)
}

func (a *CampaignAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateCampaignRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	startDate, err := helpers.ParseDatePointer(payload.StartDate, "start_date")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	endDate, err := helpers.ParseDatePointer(payload.EndDate, "end_date")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	created, err := a.service.Create(r.Context(), service.CreateCampaignParams{
		WorldID:   model.ULID(payload.WorldID),
		Name:      payload.Name,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, helpers.MapCampaignToDTO(created))
}

func (a *CampaignAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)

	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapCampaignToDTO(item))
}

func (a *CampaignAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.List(r.Context(), service.ListCampaignsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.CampaignResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.CampaignResponse{
			ID:        string(item.ID),
			WorldID:   string(item.WorldID),
			Name:      item.Name,
			StartDate: helpers.FormatDatePointer(item.StartDate),
			EndDate:   helpers.FormatDatePointer(item.EndDate),
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
			Version:   int32(item.Version),
		})
	}

	helpers.WriteJSON(w, http.StatusOK, dto.ListCampaignsResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *CampaignAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	var payload dto.UpdateCampaignRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	startDate, err := helpers.ParseDatePointer(payload.StartDate, "start_date")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	endDate, err := helpers.ParseDatePointer(payload.EndDate, "end_date")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	updated, err := a.service.Update(r.Context(), service.UpdateCampaignParams{
		ID:              id,
		Name:            payload.Name,
		StartDate:       startDate,
		EndDate:         endDate,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapCampaignToDTO(updated))
}

func (a *CampaignAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *CampaignAPI) addPlayer(w http.ResponseWriter, r *http.Request) {
	campaignID := model.ULID(r.PathValue("id"))
	playerID := model.ULID(r.PathValue("player_id"))

	rel, err := a.service.AddPlayer(r.Context(), campaignID, playerID)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapCampaignPlayerToDTO(rel))
}

func (a *CampaignAPI) removePlayer(w http.ResponseWriter, r *http.Request) {
	campaignID := model.ULID(r.PathValue("id"))
	playerID := model.ULID(r.PathValue("player_id"))

	if err := a.service.RemovePlayer(r.Context(), campaignID, playerID); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *CampaignAPI) getPlayerRelation(w http.ResponseWriter, r *http.Request) {
	campaignID := model.ULID(r.PathValue("id"))
	playerID := model.ULID(r.PathValue("player_id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)

	rel, err := a.service.GetPlayerRelation(r.Context(), campaignID, playerID, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapCampaignPlayerToDTO(rel))
}

func (a *CampaignAPI) listPlayersByCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.ListPlayers(r.Context(), campaignID, service.ListCampaignsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.CampaignPlayerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapCampaignPlayerToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListCampaignPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *CampaignAPI) listCampaignsByPlayer(w http.ResponseWriter, r *http.Request) {
	playerID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.ListCampaignsForPlayer(r.Context(), playerID, service.ListCampaignsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.CampaignPlayerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapCampaignPlayerToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListCampaignPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}
