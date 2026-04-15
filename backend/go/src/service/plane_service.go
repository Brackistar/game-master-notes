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

const planeServiceName string = "plane"

type PlanePolicy interface {
	NormalizeAndValidate(name, description string) (string, string, error)
}

type DefaultPlanePolicy struct{}

func (DefaultPlanePolicy) NormalizeAndValidate(name, description string) (string, string, error) {
	normalizedName := shared.NormalizeSpaces(name)
	if normalizedName == "" {
		return "", "", fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "name")
	}
	return normalizedName, strings.TrimSpace(description), nil
}

type CreatePlaneParams struct {
	Name        string
	Description string
}

type UpdatePlaneParams struct {
	ID              model.ULID
	Name            string
	Description     string
	ExpectedVersion model.Version
}

type ListPlanesParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type PlaneListItem struct {
	ID          model.ULID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	Version     model.Version
}

type PlaneService struct {
	repo        repo.PlaneRepository
	clock       shared.Clock
	policy      PlanePolicy
	idGenerator shared.IDGenerator
}

type PlaneServiceDeps struct {
	Repo        repo.PlaneRepository
	Clock       shared.Clock
	Policy      PlanePolicy
	IDGenerator shared.IDGenerator
}

func NewPlaneService(repo repo.PlaneRepository, idGenerator shared.IDGenerator) *PlaneService {
	return NewPlaneServiceWithDeps(PlaneServiceDeps{
		Repo:        repo,
		Clock:       shared.SystemClock{},
		Policy:      DefaultPlanePolicy{},
		IDGenerator: idGenerator,
	})
}

func NewPlaneServiceWithDeps(deps PlaneServiceDeps) *PlaneService {
	shared.PanicIfNilDependency(planeServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(planeServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(planeServiceName, "Policy", deps.Policy)
	shared.PanicIfNilDependency(planeServiceName, "IDGenerator", deps.IDGenerator)
	return &PlaneService{
		repo:        deps.Repo,
		clock:       deps.Clock,
		policy:      deps.Policy,
		idGenerator: deps.IDGenerator,
	}
}

func (s *PlaneService) Create(ctx context.Context, params CreatePlaneParams) (model.Plane, error) {
	defer shared.LogServiceCall()()
	op := "plane_service.create"
	name, description, err := s.policy.NormalizeAndValidate(params.Name, params.Description)
	if err != nil {
		return model.Plane{}, serviceerrors.WrapValidation(op, planeServiceName, err)
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Plane{}, serviceerrors.WrapUnknown(op, planeServiceName, err)
	}

	now := s.clock.Now()
	plane, repoErr := s.repo.Create(ctx, model.Plane{
		ID:          id,
		Name:        name,
		Description: description,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Plane{}, shared.MapRepositoryError(repoErr, op, planeServiceName)
	}
	return plane, nil
}

func (s *PlaneService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Plane, error) {
	defer shared.LogServiceCall()()
	op := "plane_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Plane{}, serviceerrors.WrapValidation(op, planeServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	plane, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Plane{}, shared.MapRepositoryError(err, op, planeServiceName)
	}
	return plane, nil
}

func (s *PlaneService) List(ctx context.Context, params ListPlanesParams) ([]PlaneListItem, error) {
	defer shared.LogServiceCall()()
	op := "plane_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, planeServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, planeServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	rows, err := s.repo.List(ctx, repo.ListPlanesParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, planeServiceName)
	}
	return toPlaneListItems(rows), nil
}

func (s *PlaneService) Update(ctx context.Context, params UpdatePlaneParams) (model.Plane, error) {
	defer shared.LogServiceCall()()
	op := "plane_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Plane{}, serviceerrors.WrapValidation(op, planeServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Plane{}, serviceerrors.WrapValidation(op, planeServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}
	name, description, err := s.policy.NormalizeAndValidate(params.Name, params.Description)
	if err != nil {
		return model.Plane{}, serviceerrors.WrapValidation(op, planeServiceName, err)
	}
	plane, repoErr := s.repo.Update(ctx, repo.UpdatePlaneParams{
		ID:              params.ID,
		Name:            name,
		Description:     description,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Plane{}, shared.MapRepositoryError(repoErr, op, planeServiceName)
	}
	return plane, nil
}

func (s *PlaneService) Delete(ctx context.Context, id model.ULID) error {
	defer shared.LogServiceCall()()
	op := "plane_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, planeServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, planeServiceName)
	}
	return nil
}

func toPlaneListItems(planes []model.Plane) []PlaneListItem {
	out := make([]PlaneListItem, 0, len(planes))
	for _, plane := range planes {
		out = append(out, PlaneListItem{
			ID:          plane.ID,
			Name:        plane.Name,
			Description: plane.Description,
			CreatedAt:   plane.AuditFields.CreatedAt,
			UpdatedAt:   plane.AuditFields.UpdatedAt,
			DeletedAt:   plane.AuditFields.DeletedAt,
			Version:     plane.AuditFields.Version,
		})
	}
	return out
}
