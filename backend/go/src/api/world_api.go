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

type WorldService interface {
	Create(ctx context.Context, params service.CreateWorldParams) (model.World, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error)
	List(ctx context.Context, params service.ListWorldsParams) ([]service.WorldListItem, error)
	Update(ctx context.Context, params service.UpdateWorldParams) (model.World, error)
	Delete(ctx context.Context, id model.ULID) error
}

type WorldAPI struct {
	service WorldService
}

func NewWorldAPI(service WorldService) *WorldAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "world", "service"))
	}
	return &WorldAPI{service: service}
}

func (a *WorldAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /worlds", a.create)
	mux.HandleFunc("GET /worlds", a.list)
	mux.HandleFunc("GET /worlds/{id}", a.getByID)
	mux.HandleFunc("PATCH /worlds/{id}", a.update)
	mux.HandleFunc("DELETE /worlds/{id}", a.delete)
}

func (a *WorldAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateWorldRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	status, err := helpers.ParseWorldStatus(payload.Status)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	created, err := a.service.Create(r.Context(), service.CreateWorldParams{
		PlaneID:     model.ULID(payload.PlaneID),
		Name:        payload.Name,
		Description: payload.Description,
		Status:      status,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapWorldToDTO(created))
}

func (a *WorldAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)

	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapWorldToDTO(item))
}

func (a *WorldAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.List(r.Context(), service.ListWorldsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}

	out := make([]dto.WorldResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.WorldResponse{
			ID:          string(item.ID),
			PlaneID:     string(item.PlaneID),
			Name:        item.Name,
			Description: item.Description,
			Status:      item.Status.String(),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			DeletedAt:   item.DeletedAt,
			Version:     int32(item.Version),
		})
	}

	helpers.WriteJSON(w, http.StatusOK, dto.ListWorldsResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *WorldAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateWorldRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	status, err := helpers.ParseWorldStatus(payload.Status)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}

	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}

	updated, err := a.service.Update(r.Context(), service.UpdateWorldParams{
		ID:              id,
		Name:            payload.Name,
		Description:     payload.Description,
		Status:          status,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapWorldToDTO(updated))
}

func (a *WorldAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
