package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	"github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const noteServiceName string = "note"

type NotePolicy interface {
	NormalizeAndValidate(title, contentMD string, noteType constants.NoteType, metadataJSON []byte) (string, string, []byte, error)
}

type DefaultNotePolicy struct{}

func (DefaultNotePolicy) NormalizeAndValidate(title, contentMD string, noteType constants.NoteType, metadataJSON []byte) (string, string, []byte, error) {
	normalizedTitle := shared.NormalizeSpaces(title)
	if normalizedTitle == "" {
		return "", "", nil, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "title")
	}
	if !isValidNoteType(noteType) {
		return "", "", nil, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "note_type")
	}
	if len(metadataJSON) == 0 {
		metadataJSON = []byte("{}")
	}
	if !json.Valid(metadataJSON) {
		return "", "", nil, errors.New(serviceerrors.SERVMETADATAJSONVALIDMESSAGE)
	}
	return normalizedTitle, strings.TrimSpace(contentMD), metadataJSON, nil
}

type CreateNoteParams struct {
	Title        string
	ContentMD    string
	NoteType     constants.NoteType
	MetadataJSON []byte
}

type UpdateNoteParams struct {
	ID              model.ULID
	Title           string
	ContentMD       string
	NoteType        constants.NoteType
	MetadataJSON    []byte
	ExpectedVersion model.Version
}

type ListNotesParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type RelationListParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type NoteListItem struct {
	ID           model.ULID
	Title        string
	ContentMD    string
	NoteType     constants.NoteType
	MetadataJSON []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	Version      model.Version
}

type AddNoteOwnerParams struct {
	NoteID    model.ULID
	OwnerType constants.OwnerType
	OwnerID   model.ULID
}

type AddNoteTagParams struct {
	NoteID model.ULID
	TagID  model.ULID
}

type CreateNoteLinkParams struct {
	SourceNoteID model.ULID
	TargetNoteID model.ULID
	LinkType     constants.NoteLinkType
}

type UpdateNoteLinkParams struct {
	ID              model.ULID
	LinkType        constants.NoteLinkType
	ExpectedVersion model.Version
}

type CreateNoteAssetParams struct {
	NoteID      model.ULID
	AssetType   constants.AssetType
	StoragePath string
	MIMEType    string
}

type UpdateNoteAssetParams struct {
	ID              model.ULID
	AssetType       constants.AssetType
	StoragePath     string
	MIMEType        string
	ExpectedVersion model.Version
}

type UpsertMapNotePlacementParams struct {
	MapNoteID    model.ULID
	TargetNoteID model.ULID
	X            uint8
	Y            uint8
}

type UpdateMapNotePlacementParams struct {
	ID              model.ULID
	X               uint8
	Y               uint8
	ExpectedVersion model.Version
}

type NoteService struct {
	repo                 repo.NoteRepository
	noteOwnerRepo        repo.NoteOwnerRepository
	noteTagRepo          repo.NoteTagRepository
	noteLinkRepo         repo.NoteLinkRepository
	noteAssetRepo        repo.NoteAssetRepository
	mapNotePlacementRepo repo.MapNotePlacementRepository
	clock                shared.Clock
	policy               NotePolicy
	idGenerator          shared.IDGenerator
}

type NoteServiceDeps struct {
	Repo                 repo.NoteRepository
	NoteOwnerRepo        repo.NoteOwnerRepository
	NoteTagRepo          repo.NoteTagRepository
	NoteLinkRepo         repo.NoteLinkRepository
	NoteAssetRepo        repo.NoteAssetRepository
	MapNotePlacementRepo repo.MapNotePlacementRepository
	Clock                shared.Clock
	Policy               NotePolicy
	IDGenerator          shared.IDGenerator
}

func NewNoteService(
	repo repo.NoteRepository,
	noteOwnerRepo repo.NoteOwnerRepository,
	noteTagRepo repo.NoteTagRepository,
	noteLinkRepo repo.NoteLinkRepository,
	noteAssetRepo repo.NoteAssetRepository,
	mapNotePlacementRepo repo.MapNotePlacementRepository,
	idGenerator shared.IDGenerator,
) *NoteService {
	return NewNoteServiceWithDeps(NoteServiceDeps{
		Repo:                 repo,
		NoteOwnerRepo:        noteOwnerRepo,
		NoteTagRepo:          noteTagRepo,
		NoteLinkRepo:         noteLinkRepo,
		NoteAssetRepo:        noteAssetRepo,
		MapNotePlacementRepo: mapNotePlacementRepo,
		Clock:                shared.SystemClock{},
		Policy:               DefaultNotePolicy{},
		IDGenerator:          idGenerator,
	})
}

