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

func TestNoteRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
	ctx := context.Background()
	conn := openIntegrationConn(t, ctx)
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	repo := repos.NewNoteRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, model.Note{
		ID:           model.ULID(testULID("note-create")),
		Title:        "Note One",
		ContentMD:    "content",
		NoteType:     constants.General,
		MetadataJSON: []byte(`{"k":"v"}`),
		AuditFields:  model.AuditFields{CreatedAt: now, UpdatedAt: now, Version: 1},
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID, false)
	if err != nil || got.Title != "Note One" {
		t.Fatalf("get note failed: err=%v title=%s", err, got.Title)
	}

	_, _ = repo.Create(ctx, model.Note{
		ID:           model.ULID(testULID("note-list")),
		Title:        "Note Two",
		ContentMD:    "content2",
		NoteType:     constants.Location,
		MetadataJSON: []byte(`{}`),
		AuditFields:  model.AuditFields{CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute), Version: 1},
	})

	list, err := repo.List(ctx, interfaces.ListNotesParams{Offset: 0, Limit: 10, IncludeDeleted: false})
	if err != nil || len(list) != 2 {
		t.Fatalf("list notes failed: err=%v len=%d", err, len(list))
	}

	updated, err := repo.Update(ctx, interfaces.UpdateNoteParams{
		ID:              created.ID,
		Title:           "Note Updated",
		ContentMD:       "updated",
		NoteType:        constants.Character,
		MetadataJSON:    []byte(`{"updated":true}`),
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update note: %v", err)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment")
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete note: %v", err)
	}
	_, err = repo.GetByID(ctx, created.ID, false)
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

