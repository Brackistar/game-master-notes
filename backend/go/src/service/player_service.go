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
	repo "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	serviceerrors "github.com/Brackistar/game-master-notes/backend/go/src/service/error"
	shared "github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
)

const (
	playerMinNameLen         = 3
	playerMaxNameLen         = 50
	playerServiceName string = "player"
)

var playerAllowedNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 '\-]*[A-Za-z0-9]$`)

type PlayerNamePolicy interface {
	NormalizeAndValidate(name string) (string, error)
}

type DefaultPlayerNamePolicy struct{}

func (DefaultPlayerNamePolicy) NormalizeAndValidate(name string) (string, error) {
	normalized := shared.NormalizeSpaces(name)
	if len(normalized) < playerMinNameLen {
		return "", fmt.Errorf(serviceerrors.SERVFIELDMINCHARSMESSAGE, "name", playerMinNameLen)
	}
	if len(normalized) > playerMaxNameLen {
		return "", fmt.Errorf(serviceerrors.SERVFIELDMAXCHARSMESSAGE, "name", playerMaxNameLen)
	}
	if !playerAllowedNamePattern.MatchString(normalized) {
		return "", errors.New(serviceerrors.SERVNAMEUNSUPPORTEDCHARSMESSAGE)
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
	clock       shared.Clock
	namePolicy  PlayerNamePolicy
	idGenerator shared.IDGenerator
}

type PlayerServiceDeps struct {
	Repo        repo.PlayerRepository
	Clock       shared.Clock
	NamePolicy  PlayerNamePolicy
	IDGenerator shared.IDGenerator
}

func NewPlayerService(repo repo.PlayerRepository, idGenerator shared.IDGenerator) *PlayerService {
	return NewPlayerServiceWithDeps(PlayerServiceDeps{
		Repo:        repo,
		Clock:       shared.SystemClock{},
		NamePolicy:  DefaultPlayerNamePolicy{},
		IDGenerator: idGenerator,
	})
}

func NewPlayerServiceWithDeps(deps PlayerServiceDeps) *PlayerService {
	shared.PanicIfNilDependency(playerServiceName, "repo", deps.Repo)
	shared.PanicIfNilDependency(playerServiceName, "Clock", deps.Clock)
	shared.PanicIfNilDependency(playerServiceName, "NamePolicy", deps.NamePolicy)
	shared.PanicIfNilDependency(playerServiceName, "IDGenerator", deps.IDGenerator)

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
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, err)
	}
	id, err := s.idGenerator.NewULID()
	if err != nil {
		return model.Player{}, serviceerrors.WrapUnknown(op, playerServiceName, err)
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
		return model.Player{}, shared.MapRepositoryError(repoErr, op, playerServiceName)
	}
	return player, nil
}

func (s *PlayerService) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error) {
	op := "player_service.get_by_id"
	if strings.TrimSpace(string(id)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	player, err := s.repo.GetByID(ctx, id, includeDeleted)
	if err != nil {
		return model.Player{}, shared.MapRepositoryError(err, op, playerServiceName)
	}
	return player, nil
}

func (s *PlayerService) List(ctx context.Context, params ListPlayersParams) ([]PlayerListItem, error) {
	op := "player_service.list"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, playerServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, playerServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}

	rows, err := s.repo.List(ctx, repo.ListPlayersParams{
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, playerServiceName)
	}
	out := toPlayerListItems(rows)
	sortPlayerItems(out)
	return out, nil
}

func (s *PlayerService) SearchByName(ctx context.Context, params SearchPlayersParams) ([]PlayerListItem, error) {
	op := "player_service.search_by_name"
	if params.Offset < 0 {
		return nil, serviceerrors.WrapValidation(op, playerServiceName, errors.New(serviceerrors.SERVOFFSETGTEZEROMESSAGE))
	}
	if params.Limit <= 0 {
		return nil, serviceerrors.WrapValidation(op, playerServiceName, errors.New(serviceerrors.SERVLIMITGTZEROMESSAGE))
	}
	query := shared.NormalizeSpaces(params.Query)
	if len(query) < playerMinNameLen {
		return nil, serviceerrors.WrapValidation(op, playerServiceName, fmt.Errorf(serviceerrors.SERVFIELDMINCHARSMESSAGE, "search query", playerMinNameLen))
	}

	rows, err := s.repo.SearchByName(ctx, repo.SearchPlayersParams{
		Query:          query,
		Offset:         params.Offset,
		Limit:          params.Limit,
		IncludeDeleted: params.IncludeDeleted,
	})
	if err != nil {
		return nil, shared.MapRepositoryError(err, op, playerServiceName)
	}
	out := toPlayerListItems(rows)
	sortPlayerItems(out)
	return out, nil
}

func (s *PlayerService) Update(ctx context.Context, params UpdatePlayerParams) (model.Player, error) {
	op := "player_service.update"
	if strings.TrimSpace(string(params.ID)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if params.ExpectedVersion <= 0 {
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, errors.New(serviceerrors.SERVEXPECTEDVERSIONGTZEROMESSAGE))
	}

	name, err := s.namePolicy.NormalizeAndValidate(params.Name)
	if err != nil {
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, err)
	}

	player, repoErr := s.repo.Update(ctx, repo.UpdatePlayerParams{
		ID:              params.ID,
		Name:            name,
		ExpectedVersion: params.ExpectedVersion,
	})
	if repoErr != nil {
		return model.Player{}, shared.MapRepositoryError(repoErr, op, playerServiceName)
	}
	return player, nil
}

func (s *PlayerService) Delete(ctx context.Context, id model.ULID) error {
	op := "player_service.delete"
	if strings.TrimSpace(string(id)) == "" {
		return serviceerrors.WrapValidation(op, playerServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return shared.MapRepositoryError(err, op, playerServiceName)
	}
	return nil
}

func (s *PlayerService) Restore(ctx context.Context, id model.ULID) (model.Player, error) {
	op := "player_service.restore"
	if strings.TrimSpace(string(id)) == "" {
		return model.Player{}, serviceerrors.WrapValidation(op, playerServiceName, fmt.Errorf(serviceerrors.SERVFIELDREQUIREDMESSAGE, "id"))
	}

	current, err := s.repo.GetByID(ctx, id, true)
	if err != nil {
		return model.Player{}, shared.MapRepositoryError(err, op, playerServiceName)
	}
	if current.AuditFields.DeletedAt == nil {
		return model.Player{}, serviceerrors.NewConflict(op, playerServiceName)
	}

	restored, err := s.repo.Restore(ctx, id)
	if err != nil {
		return model.Player{}, shared.MapRepositoryError(err, op, playerServiceName)
	}
	return restored, nil
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
