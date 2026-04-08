package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	shared "github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const (
	playerMinNameLen        = 3
	playerMaxNameLen        = 50
	serviceName      string = "player"
)

var playerAllowedNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 '\-]*[A-Za-z0-9]$`)

type Clock interface {
	Now() time.Time
}

type PlayerNamePolicy interface {
	NormalizeAndValidate(name string) (string, error)
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type DefaultPlayerNamePolicy struct{}

func (DefaultPlayerNamePolicy) NormalizeAndValidate(name string) (string, error) {
	normalized := normalizeSpaces(name)
	if len(normalized) < playerMinNameLen {
		return "", fmt.Errorf("name must be at least %d characters", playerMinNameLen)
	}
	if len(normalized) > playerMaxNameLen {
		return "", fmt.Errorf("name must be at most %d characters", playerMaxNameLen)
	}
	if !playerAllowedNamePattern.MatchString(normalized) {
		return "", errors.New("name contains unsupported characters")
	}
	return normalized, nil
}

type PlayerListItem struct {
	ID        model.ULID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Version   model.Version
}

type CreatePlayerParams struct {
	Name string
}

type UpdatePlayerParams struct {
	ID              model.ULID
	Name            string
	ExpectedVersion model.Version
}

type ListPlayersParams struct {
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type SearchPlayersParams struct {
	Query          string
	Offset         int32
	Limit          int32
	IncludeDeleted bool
}

type PlayerService struct {
	repo        repo.PlayerRepository
	clock       Clock
	namePolicy  PlayerNamePolicy
	idGenerator shared.IDGenerator
}

type PlayerServiceDeps struct {
	Repo        repo.PlayerRepository
	Clock       Clock
	NamePolicy  PlayerNamePolicy
	IDGenerator shared.IDGenerator
}

func NewPlayerService(repo repo.PlayerRepository, idGenerator shared.IDGenerator) *PlayerService {
	return NewPlayerServiceWithDeps(PlayerServiceDeps{
		Repo:        repo,
		Clock:       SystemClock{},
		NamePolicy:  DefaultPlayerNamePolicy{},
		IDGenerator: idGenerator,
	})
}

func NewPlayerServiceWithDeps(deps PlayerServiceDeps) *PlayerService {
	if deps.Repo == nil {
		panic(fmt.Sprintf(serviceerrors.SERVDEPNILMESSAGE, serviceName, "repo"))
	}
	if deps.Clock == nil {
		panic(fmt.Sprintf(serviceerrors.SERVDEPNILMESSAGE, serviceName, "Clock"))
	}
	if deps.NamePolicy == nil {
		panic(fmt.Sprintf(serviceerrors.SERVDEPNILMESSAGE, serviceName, "NamePolicy"))
	}
	if deps.IDGenerator == nil {
		panic(fmt.Sprintf(serviceerrors.SERVDEPNILMESSAGE, serviceName, "IDGenerator"))
	}
	return &PlayerService{
		repo:        deps.Repo,
		clock:       deps.Clock,
		namePolicy:  deps.NamePolicy,
		idGenerator: deps.IDGenerator,
	}
}

func (s *PlayerService) Create(ctx context.Context, params CreatePlayerParams) (model.Player, error) {
	op := "player_service.create"
	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, err)
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Player{}, serviceerrors.WrapUnknown(op, serviceName, err)
	}

	now := s.clock.Now()
	player, repoErr := s.repo.Create(ctx, model.Player{
		ID:   id,
		Name: name,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if repoErr != nil {
		return model.Player{}, mapRepositoryError(repoErr, op, serviceName)
	}
	return player, nil
}

func (s *PlayerService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error) {
	op := "player_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, errors.New("id is required"))
	}
	player, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Player{}, mapRepositoryError(err, op, serviceName)
	}
	return player, nil
}

