package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/api/shared"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
)

type PlayerService interface {
	Create(ctx context.Context, params service.CreatePlayerParams) (model.Player, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error)
	List(ctx context.Context, params service.ListPlayersParams) ([]service.PlayerListItem, error)
	SearchByName(ctx context.Context, params service.SearchPlayersParams) ([]service.PlayerListItem, error)
	Update(ctx context.Context, params service.UpdatePlayerParams) (model.Player, error)
	Delete(ctx context.Context, id model.ULID) error
	Restore(ctx context.Context, id model.ULID) (model.Player, error)
}

type PlayerAPI struct {
	service PlayerService
}

func NewPlayerAPI(service PlayerService) *PlayerAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "player", "service"))
	}
	return &PlayerAPI{service: service}
}

func (a *PlayerAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /players", a.create)
	mux.HandleFunc("GET /players", a.list)
	mux.HandleFunc("GET /players/search", a.search)
	mux.HandleFunc("GET /players/{id}", a.getByID)
	mux.HandleFunc("PATCH /players/{id}", a.update)
	mux.HandleFunc("DELETE /players/{id}", a.delete)
	mux.HandleFunc("POST /players/{id}/restore", a.restore)
}

func (a *PlayerAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePlayerRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	created, err := a.service.Create(r.Context(), service.CreatePlayerParams{
		Name: payload.Name,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, helpers.MapPlayerToDTO(created))
}

func (a *PlayerAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)

	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.MapPlayerToDTO(item))
}

func (a *PlayerAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.List(r.Context(), service.ListPlayersParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.PlayerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.PlayerResponse{
			ID:        string(item.ID),
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
			Version:   int32(item.Version),
		})
	}

	helpers.WriteJSON(w, http.StatusOK, dto.ListPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *PlayerAPI) search(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		helpers.WriteBadRequest(w, apierrors.APIQUERYREQUIREDMESSAGE)
		return
	}

	items, svcErr := a.service.SearchByName(r.Context(), service.SearchPlayersParams{
		Query:          query,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.PlayerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.PlayerResponse{
			ID:        string(item.ID),
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
			Version:   int32(item.Version),
		})
	}

	helpers.WriteJSON(w, http.StatusOK, dto.ListPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *PlayerAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	var payload dto.UpdatePlayerRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	updated, err := a.service.Update(r.Context(), service.UpdatePlayerParams{
		ID:              id,
		Name:            payload.Name,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.MapPlayerToDTO(updated))
}

func (a *PlayerAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *PlayerAPI) restore(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	restored, err := a.service.Restore(r.Context(), id)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.MapPlayerToDTO(restored))
}
