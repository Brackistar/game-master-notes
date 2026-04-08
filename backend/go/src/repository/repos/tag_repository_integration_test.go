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

func TestTagRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	worldRepo := repos.NewWorldRepository(tx)
	campaignRepo := repos.NewCampaignRepository(tx)
	tagRepo := repos.NewTagRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	world, _ := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("tag-world")),
		Name:        "Tag World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	campaign, _ := campaignRepo.Create(ctx, model.Campaign{
		ID:          model.ULID(testULID("tag-campaign")),
		WorldID:     world.ID,
		Name:        "Tag Campaign",
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})

	created, err := tagRepo.Create(ctx, model.Tag{
		ID:          model.ULID(testULID("tag-create")),
		Name:        "Important",
		CampaignID:  &campaign.ID,
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	got, err := tagRepo.GetByID(ctx, created.ID, false)
	if err != nil || got.Name != "Important" {
		t.Fatalf("get tag failed: err=%v name=%s", err, got.Name)
	}

	_, _ = tagRepo.Create(ctx, model.Tag{
		ID:          model.ULID(testULID("tag-list")),
		Name:        "GlobalTag",
		CampaignID:  nil,
		AuditFields: model.AuditFields{CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute), Version: 1},
	})

	list, err := tagRepo.List(ctx, interfaces.ListTagsParams{Offset: 0, Limit: 10, IncludeDeleted: false})
	if err != nil || len(list) != 2 {
		t.Fatalf("list tags failed: err=%v len=%d", err, len(list))
	}

	updated, err := tagRepo.Update(ctx, interfaces.UpdateTagParams{
		ID:              created.ID,
		Name:            "ImportantUpdated",
		CampaignID:      &campaign.ID,
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update tag: %v", err)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment")
	}

	if err := tagRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	_, err = tagRepo.GetByID(ctx, created.ID, false)
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