func (s *PlayerService) List(ctx context.Context, params ListPlayersParams) ([]PlayerListItem, error) {
	op := "player_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, serviceName, errors.New("offset must be >= 0"))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, serviceName, errors.New("limit must be > 0"))
	}

	rows, err := s.repo.List(ctx, repo.ListPlayersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, mapRepositoryError(err, op, serviceName)
	}
	out := toPlayerListItems(rows)
	sortPlayerItems(out)
	return out, nil
}

func (s *PlayerService) SearchByName(ctx context.Context, params SearchPlayersParams) ([]PlayerListItem, error) {
	op := "player_service.search_by_name"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, serviceName, errors.New("offset must be >= 0"))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, serviceName, errors.New("limit must be > 0"))
	}
	query := normalizeSpaces(params.Query)
	if len(query) < playerMinNameLen {
		return nil, serviceerrors.WrapValidation(op, serviceName, fmt.Errorf("search query must be at least %d characters", playerMinNameLen))
	}

	rows, err := s.repo.SearchByName(ctx, repo.SearchPlayersParams{
		Query:          query,
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, mapRepositoryError(err, op, serviceName)
	}
	out := toPlayerListItems(rows)
	sortPlayerItems(out)
	return out, nil
}

func (s *PlayerService) Update(ctx context.Context, params UpdatePlayerParams) (model.Player, error) {
	op := "player_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, errors.New("id is required"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, errors.New("expected_version must be > 0"))
	}

	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, err)
	}

	player, repoErr := s.repo.Update(ctx, repo.UpdatePlayerParams{
		ID:              params.ID,
		Name:            name,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Player{}, mapRepositoryError(repoErr, op, serviceName)
	}
	return player, nil
}

func (s *PlayerService) Delete(ctx context.Context, id model.ULID) error {
	op := "player_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, serviceName, errors.New("id is required"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapRepositoryError(err, op, serviceName)
	}
	return nil
}

func (s *PlayerService) Restore(ctx context.Context, id model.ULID) (model.Player, error) {
	op := "player_service.restore"
	if strings.TrimSpace(string(id)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, serviceName, errors.New("id is required"))
	}

	current, err := s.repo.GetByID(ctx, id, true)
	if err != nil {
		return model.Player{}, mapRepositoryError(err, op, serviceName)
	}
	if current.AuditFields.DeletedAt == nil {
		return model.Player{}, serviceerrors.NewConflict(op, serviceName)
	}

	restored, err := s.repo.Restore(ctx, id)
	if err != nil {
		return model.Player{}, mapRepositoryError(err, op, serviceName)
	}
	return restored, nil
}

func normalizeSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func mapRepositoryError(err error, op string, entity string) error {
	switch {
	case errors.Is(err, repoerrors.ErrNotFound):
		return serviceerrors.NewNotFound(op, entity)
	case errors.Is(err, repoerrors.ErrConflict):
		return serviceerrors.NewConflict(op, entity)
	case errors.Is(err, repoerrors.ErrValidation):
		return serviceerrors.WrapValidation(op, entity, err)
	case errors.Is(err, repoerrors.ErrUnknown):
		return serviceerrors.WrapUnknown(op, entity, err)
	default:
		return serviceerrors.WrapUnknown(op, entity, err)
	}
}

func toPlayerListItems(players []model.Player) []PlayerListItem {
	out := make([]PlayerListItem, 0, len(players))
	for _, item := range players {
		out = append(out, PlayerListItem{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.AuditFields.CreatedAt,
			UpdatedAt: item.AuditFields.UpdatedAt,
			DeletedAt: item.AuditFields.DeletedAt,
			Version:   item.AuditFields.Version,
		})
	}
	return out
}

func sortPlayerItems(items []PlayerListItem) {
	sort.SliceStable(items, func(i, j int) bool {
		left := strings.ToLower(items[i].Name)
		right := strings.ToLower(items[j].Name)
		if left == right {
			return string(items[i].ID) < string(items[j].ID)
		}
		return left < right
	})
}
