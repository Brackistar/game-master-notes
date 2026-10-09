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

type SessionService interface {
	Create(ctx context.Context, params service.CreateSessionParams) (model.Session, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Session, error)
	List(ctx context.Context, params service.ListSessionsParams) ([]service.SessionListItem, error)
	Update(ctx context.Context, params service.UpdateSessionParams) (model.Session, error)
	Delete(ctx context.Context, id model.ULID) error
}

type SessionAPI struct {
	service SessionService
}

func NewSessionAPI(service SessionService) *SessionAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "session", "service"))
	}
	return &SessionAPI{service: service}
}

func (a *SessionAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /sessions", a.create)
	mux.HandleFunc("GET /sessions", a.list)
	mux.HandleFunc("GET /sessions/{id}", a.getByID)
	mux.HandleFunc("PATCH /sessions/{id}", a.update)
	mux.HandleFunc("DELETE /sessions/{id}", a.delete)
}

func (a *SessionAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateSessionRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	playedOn, err := helpers.ParseDatePointer(payload.PlayedOn, "played_on")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	created, err := a.service.Create(r.Context(), service.CreateSessionParams{
		CampaignID: model.ULID(payload.CampaignID),
		PlayedOn:   playedOn,
		SummaryMD:  payload.SummaryMD,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapSessionToDTO(created))
}

func (a *SessionAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapSessionToDTO(item))
}

func (a *SessionAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.List(r.Context(), service.ListSessionsParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.SessionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.SessionResponse{
			ID:         string(item.ID),
			CampaignID: string(item.CampaignID),
			PlayedOn:   helpers.FormatDatePointer(item.PlayedOn),
			SummaryMD:  item.SummaryMD,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			DeletedAt:  item.DeletedAt,
			Version:    int32(item.Version),
		})
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListSessionsResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *SessionAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateSessionRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	playedOn, err := helpers.ParseDatePointer(payload.PlayedOn, "played_on")
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.Update(r.Context(), service.UpdateSessionParams{
		ID:              id,
		PlayedOn:        playedOn,
		SummaryMD:       payload.SummaryMD,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapSessionToDTO(updated))
}

func (a *SessionAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
