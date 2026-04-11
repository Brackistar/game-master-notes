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

const campaignServiceName string = "campaign"

type CampaignNamePolicy interface {
	NormalizeAndValidate(name string) (string, error)
}

type DefaultCampaignNamePolicy struct{}

func (DefaultCampaignNamePolicy) NormalizeAndValidate(name string) (string, error) {
	normalized := shared.NormalizeSpaces(name)

	if normalized == "" {
		return "", fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "name")
	}

	return normalized, nil
}

type CreateCampaignParams struct {
	WorldID   model.ULID
	Name      string
	StartDate *time.Time
	EndDate   *time.Time
}

type UpdateCampaignParams struct {
	ID              model.ULID
	Name            string
	StartDate       *time.Time
	EndDate         *time.Time
	ExpectedVersion model.Version
}

type ListCampaignsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type CampaignListItem struct {
	ID        model.ULID
	WorldID   model.ULID
	Name      string
	StartDate *time.Time
	EndDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Version   model.Version
}

type CampaignService struct {
	repo               repo.CampaignRepository
	campaignPlayerRepo repo.CampaignPlayerRepository
	clock              shared.Clock
	namePolicy         CampaignNamePolicy
	idGenerator        shared.IDGenerator
}

type CampaignServiceDeps struct {
	Repo               repo.CampaignRepository
	CampaignPlayerRepo repo.CampaignPlayerRepository
	Clock              shared.Clock
	NamePolicy         CampaignNamePolicy
	IDGenerator        shared.IDGenerator
}

func NewCampaignService(
	repo repo.CampaignRepository,
	campaignPlayerRepo repo.CampaignPlayerRepository,
	idGenerator shared.IDGenerator,
) *CampaignService {
	return NewCampaignServiceWithDeps(CampaignServiceDeps{
		Repo:               repo,
		CampaignPlayerRepo: campaignPlayerRepo,
		Clock:              shared.SystemClock{},
		NamePolicy:         DefaultCampaignNamePolicy{},
		IDGenerator:        idGenerator,
	})
}

func NewCampaignServiceWithDeps(deps CampaignServiceDeps) *CampaignService {
	shared.PanicIfNilDependency(campaignServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(campaignServiceName, "CampaignPlayerRepo", deps.CampaignPlayerRepo)
	shared.PanicIfNilDependency(campaignServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(campaignServiceName, "NamePolicy", deps.NamePolicy)
	shared.PanicIfNilDependency(campaignServiceName, "IDGenerator", deps.IDGenerator)

	return &CampaignService{
		repo:               deps.Repo,
		campaignPlayerRepo: deps.CampaignPlayerRepo,
		clock:              deps.Clock,
		namePolicy:         deps.NamePolicy,
		idGenerator:        deps.IDGenerator,
	}
}

func (s *CampaignService) Create(ctx context.Context, params CreateCampaignParams) (model.Campaign, error) {
	op := "campaign_service.create"
	if strings.TrimSpace(string(params.WorldID)) == "" {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "world_id"))
	}

	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, err)
	}
	if err := validateCampaignDateRange(params.StartDate, params.EndDate); err != nil {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, err)
	}

	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Campaign{}, serviceerrors.WrapUnknown(op, campaignServiceName, err)
	}

	now := s.clock.Now()
	campaign, repoErr := s.repo.Create(ctx, model.Campaign{
		ID:        id,
		WorldID:   params.WorldID,
		Name:      name,
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Campaign{}, shared.MapRepositoryError(repoErr, op, campaignServiceName)
	}
	return campaign, nil
}

func (s *CampaignService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Campaign, error) {
	op := "campaign_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	campaign, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Campaign{}, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return campaign, nil
}

func (s *CampaignService) List(ctx context.Context, params ListCampaignsParams) ([]CampaignListItem, error) {
	op := "campaign_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}

	rows, err := s.repo.List(ctx, repo.ListCampaignsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return toCampaignListItems(rows), nil
}

func (s *CampaignService) Update(ctx context.Context, params UpdateCampaignParams) (model.Campaign, error) {
	op := "campaign_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}

	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, err)
	}
	if err := validateCampaignDateRange(params.StartDate, params.EndDate); err != nil {
		return model.Campaign{}, serviceerrors.WrapValidation(op, campaignServiceName, err)
	}

	campaign, repoErr := s.repo.Update(ctx, repo.UpdateCampaignParams{
		ID:              params.ID,
		Name:            name,
		StartDate:       params.StartDate,
		EndDate:         params.EndDate,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Campaign{}, shared.MapRepositoryError(repoErr, op, campaignServiceName)
	}
	return campaign, nil
}

