package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	shared "github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const worldServiceName string = "world"

type WorldPolicy interface {
	NormalizeAndValidate(name, description string, status constants.WorldStatus) (string, string, error)
}

type DefaultWorldPolicy struct{}

func (DefaultWorldPolicy) NormalizeAndValidate(name, description string, status constants.WorldStatus) (string, string, error) {
	normalizedName := shared.NormalizeSpaces(name)
	if normalizedName == "" {
		return "", "", fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "name")
	}
	if !isValidWorldStatus(status) {
		return "", "", fmt.Errorf(serviceerrors.SERVINVALIDFIELDMESSAGE, "world_status")
	}
	return normalizedName, strings.TrimSpace(description), nil
}

type CreateWorldParams struct {
	Name        string
	Description string
	Status      constants.WorldStatus
}

type UpdateWorldParams struct {
	ID              model.ULID
	Name            string
	Description     string
	Status          constants.WorldStatus
	ExpectedVersion model.Version
}

type ListWorldsParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type WorldListItem struct {
	ID          model.ULID
	Name        string
	Description string
	Status      constants.WorldStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Version     model.Version
}

type WorldService struct {
	repo        repo.WorldRepository
	clock       shared.Clock
	policy      WorldPolicy
	idGenerator shared.IDGenerator
}

type WorldServiceDeps struct {
	Repo        repo.WorldRepository
	Clock       shared.Clock
	Policy      WorldPolicy
	IDGenerator shared.IDGenerator
}

func NewWorldService(repo repo.WorldRepository, idGenerator shared.IDGenerator) *WorldService {
	return NewWorldServiceWithDeps(WorldServiceDeps{
		Repo:        repo,
		Clock:       shared.SystemClock{},
		Policy:      DefaultWorldPolicy{},
		IDGenerator: idGenerator,
	})
}

func NewWorldServiceWithDeps(deps WorldServiceDeps) *WorldService {
	shared.PanicIfNilDependency(worldServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(worldServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(worldServiceName, "Policy", deps.Policy)
	shared.PanicIfNilDependency(worldServiceName, "IDGenerator", deps.IDGenerator)
	return &WorldService{
		repo:        deps.Repo,
		clock:       deps.Clock,
		policy:      deps.Policy,
		idGenerator: deps.IDGenerator,
	}
}

func (s *WorldService) Create(ctx context.Context, params CreateWorldParams) (model.World, error) {
	op := "world_service.create"
	name, description, err := s.policy.NormalizeAndValidate(params.Name, params.Description, params.Status)
	if err != nil {
		return model.World{}, serviceerrors.WrapValidation(op, worldServiceName, err)
	}

	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.World{}, serviceerrors.WrapUnknown(op, worldServiceName, err)
	}

	now := s.clock.Now()
	world, repoErr := s.repo.Create(ctx, model.World{
		ID:          id,
		Name:        name,
		Description: description,
		Status:      params.Status,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.World{}, shared.MapRepositoryError(repoErr, op, worldServiceName)
	}
	return world, nil
}

func (s *WorldService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.World, error) {
	op := "world_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.World{}, serviceerrors.WrapValidation(op, worldServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	world, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.World{}, shared.MapRepositoryError(err, op, worldServiceName)
	}
	return world, nil
}

func (s *WorldService) List(ctx context.Context, params ListWorldsParams) ([]WorldListItem, error) {
	op := "world_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, worldServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, worldServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}

	rows, err := s.repo.List(ctx, repo.ListWorldsParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, worldServiceName)
	}
	return toWorldListItems(rows), nil
}

func (s *WorldService) Update(ctx context.Context, params UpdateWorldParams) (model.World, error) {
	op := "world_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.World{}, serviceerrors.WrapValidation(op, worldServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.World{}, serviceerrors.WrapValidation(op, worldServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}

	name, description, err := s.policy.NormalizeAndValidate(params.Name, params.Description, params.Status)
	if err != nil {
		return model.World{}, serviceerrors.WrapValidation(op, worldServiceName, err)
	}

	world, repoErr := s.repo.Update(ctx, repo.UpdateWorldParams{
		ID:              params.ID,
		Name:            name,
		Description:     description,
		Status:          params.Status,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.World{}, shared.MapRepositoryError(repoErr, op, worldServiceName)
	}
	return world, nil
}

func (s *WorldService) Delete(ctx context.Context, id model.ULID) error {
	op := "world_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, worldServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, worldServiceName)
	}
	return nil
}

func toWorldListItems(worlds []model.World) []WorldListItem {
	out := make([]WorldListItem, 0, len(worlds))
	for _, world := range worlds {
		out = append(out, WorldListItem{
			ID:          world.ID,
			Name:        world.Name,
			Description: world.Description,
			Status:      world.Status,
			CreatedAt:   world.AuditFields.CreatedAt,
			UpdatedAt:   world.AuditFields.UpdatedAt,
			DeletedAt:   world.AuditFields.DeletedAt,
			Version:     world.AuditFields.Version,
		})
	}
	return out
}

func isValidWorldStatus(status constants.WorldStatus) bool {
	switch status {
	case constants.Draft, constants.Active, constants.Archived:
		return true
	default:
		return false
	}
}
