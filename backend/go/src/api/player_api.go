package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

const (
	defaultListOffset int32 = 0
	defaultListLimit  int32 = 20
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
	if err := decodeJSONBody(r, &payload); err != nil {
		writeBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	created, err := a.service.Create(r.Context(), service.CreatePlayerParams{
		Name: payload.Name,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapPlayerToDTO(created))
}

func (a *PlayerAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := parseBool(r.URL.Query().Get("include_deleted"), false)

	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapPlayerToDTO(item))
}

func (a *PlayerAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := parseListQuery(r)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}

	items, svcErr := a.service.List(r.Context(), service.ListPlayersParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		writeServiceError(w, svcErr)
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

	writeJSON(w, http.StatusOK, dto.ListPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *PlayerAPI) search(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := parseListQuery(r)
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeBadRequest(w, apierrors.APIQUERYREQUIREDMESSAGE)
		return
	}

	items, svcErr := a.service.SearchByName(r.Context(), service.SearchPlayersParams{
		Query:          query,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		writeServiceError(w, svcErr)
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

	writeJSON(w, http.StatusOK, dto.ListPlayersResponse{
		Items:          out,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
}

func (a *PlayerAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	var payload dto.UpdatePlayerRequest
	if err := decodeJSONBody(r, &payload); err != nil {
		writeBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}

	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	updated, err := a.service.Update(r.Context(), service.UpdatePlayerParams{
		ID:              id,
		Name:            payload.Name,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapPlayerToDTO(updated))
}

func (a *PlayerAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	if err := a.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *PlayerAPI) restore(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))

	restored, err := a.service.Restore(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapPlayerToDTO(restored))
}

func mapPlayerToDTO(player model.Player) dto.PlayerResponse {
	return dto.PlayerResponse{
		ID:        string(player.ID),
		Name:      player.Name,
		CreatedAt: player.AuditFields.CreatedAt,
		UpdatedAt: player.AuditFields.UpdatedAt,
		DeletedAt: player.AuditFields.DeletedAt,
		Version:   int32(player.AuditFields.Version),
	}
}

func parseListQuery(r *http.Request) (int32, int32, bool, error) {
	offset, err := parseInt32OrDefault(r.URL.Query().Get("offset"), defaultListOffset, "offset")
	if err != nil {
		return 0, 0, false, err
	}
	limit, err := parseInt32OrDefault(r.URL.Query().Get("limit"), defaultListLimit, "limit")
	if err != nil {
		return 0, 0, false, err
	}
	includeDeleted := parseBool(r.URL.Query().Get("include_deleted"), false)
	return offset, limit, includeDeleted, nil
}

func parseInt32OrDefault(value string, fallback int32, field string) (int32, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf(apierrors.APIFIELDVALIDINTEGER, field)
	}
	return int32(parsed), nil
}

func parseBool(value string, fallback bool) bool {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func decodeJSONBody(r *http.Request, out any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || strings.EqualFold(trimmed, "null") {
		return errors.New(apierrors.APIBODYEMPTYMESSAGE)
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New(apierrors.APIINVALIDJSONPAYLOAD)
	}
	return nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, serviceerrors.ErrValidation):
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: apierrors.APIINVALIDREQUESTMESSAGE})
	case errors.Is(err, serviceerrors.ErrNotFound):
		writeJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: apierrors.APIRESOURCENOTFOUNDMESSAGE})
	case errors.Is(err, serviceerrors.ErrConflict):
		writeJSON(w, http.StatusConflict, dto.ErrorResponse{Message: apierrors.APIREQUESTCONFLICTMESSAGE})
	default:
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: apierrors.APIINTERNALSERVERERROR})
	}
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