func (s *CampaignService) Delete(ctx context.Context, id model.ULID) error {
	op := "campaign_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return nil
}

func (s *CampaignService) AddPlayer(ctx context.Context, campaignID, playerID model.ULID) (model.CampaignPlayer, error) {
	op := "campaign_service.add_player"
	if strings.TrimSpace(string(campaignID)) == "" {
		return model.CampaignPlayer{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "campaign_id"))
	}
	if strings.TrimSpace(string(playerID)) == "" {
		return model.CampaignPlayer{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "player_id"))
	}
	rel, err := s.campaignPlayerRepo.Create(ctx, model.CampaignPlayer{
		CampaignID: campaignID,
		PlayerID:   playerID,
	})
	if err != nil {
		return model.CampaignPlayer{}, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return rel, nil
}

func (s *CampaignService) RemovePlayer(ctx context.Context, campaignID, playerID model.ULID) error {
	op := "campaign_service.remove_player"
	if strings.TrimSpace(string(campaignID)) == "" {
		return serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "campaign_id"))
	}
	if strings.TrimSpace(string(playerID)) == "" {
		return serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "player_id"))
	}
	if err := s.campaignPlayerRepo.Delete(ctx, campaignID, playerID); err != nil {
		return shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return nil
}

func (s *CampaignService) GetPlayerRelation(ctx context.Context, campaignID, playerID model.ULID, includeDeleted bool) (model.CampaignPlayer, error) {
	op := "campaign_service.get_player_relation"
	if strings.TrimSpace(string(campaignID)) == "" {
		return model.CampaignPlayer{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "campaign_id"))
	}
	if strings.TrimSpace(string(playerID)) == "" {
		return model.CampaignPlayer{}, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "player_id"))
	}
	rel, err := s.campaignPlayerRepo.Get(ctx, campaignID, playerID, includeDeleted)
	if err != nil {
		return model.CampaignPlayer{}, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return rel, nil
}

func (s *CampaignService) ListPlayers(ctx context.Context, campaignID model.ULID, params ListCampaignsParams) ([]model.CampaignPlayer, error) {
	op := "campaign_service.list_players"
	if strings.TrimSpace(string(campaignID)) == "" {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "campaign_id"))
	}
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	rows, err := s.campaignPlayerRepo.ListByCampaign(ctx, campaignID, repo.ListCampaignPlayersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return rows, nil
}

func (s *CampaignService) ListCampaignsForPlayer(ctx context.Context, playerID model.ULID, params ListCampaignsParams) ([]model.CampaignPlayer, error) {
	op := "campaign_service.list_campaigns_for_player"
	if strings.TrimSpace(string(playerID)) == "" {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "player_id"))
	}
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, campaignServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	rows, err := s.campaignPlayerRepo.ListByPlayer(ctx, playerID, repo.ListCampaignPlayersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, campaignServiceName)
	}
	return rows, nil
}

func toCampaignListItems(campaigns []model.Campaign) []CampaignListItem {
	out := make([]CampaignListItem, 0, len(campaigns))
	for _, campaign := range campaigns {
		out = append(out, CampaignListItem{
			ID:        campaign.ID,
			WorldID:   campaign.WorldID,
			Name:      campaign.Name,
			StartDate: campaign.StartDate,
			EndDate:   campaign.EndDate,
			CreatedAt: campaign.AuditFields.CreatedAt,
			UpdatedAt: campaign.AuditFields.UpdatedAt,
			DeletedAt: campaign.AuditFields.DeletedAt,
			Version:   campaign.AuditFields.Version,
		})
	}
	return out
}

func validateCampaignDateRange(startDate, endDate *time.Time) error {
	if startDate == nil || endDate == nil {
		return nil
	}
	if endDate.Before(*startDate) {
		return errors.New(serviceerrors.SERVENDDATEGTESTARTDATEMESSAGE)
	}
	return nil
}
