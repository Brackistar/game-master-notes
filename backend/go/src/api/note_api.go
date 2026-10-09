package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Brackistar/game-master-notes/backend/go/src/api/dto"
	apierrors "github.com/Brackistar/game-master-notes/backend/go/src/api/error"
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/api/shared"
	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	service "github.com/Brackistar/game-master-notes/backend/go/src/service"
)

type NoteService interface {
	Create(ctx context.Context, params service.CreateNoteParams) (model.Note, error)
	GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Note, error)
	List(ctx context.Context, params service.ListNotesParams) ([]service.NoteListItem, error)
	Update(ctx context.Context, params service.UpdateNoteParams) (model.Note, error)
	Delete(ctx context.Context, id model.ULID) error

	AddOwner(ctx context.Context, params service.AddNoteOwnerParams) (model.NoteOwner, error)
	RemoveOwner(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID) error
	GetOwner(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, includeDeleted bool) (model.NoteOwner, error)
	ListOwnersByNote(ctx context.Context, noteID model.ULID, params service.RelationListParams) ([]model.NoteOwner, error)
	ListNotesByOwner(ctx context.Context, ownerType constants.OwnerType, ownerID model.ULID, params service.RelationListParams) ([]model.NoteOwner, error)

	AddTag(ctx context.Context, params service.AddNoteTagParams) (model.NoteTag, error)
	RemoveTag(ctx context.Context, noteID, tagID model.ULID) error
	GetTag(ctx context.Context, noteID, tagID model.ULID, includeDeleted bool) (model.NoteTag, error)
	ListTagsByNote(ctx context.Context, noteID model.ULID, params service.RelationListParams) ([]model.NoteTag, error)
	ListNotesByTag(ctx context.Context, tagID model.ULID, params service.RelationListParams) ([]model.NoteTag, error)

	CreateLink(ctx context.Context, params service.CreateNoteLinkParams) (model.NoteLink, error)
	GetLinkByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteLink, error)
	ListLinksBySource(ctx context.Context, sourceNoteID model.ULID, params service.RelationListParams) ([]model.NoteLink, error)
	ListLinksByTarget(ctx context.Context, targetNoteID model.ULID, params service.RelationListParams) ([]model.NoteLink, error)
	UpdateLink(ctx context.Context, params service.UpdateNoteLinkParams) (model.NoteLink, error)
	DeleteLink(ctx context.Context, id model.ULID) error

	CreateAsset(ctx context.Context, params service.CreateNoteAssetParams) (model.NoteAsset, error)
	GetAssetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteAsset, error)
	ListAssetsByNote(ctx context.Context, noteID model.ULID, params service.RelationListParams) ([]model.NoteAsset, error)
	UpdateAsset(ctx context.Context, params service.UpdateNoteAssetParams) (model.NoteAsset, error)
	DeleteAsset(ctx context.Context, id model.ULID) error

	UpsertMapPlacement(ctx context.Context, params service.UpsertMapNotePlacementParams) (model.MapNotePlacement, error)
	GetMapPlacementByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.MapNotePlacement, error)
	ListMapPlacementsByMap(ctx context.Context, mapNoteID model.ULID, params service.RelationListParams) ([]model.MapNotePlacement, error)
	ListMapPlacementsByTarget(ctx context.Context, targetNoteID model.ULID, params service.RelationListParams) ([]model.MapNotePlacement, error)
	UpdateMapPlacement(ctx context.Context, params service.UpdateMapNotePlacementParams) (model.MapNotePlacement, error)
	DeleteMapPlacement(ctx context.Context, id model.ULID) error
}

type NoteAPI struct {
	service NoteService
}

func NewNoteAPI(service NoteService) *NoteAPI {
	if service == nil {
		panic(fmt.Sprintf(apierrors.APIDEPNILMESSAGE, "note", "service"))
	}
	return &NoteAPI{service: service}
}

