package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	"github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const tagServiceName string = "tag"

type TagNamePolicy interface {
	NormalizeAndValidate(name string) (string, error)
}

type DefaultTagNamePolicy struct{}

func (DefaultTagNamePolicy) NormalizeAndValidate(name string) (string, error) {
	normalized := shared.NormalizeSpaces(name)
	if normalized == "" {
		return "", fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "name")
	}
	return normalized, nil
}

type CreateTagParams struct {
	Name       string
	CampaignID *model.ULID
}

type UpdateTagParams struct {
	ID              model.ULID
	Name            string
	CampaignID      *model.ULID
	ExpectedVersion model.Version
}

type ListTagsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type TagListItem struct {
	ID         model.ULID
	Name       string
	CampaignID *model.ULID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Version    model.Version
}

type TagService struct {
	repo        repo.TagRepository
	clock       shared.Clock
	namePolicy  TagNamePolicy
	idGenerator shared.IDGenerator
}

type TagServiceDeps struct {
	Repo        repo.TagRepository
	Clock       shared.Clock
	NamePolicy  TagNamePolicy
	IDGenerator shared.IDGenerator
}

func NewTagService(repo repo.TagRepository, idGenerator shared.IDGenerator) *TagService {
	return NewTagServiceWithDeps(TagServiceDeps{
		Repo:        repo,
		Clock:       shared.SystemClock{},
		NamePolicy:  DefaultTagNamePolicy{},
		IDGenerator: idGenerator,
	})
}

func NewTagServiceWithDeps(deps TagServiceDeps) *TagService {
	shared.PanicIfNilDependency(tagServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(tagServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(tagServiceName, "NamePolicy", deps.NamePolicy)
	shared.PanicIfNilDependency(tagServiceName, "IDGenerator", deps.IDGenerator)
	return &TagService{
		repo:        deps.Repo,
		clock:       deps.Clock,
		namePolicy:  deps.NamePolicy,
		idGenerator: deps.IDGenerator,
	}
}

func (s *TagService) Create(ctx context.Context, params CreateTagParams) (model.Tag, error) {
	op := "tag_service.create"
	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, err)
	}
	if params.CampaignID != nil && strings.TrimSpace(string(*params.CampaignID)) == "" {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, fmt.Errorf(serviceerrors.SERVFIELDCANNOTBEEMPTYMESSAGE, "campaign_id"))
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Tag{}, serviceerrors.WrapUnknown(op, tagServiceName, err)
	}
	now := s.clock.Now()
	tag, repoErr := s.repo.Create(ctx, model.Tag{
		ID:         id,
		Name:       name,
		CampaignID: params.CampaignID,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Tag{}, shared.MapRepositoryError(repoErr, op, tagServiceName)
	}
	return tag, nil
}

func (s *TagService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Tag, error) {
	op := "tag_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	tag, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Tag{}, shared.MapRepositoryError(err, op, tagServiceName)
	}
	return tag, nil
}

func (s *TagService) List(ctx context.Context, params ListTagsParams) ([]TagListItem, error) {
	op := "tag_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, tagServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, tagServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	rows, err := s.repo.List(ctx, repo.ListTagsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, tagServiceName)
	}
	return toTagListItems(rows), nil
}

func (s *TagService) Update(ctx context.Context, params UpdateTagParams) (model.Tag, error) {
	op := "tag_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, err)
	}
	if params.CampaignID != nil && strings.TrimSpace(string(*params.CampaignID)) == "" {
		return model.Tag{}, serviceerrors.WrapValidation(op, tagServiceName, fmt.Errorf(serviceerrors.SERVFIELDCANNOTBEEMPTYMESSAGE, "campaign_id"))
	}
	tag, repoErr := s.repo.Update(ctx, repo.UpdateTagParams{
		ID:              params.ID,
		Name:            name,
		CampaignID:      params.CampaignID,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Tag{}, shared.MapRepositoryError(repoErr, op, tagServiceName)
	}
	return tag, nil
}

func (s *TagService) Delete(ctx context.Context, id model.ULID) error {
	op := "tag_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, tagServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, tagServiceName)
	}
	return nil
}

func toTagListItems(rows []model.Tag) []TagListItem {
	out := make([]TagListItem, 0, len(rows))
	for _, tag := range rows {
		out = append(out, TagListItem{
			ID:         tag.ID,
			Name:       tag.Name,
			CampaignID: tag.CampaignID,
			CreatedAt:  tag.AuditFields.CreatedAt,
			UpdatedAt:  tag.AuditFields.UpdatedAt,
			DeletedAt:  tag.AuditFields.DeletedAt,
			Version:    tag.AuditFields.Version,
		})
	}
	return out
}
