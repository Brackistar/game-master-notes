//go:build integration

package repos_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repoerror "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/repos"
)

func TestPlaneRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	worldRepo := repos.NewWorldRepository(tx)
	planeRepo := repos.NewPlaneRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	world, err := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("plane-world")),
		Name:        "Plane World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	created, err := planeRepo.Create(ctx, model.Plane{
		ID:          model.ULID(testULID("plane-create")),
		WorldID:     world.ID,
		Name:        "Plane One",
		Description: "desc",
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create plane: %v", err)
	}

	got, err := planeRepo.GetByID(ctx, created.ID, false)
	if err != nil {
		t.Fatalf("get plane: %v", err)
	}
	if got.Name != "Plane One" {
		t.Fatalf("unexpected plane name: %s", got.Name)
	}

	_, _ = planeRepo.Create(ctx, model.Plane{
		ID:          model.ULID(testULID("plane-list")),
		WorldID:     world.ID,
		Name:        "Plane Two",
		Description: "desc2",
		AuditFields: model.AuditFields{CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute), Version: 1},
	})

	list, err := planeRepo.List(ctx, interfaces.ListPlanesParams{Offset: 0, Limit: 10, IncludeDeleted: false})
	if err != nil {
		t.Fatalf("list planes: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 planes, got %d", len(list))
	}

	updated, err := planeRepo.Update(ctx, interfaces.UpdatePlaneParams{
		ID:              created.ID,
		Name:            "Plane Updated",
		Description:     "updated",
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update plane: %v", err)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment")
	}

	if err := planeRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete plane: %v", err)
	}
	_, err = planeRepo.GetByID(ctx, created.ID, false)
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