func (a *NoteAPI) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /notes", a.create)
	mux.HandleFunc("GET /notes", a.list)
	mux.HandleFunc("GET /notes/{id}", a.getByID)
	mux.HandleFunc("PATCH /notes/{id}", a.update)
	mux.HandleFunc("DELETE /notes/{id}", a.delete)

	mux.HandleFunc("POST /notes/{id}/owners", a.addOwner)
	mux.HandleFunc("GET /notes/{id}/owners", a.listOwnersByNote)
	mux.HandleFunc("GET /notes/{id}/owners/{owner_type}/{owner_id}", a.getOwner)
	mux.HandleFunc("DELETE /notes/{id}/owners/{owner_type}/{owner_id}", a.removeOwner)
	mux.HandleFunc("GET /owners/{owner_type}/{owner_id}/notes", a.listNotesByOwner)

	mux.HandleFunc("POST /notes/{id}/tags/{tag_id}", a.addTag)
	mux.HandleFunc("GET /notes/{id}/tags", a.listTagsByNote)
	mux.HandleFunc("GET /notes/{id}/tags/{tag_id}", a.getTag)
	mux.HandleFunc("DELETE /notes/{id}/tags/{tag_id}", a.removeTag)
	mux.HandleFunc("GET /tags/{id}/notes", a.listNotesByTag)

	mux.HandleFunc("POST /note-links", a.createLink)
	mux.HandleFunc("GET /note-links/{id}", a.getLinkByID)
	mux.HandleFunc("PATCH /note-links/{id}", a.updateLink)
	mux.HandleFunc("DELETE /note-links/{id}", a.deleteLink)
	mux.HandleFunc("GET /notes/{id}/outgoing-links", a.listLinksBySource)
	mux.HandleFunc("GET /notes/{id}/incoming-links", a.listLinksByTarget)

	mux.HandleFunc("POST /notes/{id}/assets", a.createAsset)
	mux.HandleFunc("GET /notes/{id}/assets", a.listAssetsByNote)
	mux.HandleFunc("GET /note-assets/{id}", a.getAssetByID)
	mux.HandleFunc("PATCH /note-assets/{id}", a.updateAsset)
	mux.HandleFunc("DELETE /note-assets/{id}", a.deleteAsset)

	mux.HandleFunc("POST /map-note-placements", a.upsertMapPlacement)
	mux.HandleFunc("GET /map-note-placements/{id}", a.getMapPlacementByID)
	mux.HandleFunc("PATCH /map-note-placements/{id}", a.updateMapPlacement)
	mux.HandleFunc("DELETE /map-note-placements/{id}", a.deleteMapPlacement)
	mux.HandleFunc("GET /notes/{id}/map-placements/as-map", a.listMapPlacementsByMap)
	mux.HandleFunc("GET /notes/{id}/map-placements/as-target", a.listMapPlacementsByTarget)
}

func (a *NoteAPI) create(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateNoteRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	noteType, err := helpers.ParseNoteType(payload.NoteType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	created, err := a.service.Create(r.Context(), service.CreateNoteParams{
		Title:        payload.Title,
		ContentMD:    payload.ContentMD,
		NoteType:     noteType,
		MetadataJSON: payload.MetadataJSON,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapNoteToDTO(created))
}

func (a *NoteAPI) getByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteToDTO(item))
}

