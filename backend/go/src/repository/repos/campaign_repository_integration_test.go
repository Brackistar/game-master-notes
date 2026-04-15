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

func TestCampaignRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
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

	worldRepo := repos.NewWorldRepository(tx)
	campaignRepo := repos.NewCampaignRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)
	plane := createIntegrationPlane(t, ctx, tx, "campaign-plane", now)

	world, err := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("campaign-world")),
		PlaneID:     plane.ID,
		Name:        "Campaign World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	startDate := now
	endDate := now.AddDate(0, 1, 0)
	campaignIn := model.Campaign{
		ID:        model.ULID(testULID("campaign-create")),
		WorldID:   world.ID,
		Name:      "Campaign One",
		StartDate: &startDate,
		EndDate:   &endDate,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	}

	created, err := campaignRepo.Create(ctx, campaignIn)
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	if created.ID != campaignIn.ID {
		t.Fatalf("created id mismatch: got=%s want=%s", created.ID, campaignIn.ID)
	}
	if created.WorldID != world.ID {
		t.Fatalf("created world id mismatch: got=%s want=%s", created.WorldID, world.ID)
	}

	got, err := campaignRepo.GetByID(ctx, created.ID, false)
	if err != nil {
		t.Fatalf("get campaign by id: %v", err)
	}
	if got.Name != "Campaign One" {
		t.Fatalf("get campaign name mismatch: got=%s", got.Name)
	}

	second := model.Campaign{
		ID:      model.ULID(testULID("campaign-list")),
		WorldID: world.ID,
		Name:    "Campaign Two",
		AuditFields: model.AuditFields{
			CreatedAt: now.Add(1 * time.Minute),
			UpdatedAt: now.Add(1 * time.Minute),
			Version:   1,
		},
	}
	if _, err := campaignRepo.Create(ctx, second); err != nil {
		t.Fatalf("create second campaign: %v", err)
	}

	list, err := campaignRepo.List(ctx, interfaces.ListCampaignsParams{
		Offset:         0,
		Limit:          10,
		IncludeDeleted: false,
	})
	if err != nil {
		t.Fatalf("list campaigns: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 campaigns in list, got %d", len(list))
	}

	newEnd := endDate.AddDate(0, 1, 0)
	updated, err := campaignRepo.Update(ctx, interfaces.UpdateCampaignParams{
		ID:              created.ID,
		Name:            "Campaign One Updated",
		StartDate:       &startDate,
		EndDate:         &newEnd,
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update campaign: %v", err)
	}
	if updated.Name != "Campaign One Updated" {
		t.Fatalf("updated name mismatch: got=%s", updated.Name)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment, got %d", updated.AuditFields.Version)
	}
}

func TestCampaignRepositoryIntegration_UpdateConflict(t *testing.T) {
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

	worldRepo := repos.NewWorldRepository(tx)
	campaignRepo := repos.NewCampaignRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)
	plane := createIntegrationPlane(t, ctx, tx, "campaign-conflict-plane", now)

	world, err := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("campaign-conflict-world")),
		PlaneID:     plane.ID,
		Name:        "Conflict World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	created, err := campaignRepo.Create(ctx, model.Campaign{
		ID:      model.ULID(testULID("campaign-conflict")),
		WorldID: world.ID,
		Name:    "Conflict Campaign",
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}

	_, err = campaignRepo.Update(ctx, interfaces.UpdateCampaignParams{
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

func TestCampaignRepositoryIntegration_DeleteStrictAndIncludeDeleted(t *testing.T) {
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

	worldRepo := repos.NewWorldRepository(tx)
	campaignRepo := repos.NewCampaignRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)
	plane := createIntegrationPlane(t, ctx, tx, "campaign-delete-plane", now)

	world, err := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("campaign-delete-world")),
		PlaneID:     plane.ID,
		Name:        "Delete World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	created, err := campaignRepo.Create(ctx, model.Campaign{
		ID:      model.ULID(testULID("campaign-delete")),
		WorldID: world.ID,
		Name:    "Delete Campaign",
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}

	if err := campaignRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete campaign: %v", err)
	}

	_, err = campaignRepo.GetByID(ctx, created.ID, false)
	if err == nil {
		t.Fatalf("expected not found for deleted campaign with includeDeleted=false")
	}
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}

	deletedView, err := campaignRepo.GetByID(ctx, created.ID, true)
	if err != nil {
		t.Fatalf("get deleted campaign includeDeleted=true: %v", err)
	}
	if deletedView.AuditFields.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}

	err = campaignRepo.Delete(ctx, created.ID)
	if err == nil {
		t.Fatalf("expected strict not found on second delete")
	}
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found error on second delete, got %v", err)
	}
}
