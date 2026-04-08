//go:build integration

package repos_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/model"
	"github.com/Brackistar/game-master-notes/backend/go/src/model/constants"
	repoerror "github.com/Brackistar/game-master-notes/backend/go/src/repository/error"
	interfaces "github.com/Brackistar/game-master-notes/backend/go/src/repository/interfaces"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/repos"
	"github.com/jackc/pgx/v5"
)

func TestWorldRepositoryIntegration_CreateGetListUpdateDelete(t *testing.T) {
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

	repo := repos.NewWorldRepository(tx)

	now := time.Now().UTC().Truncate(time.Second)
	worldIn := model.World{
		ID:          model.ULID(testULID("world-create")),
		Name:        "World One",
		Description: "first world",
		Status:      constants.Draft,
		AuditFields: model.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
			Version:   1,
		},
	}

	created, err := repo.Create(ctx, worldIn)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	if created.ID != worldIn.ID {
		t.Fatalf("created id mismatch: got=%s want=%s", created.ID, worldIn.ID)
	}
	if created.Status != constants.Draft {
		t.Fatalf("created status mismatch: got=%v", created.Status)
	}

	got, err := repo.GetByID(ctx, worldIn.ID, false)
	if err != nil {
		t.Fatalf("get world by id: %v", err)
	}
	if got.Name != "World One" {
		t.Fatalf("get world name mismatch: got=%s", got.Name)
	}

	second := model.World{
		ID:          model.ULID(testULID("world-list")),
		Name:        "World Two",
		Description: "second world",
		Status:      constants.Active,
		AuditFields: model.AuditFields{
			CreatedAt: now.Add(1 * time.Minute),
			UpdatedAt: now.Add(1 * time.Minute),
			Version:   1,
		},
	}
	if _, err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second world: %v", err)
	}

	list, err := repo.List(ctx, interfaces.ListWorldsParams{
		Offset:         0,
		Limit:          10,
		IncludeDeleted: false,
	})
	if err != nil {
		t.Fatalf("list worlds: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 worlds in list, got %d", len(list))
	}

	updated, err := repo.Update(ctx, interfaces.UpdateWorldParams{
		ID:              created.ID,
		Name:            "World One Updated",
		Description:     "updated",
		Status:          constants.Archived,
		ExpectedVersion: created.AuditFields.Version,
	})
	if err != nil {
		t.Fatalf("update world: %v", err)
	}
	if updated.AuditFields.Version != created.AuditFields.Version+1 {
		t.Fatalf("expected version increment, got %d", updated.AuditFields.Version)
	}
	if updated.Status != constants.Archived {
		t.Fatalf("updated status mismatch: got=%v", updated.Status)
	}
}

func TestWorldRepositoryIntegration_UpdateConflict(t *testing.T) {
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

	repo := repos.NewWorldRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, model.World{
		ID:          model.ULID(testULID("world-conflict")),
		Name:        "Conflict World",
		Description: "conflict",
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

	_, err = repo.Update(ctx, interfaces.UpdateWorldParams{
		ID:              created.ID,
		Name:            "Invalid update",
		Description:     "stale version",
		Status:          constants.Active,
		ExpectedVersion: created.AuditFields.Version - 1,
	})
	if err == nil {
		t.Fatalf("expected update conflict error, got nil")
	}
	if !errors.Is(err, repoerror.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestWorldRepositoryIntegration_DeleteStrictAndIncludeDeleted(t *testing.T) {
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

	repo := repos.NewWorldRepository(tx)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := repo.Create(ctx, model.World{
		ID:          model.ULID(testULID("world-delete")),
		Name:        "Delete World",
		Description: "to delete",
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

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete world: %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID, false)
	if err == nil {
		t.Fatalf("expected not found for deleted world with includeDeleted=false")
	}
	if !errors.Is(err, repoerror.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}

	deletedView, err := repo.GetByID(ctx, created.ID, true)
	if err != nil {
		t.Fatalf("get deleted world includeDeleted=true: %v", err)
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

func openIntegrationConn(t *testing.T, ctx context.Context) *pgx.Conn {
	t.Helper()

	url := integrationDatabaseURL(t)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	var regclass *string
	err = conn.QueryRow(ctx, "SELECT to_regclass('public.worlds')::text").Scan(&regclass)
	if err != nil {
		_ = conn.Close(ctx)
		t.Fatalf("check schema: %v", err)
	}
	if regclass == nil || *regclass == "" {
		_ = conn.Close(ctx)
		t.Skip("worlds table not found; run migrations first")
	}

	return conn
}

func integrationDatabaseURL(t *testing.T) string {
	t.Helper()

	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		return v
	}

	if v, ok := readDatabaseURLFromDotEnv(".env"); ok {
		return v
	}
	if v, ok := readDatabaseURLFromDotEnv(filepath.Join("..", "..", "..", "..", ".env")); ok {
		return v
	}

	t.Skip("DATABASE_URL not set and .env not found")
	return ""
}

func readDatabaseURLFromDotEnv(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[0]) == "DATABASE_URL" {
			return strings.TrimSpace(parts[1]), true
		}
	}
	return "", false
}

func testULID(seed string) string {
	base := fmt.Sprintf("%026s", strings.ToUpper(strings.ReplaceAll(seed, "-", "")))
	if len(base) > 26 {
		return base[:26]
	}
	return base
}
