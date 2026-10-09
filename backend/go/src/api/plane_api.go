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

type PlaneService interface {
	Create(ctx context.Context, params service.CreatePlaneParams) (model.Plane, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error)
	List(ctx context.Context, params service.ListPlanesParams) ([]service.PlaneListItem, error)
	Update(ctx context.Context, params service.UpdatePlaneParams) (model.Plane, error)
	Delete(ctx context.Context, id model.ULID) error
}

type PlaneAPI struct {
	service PlaneService
}

func NewPlaneAPI(service PlaneService) *PlaneAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "plane", "service"))
	}
	return &PlaneAPI{service: service}
}

func (a *PlaneAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /planes", a.create)
	mux.HandleFunc("GET /planes", a.list)
	mux.HandleFunc("GET /planes/{id}", a.getByID)
	mux.HandleFunc("PATCH /planes/{id}", a.update)
	mux.HandleFunc("DELETE /planes/{id}", a.delete)
}

func (a *PlaneAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePlaneRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	created, err := a.service.Create(r.Context(), service.CreatePlaneParams{
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapPlaneToDTO(created))
}

func (a *PlaneAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapPlaneToDTO(item))
}

func (a *PlaneAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.List(r.Context(), service.ListPlanesParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.PlaneResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.PlaneResponse{
			ID:          string(item.ID),
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			DeletedAt:   item.DeletedAt,
			Version:     int32(item.Version),
		})
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListPlanesResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *PlaneAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdatePlaneRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.Update(r.Context(), service.UpdatePlaneParams{
		ID:              id,
		Name:            payload.Name,
		Description:     payload.Description,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapPlaneToDTO(updated))
}

func (a *PlaneAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
