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

type TagService interface {
	Create(ctx context.Context, params service.CreateTagParams) (model.Tag, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error)
	List(ctx context.Context, params service.ListTagsParams) ([]service.TagListItem, error)
	Update(ctx context.Context, params service.UpdateTagParams) (model.Tag, error)
	Delete(ctx context.Context, id model.ULID) error
}

type TagAPI struct {
	service TagService
}

func NewTagAPI(service TagService) *TagAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "tag", "service"))
	}
	return &TagAPI{service: service}
}

func (a *TagAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /tags", a.create)
	mux.HandleFunc("GET /tags", a.list)
	mux.HandleFunc("GET /tags/{id}", a.getByID)
	mux.HandleFunc("PATCH /tags/{id}", a.update)
	mux.HandleFunc("DELETE /tags/{id}", a.delete)
}

func (a *TagAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateTagRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	created, err := a.service.Create(r.Context(), service.CreateTagParams{
		Name:       payload.Name,
		CampaignID: helpers.FromStringPointer(payload.CampaignID),
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapTagToDTO(created))
}

func (a *TagAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapTagToDTO(item))
}

func (a *TagAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.List(r.Context(), service.ListTagsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.TagResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.TagResponse{
			ID:         string(item.ID),
			Name:       item.Name,
			CampaignID: helpers.ToStringPointer(item.CampaignID),
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			DeletedAt:  item.DeletedAt,
			Version:    int32(item.Version),
		})
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListTagsResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *TagAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateTagRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.Update(r.Context(), service.UpdateTagParams{
		ID:              id,
		Name:            payload.Name,
		CampaignID:      helpers.FromStringPointer(payload.CampaignID),
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapTagToDTO(updated))
}

func (a *TagAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