func (a *NoteAPI) list(w http.ResponseWriter, r *http.Request) {
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.List(r.Context(), service.ListNotesParams{
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.NoteResponse{
			ID:           string(item.ID),
			Title:        item.Title,
			ContentMD:    item.ContentMD,
			NoteType:     item.NoteType.String(),
			MetadataJSON: item.MetadataJSON,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
			DeletedAt:    item.DeletedAt,
			Version:      int32(item.Version),
		})
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNotesResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) update(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateNoteRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	noteType, err := helpers.ParseNoteType(payload.NoteType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	current, err := a.service.GetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.Update(r.Context(), service.UpdateNoteParams{
		ID:              id,
		Title:           payload.Title,
		ContentMD:       payload.ContentMD,
		NoteType:        noteType,
		MetadataJSON:    payload.MetadataJSON,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteToDTO(updated))
}

func (a *NoteAPI) delete(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.Delete(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *NoteAPI) addOwner(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	var payload dto.AddNoteOwnerRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	ownerType, err := helpers.ParseOwnerType(payload.OwnerType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	rel, err := a.service.AddOwner(r.Context(), service.AddNoteOwnerParams{
		NoteID:    noteID,
		OwnerType: ownerType,
		OwnerID:   model.ULID(payload.OwnerID),
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapNoteOwnerToDTO(rel))
}

func (a *NoteAPI) removeOwner(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	ownerType, err := helpers.ParseOwnerType(r.PathValue("owner_type"))
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	ownerID := model.ULID(r.PathValue("owner_id"))
	if err := a.service.RemoveOwner(r.Context(), noteID, ownerType, ownerID); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *NoteAPI) getOwner(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	ownerType, err := helpers.ParseOwnerType(r.PathValue("owner_type"))
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	ownerID := model.ULID(r.PathValue("owner_id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	rel, svcErr := a.service.GetOwner(r.Context(), noteID, ownerType, ownerID, includeDeleted)
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteOwnerToDTO(rel))
}

func (a *NoteAPI) listOwnersByNote(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListOwnersByNote(r.Context(), noteID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteOwnerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteOwnerToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteOwnersResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) listNotesByOwner(w http.ResponseWriter, r *http.Request) {
	ownerType, err := helpers.ParseOwnerType(r.PathValue("owner_type"))
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	ownerID := model.ULID(r.PathValue("owner_id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListNotesByOwner(r.Context(), ownerType, ownerID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteOwnerResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteOwnerToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteOwnersResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) addTag(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	tagID := model.ULID(r.PathValue("tag_id"))
	rel, err := a.service.AddTag(r.Context(), service.AddNoteTagParams{NoteID: noteID, TagID: tagID})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapNoteTagToDTO(rel))
}

func (a *NoteAPI) removeTag(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	tagID := model.ULID(r.PathValue("tag_id"))
	if err := a.service.RemoveTag(r.Context(), noteID, tagID); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *NoteAPI) getTag(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	tagID := model.ULID(r.PathValue("tag_id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	rel, err := a.service.GetTag(r.Context(), noteID, tagID, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteTagToDTO(rel))
}

func (a *NoteAPI) listTagsByNote(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListTagsByNote(r.Context(), noteID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteTagResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteTagToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteTagsResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) listNotesByTag(w http.ResponseWriter, r *http.Request) {
	tagID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListNotesByTag(r.Context(), tagID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteTagResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteTagToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteTagsResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) createLink(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreateNoteLinkRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	linkType, err := helpers.ParseNoteLinkType(payload.LinkType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	created, err := a.service.CreateLink(r.Context(), service.CreateNoteLinkParams{
		SourceNoteID: model.ULID(payload.SourceNoteID),
		TargetNoteID: model.ULID(payload.TargetNoteID),
		LinkType:     linkType,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapNoteLinkToDTO(created))
}

func (a *NoteAPI) getLinkByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetLinkByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteLinkToDTO(item))
}

func (a *NoteAPI) listLinksBySource(w http.ResponseWriter, r *http.Request) {
	sourceID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListLinksBySource(r.Context(), sourceID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteLinkResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteLinkToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteLinksResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) listLinksByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListLinksByTarget(r.Context(), targetID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteLinkResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteLinkToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteLinksResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) updateLink(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateNoteLinkRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	linkType, err := helpers.ParseNoteLinkType(payload.LinkType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	current, err := a.service.GetLinkByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.UpdateLink(r.Context(), service.UpdateNoteLinkParams{
		ID:              id,
		LinkType:        linkType,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteLinkToDTO(updated))
}

func (a *NoteAPI) deleteLink(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.DeleteLink(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *NoteAPI) createAsset(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	var payload dto.CreateNoteAssetRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	assetType, err := helpers.ParseAssetType(payload.AssetType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	created, err := a.service.CreateAsset(r.Context(), service.CreateNoteAssetParams{
		NoteID:      noteID,
		AssetType:   assetType,
		StoragePath: payload.StoragePath,
		MIMEType:    payload.MIMEType,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapNoteAssetToDTO(created))
}

func (a *NoteAPI) getAssetByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetAssetByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteAssetToDTO(item))
}

func (a *NoteAPI) listAssetsByNote(w http.ResponseWriter, r *http.Request) {
	noteID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListAssetsByNote(r.Context(), noteID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.NoteAssetResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapNoteAssetToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListNoteAssetsResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) updateAsset(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateNoteAssetRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	assetType, err := helpers.ParseAssetType(payload.AssetType)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	current, err := a.service.GetAssetByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	updated, err := a.service.UpdateAsset(r.Context(), service.UpdateNoteAssetParams{
		ID:              id,
		AssetType:       assetType,
		StoragePath:     payload.StoragePath,
		MIMEType:        payload.MIMEType,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapNoteAssetToDTO(updated))
}

func (a *NoteAPI) deleteAsset(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.DeleteAsset(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *NoteAPI) upsertMapPlacement(w http.ResponseWriter, r *http.Request) {
	var payload dto.UpsertMapNotePlacementRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	item, err := a.service.UpsertMapPlacement(r.Context(), service.UpsertMapNotePlacementParams{
		MapNoteID:    model.ULID(payload.MapNoteID),
		TargetNoteID: model.ULID(payload.TargetNoteID),
		X:            payload.X,
		Y:            payload.Y,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusCreated, helpers.MapMapPlacementToDTO(item))
}

func (a *NoteAPI) getMapPlacementByID(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	includeDeleted := helpers.ParseBool(r.URL.Query().Get("include_deleted"), false)
	item, err := a.service.GetMapPlacementByID(r.Context(), id, includeDeleted)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapMapPlacementToDTO(item))
}

func (a *NoteAPI) listMapPlacementsByMap(w http.ResponseWriter, r *http.Request) {
	mapNoteID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListMapPlacementsByMap(r.Context(), mapNoteID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.MapNotePlacementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapMapPlacementToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListMapNotePlacementsResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) listMapPlacementsByTarget(w http.ResponseWriter, r *http.Request) {
	targetID := model.ULID(r.PathValue("id"))
	offset, limit, includeDeleted, err := helpers.ParseListQuery(r)
	if err != nil {
		helpers.WriteBadRequest(w, err.Error())
		return
	}
	items, svcErr := a.service.ListMapPlacementsByTarget(r.Context(), targetID, service.RelationListParams{Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
	if svcErr != nil {
		helpers.WriteServiceError(w, svcErr)
		return
	}
	out := make([]dto.MapNotePlacementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, helpers.MapMapPlacementToDTO(item))
	}
	helpers.WriteJSON(w, http.StatusOK, dto.ListMapNotePlacementsResponse{Items: out, Offset: offset, Limit: limit, IncludeDeleted: includeDeleted})
}

func (a *NoteAPI) updateMapPlacement(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	var payload dto.UpdateMapNotePlacementRequest
	if err := helpers.DecodeJSONBody(r, &payload); err != nil {
		helpers.WriteBadRequest(w, apierrors.APIINVALIDREQUESTBODY)
		return
	}
	current, err := a.service.GetMapPlacementByID(r.Context(), id, false)
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	item, err := a.service.UpdateMapPlacement(r.Context(), service.UpdateMapNotePlacementParams{
		ID:              id,
		X:               payload.X,
		Y:               payload.Y,
		ExpectedVersion: current.AuditFields.Version,
	})
	if err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, helpers.MapMapPlacementToDTO(item))
}

func (a *NoteAPI) deleteMapPlacement(w http.ResponseWriter, r *http.Request) {
	id := model.ULID(r.PathValue("id"))
	if err := a.service.DeleteMapPlacement(r.Context(), id); err != nil {
		helpers.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
