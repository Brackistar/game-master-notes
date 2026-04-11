package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
)

const (
	defaultListOffset int32 = 0
	defaultListLimit  int32 = 20
)

func ParseListQuery(r *http.Request) (int32, int32, bool, error) {
	offset, err := parseInt32OrDefault(r.URL.Query().Get("offset"), defaultListOffset, "offset")
	if err != nil {
		return 0, 0, false, err
	}
	limit, err := parseInt32OrDefault(r.URL.Query().Get("limit"), defaultListLimit, "limit")
	if err != nil {
		return 0, 0, false, err
	}
	includeDeleted := ParseBool(r.URL.Query().Get("include_deleted"), false)
	return offset, limit, includeDeleted, nil
}

func ParseBool(value string, fallback bool) bool {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func DecodeJSONBody(r *http.Request, out any) error {
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

func WriteServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, serviceerrors.ErrValidation):
		WriteJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: apierrors.APIINVALIDREQUESTMESSAGE})
	case errors.Is(err, serviceerrors.ErrNotFound):
		WriteJSON(w, http.StatusNotFound, dto.ErrorResponse{Message: apierrors.APIRESOURCENOTFOUNDMESSAGE})
	case errors.Is(err, serviceerrors.ErrConflict):
		WriteJSON(w, http.StatusConflict, dto.ErrorResponse{Message: apierrors.APIREQUESTCONFLICTMESSAGE})
	default:
		WriteJSON(w, http.StatusInternalServerError, dto.ErrorResponse{Message: apierrors.APIINTERNALSERVERERROR})
	}
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: message})
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
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
