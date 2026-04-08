package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerrors "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/postgres/generated"
	helpers "github.com/Brackistar/game-master-notes/backend/go/src/repository/repos/shared"
	"github.com/jackc/pgx/v5"
)

var _ interfaces.PlayerRepository = (*PlayerRepository)(nil)

type PlayerRepository struct {
	queries *generated.Queries
	nowFn   func() time.Time
}

func NewPlayerRepository(db generated.DBTX) *PlayerRepository {
	return &PlayerRepository{
		queries: generated.New(db),
		nowFn:   helpers.NowFn,
	}
}

func (r *PlayerRepository) Create(ctx context.Context, player model.Player) (model.Player, error) {
	row, err := r.queries.CreatePlayer(ctx, generated.CreatePlayerParams{
		ID:        string(player.ID),
		Name:      player.Name,
		CreatedAt: toPgTimestamptz(player.AuditFields.CreatedAt),
		UpdatedAt: toPgTimestamptz(player.AuditFields.UpdatedAt),
		DeletedAt: toNullablePgTimestamptz(player.AuditFields.DeletedAt),
		Version:   int32(player.AuditFields.Version),
	})
	if err != nil {
		return model.Player{}, repoerrors.WrapUnknown("player.create", "player", err)
	}

	out, err := mapPlayerRow(row)
	if err != nil {
		return model.Player{}, repoerrors.WrapValidation("player.create", "player", err)
	}
	return out, nil
}

func (r *PlayerRepository) GetByID(ctx context.Context, id model.ULID, includeDeleted bool) (model.Player, error) {
	row, err := r.queries.GetPlayerByID(ctx, generated.GetPlayerByIDParams{
		ID:      string(id),
		Column2: includeDeleted,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Player{}, repoerrors.NewNotFound("player.get_by_id", "player")
		}
		return model.Player{}, repoerrors.WrapUnknown("player.get_by_id", "player", err)
	}

	out, err := mapPlayerRow(row)
	if err != nil {
		return model.Player{}, repoerrors.WrapValidation("player.get_by_id", "player", err)
	}
	return out, nil
}

func (r *PlayerRepository) List(ctx context.Context, params interfaces.ListPlayersParams) ([]model.Player, error) {
	rows, err := r.queries.ListPlayers(ctx, generated.ListPlayersParams{
		Column1: params.IncludeDeleted,
		Offset:  params.Offset,
		Limit:   params.Limit,
	})
	if err != nil {
		return nil, repoerrors.WrapUnknown("player.list", "player", err)
	}

	players := make([]model.Player, 0, len(rows))
	for _, row := range rows {
		p, mapErr := mapPlayerRow(row)
		if mapErr != nil {
			return nil, repoerrors.WrapValidation("player.list", "player", mapErr)
		}
		players = append(players, p)
	}
	return players, nil
}

func (r *PlayerRepository) Update(ctx context.Context, params interfaces.UpdatePlayerParams) (model.Player, error) {
	row, err := r.queries.UpdatePlayer(ctx, generated.UpdatePlayerParams{
		ID:        string(params.ID),
		Name:      params.Name,
		UpdatedAt: toPgTimestamptz(r.nowFn()),
		Version:   int32(params.ExpectedVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Player{}, repoerrors.NewConflict("player.update", "player")
		}
		return model.Player{}, repoerrors.WrapUnknown("player.update", "player", err)
	}

	out, err := mapPlayerRow(row)
	if err != nil {
		return model.Player{}, repoerrors.WrapValidation("player.update", "player", err)
	}
	return out, nil
}

func (r *PlayerRepository) Delete(ctx context.Context, id model.ULID) error {
	affected, err := r.queries.DeletePlayer(ctx, generated.DeletePlayerParams{
		ID:        string(id),
		DeletedAt: toPgTimestamptz(r.nowFn()),
	})
	if err != nil {
		return repoerrors.WrapUnknown("player.delete", "player", err)
	}
	if affected == 0 {
		return repoerrors.NewNotFound("player.delete", "player")
	}
	return nil
}

func mapPlayerRow(row generated.Player) (model.Player, error) {
	return model.Player{
		ID:   model.ULID(fmt.Sprint(row.ID)),
		Name: row.Name,
		AuditFields: model.AuditFields{
			CreatedAt: fromPgTimestamptzOrZero(row.CreatedAt),
			UpdatedAt: fromPgTimestamptzOrZero(row.UpdatedAt),
			DeletedAt: fromNullablePgTimestamptz(row.DeletedAt),
			Version:   model.Version(row.Version),
		},
	}, nil
}
