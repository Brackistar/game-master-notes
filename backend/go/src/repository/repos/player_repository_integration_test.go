//go:build integration

package repos_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	repoerror "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/repos"
)

func TestPlayerRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	repo := repos.NewPlayerRepository(tx)

	now := time.Now().UTC().Truncate(time.Second)
	playerIn := model.Player{
		ID:   model.ULID(testULID("player-create")),
		Name: "Alice",
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	}

	created, err := repo.Create(ctx, playerIn)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if created.ID != playerIn.ID {
		t.Fatalf("created id mismatch: got=%s want=%s", created.ID, playerIn.ID)
	}

	got, err := repo.GetByID(ctx, playerIn.ID, false)
	if err != nil {
		t.Fatalf("get player by id: %v", err)
	}
	if got.Name != "Alice" {
		t.Fatalf("get player name mismatch: got=%s", got.Name)
	}

	second := model.Player{
		ID:   model.ULID(testULID("player-list")),
		Name: "Bob",
		AuditFields: model.AuditFields{
			CreatedAt: now.Add(1 * time.Minute),
			UpdatedAt: now.Add(1 * time.Minute),
			Version:   1,
		},
	}
	if _, err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second player: %v", err)
	}

	list, err := repo.List(ctx, interfaces.ListPlayersParams{
		Offset:         0,
		Limit:          10,
		IncludeDeleted: false,
	})
	if err != nil {
		t.Fatalf("list players: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 players in list, got %d", len(list))
	}

	updated, err := repo.Update(ctx, interfaces.UpdatePlayerParams{
		ID:              created.ID,
		Name:            "Alice Updated",
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update player: %v", err)
	}
	if updated.Name != "Alice Updated" {
		t.Fatalf("updated player name mismatch: got=%s", updated.Name)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment, got %d", updated.AuditFields.Version)
	}
}

func TestPlayerRepositoryIntegration_UpdateConflict(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	repo := repos.NewPlayerRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, model.Player{
		ID:   model.ULID(testULID("player-conflict")),
		Name: "Conflict Player",
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	_, err = repo.Update(ctx, interfaces.UpdatePlayerParams{
		ID:              created.ID,
		Name:            "Stale Update",
		ExpectedVersion: created.AuditFields.Version - 1,
	})
	if err == nil {
		t.Fatalf("expected update conflict error, got nil")
	}
	if !errors.Is(err, repoerror.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestPlayerRepositoryIntegration_DeleteStrictAndIncludeDeleted(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	repo := repos.NewPlayerRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, model.Player{
		ID:   model.ULID(testULID("player-delete")),
		Name: "Delete Player",
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete player: %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID, false)
	if err == nil {
		t.Fatalf("expected not found for deleted player with includeDeleted=false")
	}
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}

	deletedView, err := repo.GetByID(ctx, created.ID, true)
	if err != nil {
		t.Fatalf("get deleted player includeDeleted=true: %v", err)
	}
	if deletedView.AuditFields.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}

	err = repo.Delete(ctx, created.ID)
	if err == nil {
		t.Fatalf("expected strict not found on second delete")
	}
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found error on second delete, got %v", err)
	}
}