func NewNoteServiceWithDeps(deps NoteServiceDeps) *NoteService {
	shared.PanicIfNilDependency(noteServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(noteServiceName, "NoteOwnerRepo", deps.NoteOwnerRepo)
	shared.PanicIfNilDependency(noteServiceName, "NoteTagRepo", deps.NoteTagRepo)
	shared.PanicIfNilDependency(noteServiceName, "NoteLinkRepo", deps.NoteLinkRepo)
	shared.PanicIfNilDependency(noteServiceName, "NoteAssetRepo", deps.NoteAssetRepo)
	shared.PanicIfNilDependency(noteServiceName, "MapNotePlacementRepo", deps.MapNotePlacementRepo)
	shared.PanicIfNilDependency(noteServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(noteServiceName, "Policy", deps.Policy)
	shared.PanicIfNilDependency(noteServiceName, "IDGenerator", deps.IDGenerator)
	return &NoteService{
		repo:                 deps.Repo,
		noteOwnerRepo:        deps.NoteOwnerRepo,
		noteTagRepo:          deps.NoteTagRepo,
		noteLinkRepo:         deps.NoteLinkRepo,
		noteAssetRepo:        deps.NoteAssetRepo,
		mapNotePlacementRepo: deps.MapNotePlacementRepo,
		clock:                deps.Clock,
		policy:               deps.Policy,
		idGenerator:          deps.IDGenerator,
	}
}

func (s *NoteService) Create(ctx context.Context, params CreateNoteParams) (model.Note, error) {
	defer shared.LogServiceCall()()
	op := "note_service.create"
	title, contentMD, metadataJSON, err := s.policy.NormalizeAndValidate(params.Title, params.ContentMD, params.NoteType, params.MetadataJSON)
	if err != nil {
		return model.Note{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Note{}, serviceerrors.WrapUnknown(op, noteServiceName, err)
	}
	now := s.clock.Now()
	note, repoErr := s.repo.Create(ctx, model.Note{
		ID:           id,
		Title:        title,
		ContentMD:    contentMD,
		NoteType:     params.NoteType,
		MetadataJSON: metadataJSON,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Note{}, shared.MapRepositoryError(repoErr, op, noteServiceName)
	}
	return note, nil
}

func (s *NoteService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Note, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Note{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	note, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Note{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return note, nil
}

func (s *NoteService) List(ctx context.Context, params ListNotesParams) ([]NoteListItem, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list"
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.repo.List(ctx, repo.ListNotesParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return toNoteListItems(rows), nil
}

func (s *NoteService) Update(ctx context.Context, params UpdateNoteParams) (model.Note, error) {
	defer shared.LogServiceCall()()
	op := "note_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Note{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Note{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	title, contentMD, metadataJSON, err := s.policy.NormalizeAndValidate(params.Title, params.ContentMD, params.NoteType, params.MetadataJSON)
	if err != nil {
		return model.Note{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	note, repoErr := s.repo.Update(ctx, repo.UpdateNoteParams{
		ID:              params.ID,
		Title:           title,
		ContentMD:       contentMD,
		NoteType:        params.NoteType,
		MetadataJSON:    metadataJSON,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Note{}, shared.MapRepositoryError(repoErr, op, noteServiceName)
	}
	return note, nil
}

func (s *NoteService) Delete(ctx context.Context, id model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func (s *NoteService) AddOwner(ctx context.Context, params AddNoteOwnerParams) (model.NoteOwner, error) {
	defer shared.LogServiceCall()()
	op := "note_service.add_owner"
	if err := validateNoteOwnerIDs(params.NoteID, params.OwnerID); err != nil {
		return model.NoteOwner{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if !isValidOwnerType(params.OwnerType) {
		return model.NoteOwner{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "owner_type"))
	}
	rel, err := s.noteOwnerRepo.Create(ctx, model.NoteOwner{
		NoteID:    params.NoteID,
		OwnerType: params.OwnerType,
		OwnerID:   params.OwnerID,
	})
	if err != nil {
		return model.NoteOwner{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rel, nil
}

func (s *NoteService) RemoveOwner(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.remove_owner"
	if err := validateNoteOwnerIDs(noteID, ownerID); err != nil {
		return serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if !isValidOwnerType(ownerType) {
		return serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "owner_type"))
	}
	if err := s.noteOwnerRepo.Delete(ctx, noteID, ownerType, ownerID); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func (s *NoteService) GetOwner(ctx context.Context, noteID model.ULID, ownerType constants.OwnerType, ownerID model.ULID, includeDeleted bool) (model.NoteOwner, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_owner"
	if err := validateNoteOwnerIDs(noteID, ownerID); err != nil {
		return model.NoteOwner{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if !isValidOwnerType(ownerType) {
		return model.NoteOwner{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "owner_type"))
	}
	rel, err := s.noteOwnerRepo.Get(ctx, noteID, ownerType, ownerID, includeDeleted)
	if err != nil {
		return model.NoteOwner{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rel, nil
}

func (s *NoteService) ListOwnersByNote(ctx context.Context, noteID model.ULID, params RelationListParams) ([]model.NoteOwner, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_owners_by_note"
	if strings.TrimSpace(string(noteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteOwnerRepo.ListByNote(ctx, noteID, repo.ListNoteOwnersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) ListNotesByOwner(ctx context.Context, ownerType constants.OwnerType, ownerID model.ULID, params RelationListParams) ([]model.NoteOwner, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_notes_by_owner"
	if strings.TrimSpace(string(ownerID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "owner_id"))
	}
	if !isValidOwnerType(ownerType) {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "owner_type"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteOwnerRepo.ListByOwner(ctx, ownerType, ownerID, repo.ListNoteOwnersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) AddTag(ctx context.Context, params AddNoteTagParams) (model.NoteTag, error) {
	defer shared.LogServiceCall()()
	op := "note_service.add_tag"
	if err := validateTwoIDs(params.NoteID, "note_id", params.TagID, "tag_id"); err != nil {
		return model.NoteTag{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rel, err := s.noteTagRepo.Create(ctx, model.NoteTag{
		NoteID: params.NoteID,
		TagID:  params.TagID,
	})
	if err != nil {
		return model.NoteTag{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rel, nil
}

func (s *NoteService) RemoveTag(ctx context.Context, noteID, tagID model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.remove_tag"
	if err := validateTwoIDs(noteID, "note_id", tagID, "tag_id"); err != nil {
		return serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if err := s.noteTagRepo.Delete(ctx, noteID, tagID); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func (s *NoteService) GetTag(ctx context.Context, noteID, tagID model.ULID, includeDeleted bool) (model.NoteTag, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_tag"
	if err := validateTwoIDs(noteID, "note_id", tagID, "tag_id"); err != nil {
		return model.NoteTag{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rel, err := s.noteTagRepo.Get(ctx, noteID, tagID, includeDeleted)
	if err != nil {
		return model.NoteTag{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rel, nil
}

func (s *NoteService) ListTagsByNote(ctx context.Context, noteID model.ULID, params RelationListParams) ([]model.NoteTag, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_tags_by_note"
	if strings.TrimSpace(string(noteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteTagRepo.ListByNote(ctx, noteID, repo.ListNoteTagsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) ListNotesByTag(ctx context.Context, tagID model.ULID, params RelationListParams) ([]model.NoteTag, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_notes_by_tag"
	if strings.TrimSpace(string(tagID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "tag_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteTagRepo.ListByTag(ctx, tagID, repo.ListNoteTagsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) CreateLink(ctx context.Context, params CreateNoteLinkParams) (model.NoteLink, error) {
	defer shared.LogServiceCall()()
	op := "note_service.create_link"
	if err := validateTwoIDs(params.SourceNoteID, "source_note_id", params.TargetNoteID, "target_note_id"); err != nil {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if params.SourceNoteID == params.TargetNoteID {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDSMUSTBEDIFFERENTMESSAGE, "source_note_id", "target_note_id"))
	}
	if !isValidNoteLinkType(params.LinkType) {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "link_type"))
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.NoteLink{}, serviceerrors.WrapUnknown(op, noteServiceName, err)
	}
	link, repoErr := s.noteLinkRepo.Create(ctx, model.NoteLink{
		ID:           id,
		SourceNoteID: params.SourceNoteID,
		TargetNoteID: params.TargetNoteID,
		LinkType:     params.LinkType,
	})
	if repoErr != nil {
		return model.NoteLink{}, shared.MapRepositoryError(repoErr, op, noteServiceName)
	}
	return link, nil
}

func (s *NoteService) GetLinkByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteLink, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_link_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	link, err := s.noteLinkRepo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.NoteLink{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return link, nil
}

func (s *NoteService) ListLinksBySource(ctx context.Context, sourceNoteID model.ULID, params RelationListParams) ([]model.NoteLink, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_links_by_source"
	if strings.TrimSpace(string(sourceNoteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "source_note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteLinkRepo.ListBySource(ctx, sourceNoteID, repo.ListNoteLinksParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) ListLinksByTarget(ctx context.Context, targetNoteID model.ULID, params RelationListParams) ([]model.NoteLink, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_links_by_target"
	if strings.TrimSpace(string(targetNoteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "target_note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteLinkRepo.ListByTarget(ctx, targetNoteID, repo.ListNoteLinksParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) UpdateLink(ctx context.Context, params UpdateNoteLinkParams) (model.NoteLink, error) {
	defer shared.LogServiceCall()()
	op := "note_service.update_link"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	if !isValidNoteLinkType(params.LinkType) {
		return model.NoteLink{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "link_type"))
	}
	link, err := s.noteLinkRepo.Update(ctx, repo.UpdateNoteLinkParams{
		ID:              params.ID,
		LinkType:        params.LinkType,
		ExpectedVersion: params.ExpectedVersion,
	})
	if err != nil {
		return model.NoteLink{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return link, nil
}

func (s *NoteService) DeleteLink(ctx context.Context, id model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.delete_link"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.noteLinkRepo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func (s *NoteService) CreateAsset(ctx context.Context, params CreateNoteAssetParams) (model.NoteAsset, error) {
	defer shared.LogServiceCall()()
	op := "note_service.create_asset"
	if strings.TrimSpace(string(params.NoteID)) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "note_id"))
	}
	if strings.TrimSpace(params.StoragePath) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "storage_path"))
	}
	if strings.TrimSpace(params.MIMEType) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "mime_type"))
	}
	if !isValidAssetType(params.AssetType) {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "asset_type"))
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.NoteAsset{}, serviceerrors.WrapUnknown(op, noteServiceName, err)
	}
	now := s.clock.Now()
	asset, repoErr := s.noteAssetRepo.Create(ctx, model.NoteAsset{
		ID:          id,
		NoteID:      params.NoteID,
		AssetType:   params.AssetType,
		StoragePath: strings.TrimSpace(params.StoragePath),
		MIMEType:    strings.TrimSpace(params.MIMEType),
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.NoteAsset{}, shared.MapRepositoryError(repoErr, op, noteServiceName)
	}
	return asset, nil
}

func (s *NoteService) GetAssetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.NoteAsset, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_asset_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	asset, err := s.noteAssetRepo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.NoteAsset{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return asset, nil
}

func (s *NoteService) ListAssetsByNote(ctx context.Context, noteID model.ULID, params RelationListParams) ([]model.NoteAsset, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_assets_by_note"
	if strings.TrimSpace(string(noteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.noteAssetRepo.ListByNote(ctx, noteID, repo.ListNoteAssetsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) UpdateAsset(ctx context.Context, params UpdateNoteAssetParams) (model.NoteAsset, error) {
	defer shared.LogServiceCall()()
	op := "note_service.update_asset"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	if strings.TrimSpace(params.StoragePath) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "storage_path"))
	}
	if strings.TrimSpace(params.MIMEType) == "" {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "mime_type"))
	}
	if !isValidAssetType(params.AssetType) {
		return model.NoteAsset{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "asset_type"))
	}
	asset, err := s.noteAssetRepo.Update(ctx, repo.UpdateNoteAssetParams{
		ID:              params.ID,
		AssetType:       params.AssetType,
		StoragePath:     strings.TrimSpace(params.StoragePath),
		MIMEType:        strings.TrimSpace(params.MIMEType),
		ExpectedVersion: params.ExpectedVersion,
	})
	if err != nil {
		return model.NoteAsset{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return asset, nil
}

func (s *NoteService) DeleteAsset(ctx context.Context, id model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.delete_asset"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.noteAssetRepo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func (s *NoteService) UpsertMapPlacement(ctx context.Context, params UpsertMapNotePlacementParams) (model.MapNotePlacement, error) {
	defer shared.LogServiceCall()()
	op := "note_service.upsert_map_placement"
	if err := validateTwoIDs(params.MapNoteID, "map_note_id", params.TargetNoteID, "target_note_id"); err != nil {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	if params.X > 100 || params.Y > 100 {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVXYMAXRANGEMESSAGE))
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.MapNotePlacement{}, serviceerrors.WrapUnknown(op, noteServiceName, err)
	}
	placement, repoErr := s.mapNotePlacementRepo.Create(ctx, model.MapNotePlacement{
		ID:           id,
		MapNoteID:    params.MapNoteID,
		TargetNoteID: params.TargetNoteID,
		X:            params.X,
		Y:            params.Y,
	})
	if repoErr != nil {
		return model.MapNotePlacement{}, shared.MapRepositoryError(repoErr, op, noteServiceName)
	}
	return placement, nil
}

func (s *NoteService) GetMapPlacementByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.MapNotePlacement, error) {
	defer shared.LogServiceCall()()
	op := "note_service.get_map_placement_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	placement, err := s.mapNotePlacementRepo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.MapNotePlacement{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return placement, nil
}

func (s *NoteService) ListMapPlacementsByMap(ctx context.Context, mapNoteID model.ULID, params RelationListParams) ([]model.MapNotePlacement, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_map_placements_by_map"
	if strings.TrimSpace(string(mapNoteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "map_note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.mapNotePlacementRepo.ListByMapNote(ctx, mapNoteID, repo.ListMapNotePlacementsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) ListMapPlacementsByTarget(ctx context.Context, targetNoteID model.ULID, params RelationListParams) ([]model.MapNotePlacement, error) {
	defer shared.LogServiceCall()()
	op := "note_service.list_map_placements_by_target"
	if strings.TrimSpace(string(targetNoteID)) == "" {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "target_note_id"))
	}
	if err := validateOffsetLimit(params.Offset, params.Limit); err != nil {
		return nil, serviceerrors.WrapValidation(op, noteServiceName, err)
	}
	rows, err := s.mapNotePlacementRepo.ListByTargetNote(ctx, targetNoteID, repo.ListMapNotePlacementsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return rows, nil
}

func (s *NoteService) UpdateMapPlacement(ctx context.Context, params UpdateMapNotePlacementParams) (model.MapNotePlacement, error) {
	defer shared.LogServiceCall()()
	op := "note_service.update_map_placement"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	if params.X > 100 || params.Y > 100 {
		return model.MapNotePlacement{}, serviceerrors.WrapValidation(op, noteServiceName, errors.New(serviceerrors.SERVXYMAXRANGEMESSAGE))
	}
	placement, err := s.mapNotePlacementRepo.Update(ctx, repo.UpdateMapNotePlacementParams{
		ID:              params.ID,
		X:               params.X,
		Y:               params.Y,
		ExpectedVersion: params.ExpectedVersion,
	})
	if err != nil {
		return model.MapNotePlacement{}, shared.MapRepositoryError(err, op, noteServiceName)
	}
	return placement, nil
}

func (s *NoteService) DeleteMapPlacement(ctx context.Context, id model.ULID) error {
	defer shared.LogServiceCall()()
	op := "note_service.delete_map_placement"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, noteServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.mapNotePlacementRepo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, noteServiceName)
	}
	return nil
}

func toNoteListItems(rows []model.Note) []NoteListItem {
	out := make([]NoteListItem, 0, len(rows))
	for _, note := range rows {
		out = append(out, NoteListItem{
			ID:           note.ID,
			Title:        note.Title,
			ContentMD:    note.ContentMD,
			NoteType:     note.NoteType,
			MetadataJSON: note.MetadataJSON,
			CreatedAt:    note.AuditFields.CreatedAt,
			UpdatedAt:    note.AuditFields.UpdatedAt,
			DeletedAt:    note.AuditFields.DeletedAt,
			Version:      note.AuditFields.Version,
		})
	}
	return out
}

func validateOffsetLimit(offset, limit int32) error {
	if offset < 0 {
		return errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE)
	}
	if limit <= 0 {
		return errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE)
	}
	return nil
}

func validateTwoIDs(left model.ULID, leftName string, right model.ULID, rightName string) error {
	if strings.TrimSpace(string(left)) == "" {
		return fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, leftName)
	}
	if strings.TrimSpace(string(right)) == "" {
		return fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, rightName)
	}
	return nil
}

func validateNoteOwnerIDs(noteID model.ULID, ownerID model.ULID) error {
	return validateTwoIDs(noteID, "note_id", ownerID, "owner_id")
}

func isValidNoteType(noteType constants.NoteType) bool {
	switch noteType {
	case constants.General, constants.SummaryNote, constants.Map, constants.Character, constants.Location:
		return true
	default:
		return false
	}
}

func isValidOwnerType(ownerType constants.OwnerType) bool {
	switch ownerType {
	case constants.World, constants.Plane, constants.Campaign, constants.Session, constants.Player:
		return true
	default:
		return false
	}
}

func isValidNoteLinkType(linkType constants.NoteLinkType) bool {
	switch linkType {
	case constants.Related, constants.Contains, constants.Mentions, constants.DependsOn, constants.LocatedIn:
		return true
	default:
		return false
	}
}

func isValidAssetType(assetType constants.AssetType) bool {
	switch assetType {
	case constants.Image:
		return true
	default:
		return false
	}
}
