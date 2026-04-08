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

func TestCampaignPlayerRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	worldRepo := repos.NewWorldRepository(tx)
	campaignRepo := repos.NewCampaignRepository(tx)
	playerRepo := repos.NewPlayerRepository(tx)
	relRepo := repos.NewCampaignPlayerRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	world, _ := worldRepo.Create(ctx, model.World{ID: model.ULID(testULID("cp-world")), Name: "w", Description: "d", Status: constants.Active, AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	campaign, _ := campaignRepo.Create(ctx, model.Campaign{ID: model.ULID(testULID("cp-campaign")), WorldID: world.ID, Name: "c", AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	player, _ := playerRepo.Create(ctx, model.Player{ID: model.ULID(testULID("cp-player")), Name: "p", AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})

	created, err := relRepo.Create(ctx, model.CampaignPlayer{CampaignID: campaign.ID, PlayerID: player.ID, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("create campaign_player: %v", err)
	}
	_, err = relRepo.Get(ctx, created.CampaignID, created.PlayerID, false)
	if err != nil {
		t.Fatalf("get campaign_player: %v", err)
	}
	if err := relRepo.Delete(ctx, created.CampaignID, created.PlayerID); err != nil {
		t.Fatalf("delete campaign_player: %v", err)
	}
	_, err = relRepo.Get(ctx, created.CampaignID, created.PlayerID, false)
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestNoteOwnerRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	noteRepo := repos.NewNoteRepository(tx)
	ownerRepo := repos.NewNoteOwnerRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	note, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("no-note")), Title: "n", ContentMD: "c", NoteType: constants.General, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})

	created, err := ownerRepo.Create(ctx, model.NoteOwner{NoteID: note.ID, OwnerType: constants.World, OwnerID: model.ULID(testULID("no-owner")), CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("create note_owner: %v", err)
	}
	_, err = ownerRepo.Get(ctx, created.NoteID, created.OwnerType, created.OwnerID, false)
	if err != nil {
		t.Fatalf("get note_owner: %v", err)
	}
	if err := ownerRepo.Delete(ctx, created.NoteID, created.OwnerType, created.OwnerID); err != nil {
		t.Fatalf("delete note_owner: %v", err)
	}
}

func TestNoteTagRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	noteRepo := repos.NewNoteRepository(tx)
	tagRepo := repos.NewTagRepository(tx)
	relRepo := repos.NewNoteTagRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	note, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("nt-note")), Title: "n", ContentMD: "c", NoteType: constants.General, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	tag, _ := tagRepo.Create(ctx, model.Tag{ID: model.ULID(testULID("nt-tag")), Name: "t", AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})

	created, err := relRepo.Create(ctx, model.NoteTag{NoteID: note.ID, TagID: tag.ID, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("create note_tag: %v", err)
	}
	_, err = relRepo.Get(ctx, created.NoteID, created.TagID, false)
	if err != nil {
		t.Fatalf("get note_tag: %v", err)
	}
	if err := relRepo.Delete(ctx, created.NoteID, created.TagID); err != nil {
		t.Fatalf("delete note_tag: %v", err)
	}
}

func TestNoteAssetRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	noteRepo := repos.NewNoteRepository(tx)
	assetRepo := repos.NewNoteAssetRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	note, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("na-note")), Title: "n", ContentMD: "c", NoteType: constants.Map, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	created, err := assetRepo.Create(ctx, model.NoteAsset{
		ID:          model.ULID(testULID("na-asset")),
		NoteID:      note.ID,
		AssetType:   constants.Image,
		StoragePath: "/maps/a.png",
		MIMEType:    "image/png",
		AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create note_asset: %v", err)
	}
	_, err = assetRepo.Update(ctx, interfaces.UpdateNoteAssetParams{
		ID:              created.ID,
		AssetType:       constants.Image,
		StoragePath:     "/maps/b.png",
		MIMEType:        "image/png",
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update note_asset: %v", err)
	}
}

func TestMapNotePlacementRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	noteRepo := repos.NewNoteRepository(tx)
	placementRepo := repos.NewMapNotePlacementRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	mapNote, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("mp-map")), Title: "map", ContentMD: "", NoteType: constants.Map, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	target, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("mp-target")), Title: "target", ContentMD: "", NoteType: constants.Location, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second), Version: 1}})

	created, err := placementRepo.Create(ctx, model.MapNotePlacement{
		ID:           model.ULID(testULID("mp-placement")),
		MapNoteID:    mapNote.ID,
		TargetNoteID: target.ID,
		X:            10,
		Y:            20,
		AuditFields:  model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create map_note_placement: %v", err)
	}
	_, err = placementRepo.Update(ctx, interfaces.UpdateMapNotePlacementParams{
		ID:              created.ID,
		X:               30,
		Y:               40,
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update map_note_placement: %v", err)
	}
}

func TestNoteLinkRepositoryIntegration_BasicFlow(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)
	tx, _ := conn.Begin(ctx)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	noteRepo := repos.NewNoteRepository(tx)
	linkRepo := repos.NewNoteLinkRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	source, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("nl-source")), Title: "s", ContentMD: "", NoteType: constants.General, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1}})
	target, _ := noteRepo.Create(ctx, model.Note{ID: model.ULID(testULID("nl-target")), Title: "t", ContentMD: "", NoteType: constants.General, MetadataJSON: []byte(`{}`), AuditFields: model.AuditFields{CreatedAt: now.Add(time.Second), UpdatedAt: now.Add(time.Second), Version: 1}})

	created, err := linkRepo.Create(ctx, model.NoteLink{
		ID:           model.ULID(testULID("nl-link")),
		SourceNoteID: source.ID,
		TargetNoteID: target.ID,
		LinkType:     constants.Related,
		AuditFields:  model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create note_link: %v", err)
	}
	_, err = linkRepo.Update(ctx, interfaces.UpdateNoteLinkParams{
		ID:              created.ID,
		LinkType:        constants.Mentions,
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update note_link: %v", err)
	}
}

