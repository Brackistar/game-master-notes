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

func TestSessionRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
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
	sessionRepo := repos.NewSessionRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	world, _ := worldRepo.Create(ctx, model.World{
		ID:          model.ULID(testULID("session-world")),
		Name:        "Session World",
		Description: "w",
		Status:      constants.Active,
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	campaign, _ := campaignRepo.Create(ctx, model.Campaign{
		ID:          model.ULID(testULID("session-campaign")),
		WorldID:     world.ID,
		Name:        "Session Campaign",
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})

	playedOn := now
	created, err := sessionRepo.Create(ctx, model.Session{
		ID:         model.ULID(testULID("session-create")),
		CampaignID: campaign.ID,
		PlayedOn:   &playedOn,
		SummaryMD:  "summary",
		AuditFields: model.AuditFields{
			CreatedAt: now, UpdatedAt: now, Version: 1,
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	got, err := sessionRepo.GetByID(ctx, created.ID, false)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.SummaryMD != "summary" {
		t.Fatalf("unexpected summary")
	}

	_, _ = sessionRepo.Create(ctx, model.Session{
		ID:         model.ULID(testULID("session-list")),
		CampaignID: campaign.ID,
		SummaryMD:  "s2",
		AuditFields: model.AuditFields{
			CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute), Version: 1,
		},
	})

	list, err := sessionRepo.List(ctx, interfaces.ListSessionsParams{Offset: 0, Limit: 10, IncludeDeleted: false})
	if err != nil || len(list) != 2 {
		t.Fatalf("list sessions failed: err=%v len=%d", err, len(list))
	}

	newPlayed := now.AddDate(0, 0, 1)
	updated, err := sessionRepo.Update(ctx, interfaces.UpdateSessionParams{
		ID:              created.ID,
		PlayedOn:        &newPlayed,
		SummaryMD:       "updated",
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update session: %v", err)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment")
	}

	if err := sessionRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	_, err = sessionRepo.GetByID(ctx, created.ID, false)
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}
