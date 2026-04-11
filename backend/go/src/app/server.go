package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Brackistar/game-master-notes/backend/go/src/api"
	"github.com/Brackistar/game-master-notes/backend/go/src/repository/repos"
	"github.com/Brackistar/game-master-notes/backend/go/src/service"
	serviceshared "github.com/Brackistar/game-master-notes/backend/go/src/service/shared"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultPort     = "8080"
	serverReadTO    = 10 * time.Second
	serverWriteTO   = 15 * time.Second
	serverIdleTO    = 60 * time.Second
	serverShutdownT = 10 * time.Second
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	idGenerator := serviceshared.NewOklogULIDGenerator()

	worldRepo := repos.NewWorldRepository(pool)
	playerRepo := repos.NewPlayerRepository(pool)
	campaignRepo := repos.NewCampaignRepository(pool)
	campaignPlayerRepo := repos.NewCampaignPlayerRepository(pool)
	planeRepo := repos.NewPlaneRepository(pool)
	sessionRepo := repos.NewSessionRepository(pool)
	tagRepo := repos.NewTagRepository(pool)
	noteRepo := repos.NewNoteRepository(pool)
	noteOwnerRepo := repos.NewNoteOwnerRepository(pool)
	noteTagRepo := repos.NewNoteTagRepository(pool)
	noteLinkRepo := repos.NewNoteLinkRepository(pool)
	noteAssetRepo := repos.NewNoteAssetRepository(pool)
	mapPlacementRepo := repos.NewMapNotePlacementRepository(pool)

	worldService := service.NewWorldService(worldRepo, idGenerator)
	playerService := service.NewPlayerService(playerRepo, idGenerator)
	campaignService := service.NewCampaignService(campaignRepo, campaignPlayerRepo, idGenerator)
	planeService := service.NewPlaneService(planeRepo, idGenerator)
	sessionService := service.NewSessionService(sessionRepo, idGenerator)
	tagService := service.NewTagService(tagRepo, idGenerator)
	noteService := service.NewNoteService(noteRepo, noteOwnerRepo, noteTagRepo, noteLinkRepo, noteAssetRepo, mapPlacementRepo, idGenerator)

	worldAPI := api.NewWorldAPI(worldService)
	playerAPI := api.NewPlayerAPI(playerService)
	campaignAPI := api.NewCampaignAPI(campaignService)
	planeAPI := api.NewPlaneAPI(planeService)
	sessionAPI := api.NewSessionAPI(sessionService)
	tagAPI := api.NewTagAPI(tagService)
	noteAPI := api.NewNoteAPI(noteService)

	mux := http.NewServeMux()
	registerMetaRoutes(mux)
	registerSwaggerRoutes(mux)
	worldAPI.Register(mux)
	playerAPI.Register(mux)
	campaignAPI.Register(mux)
	planeAPI.Register(mux)
	sessionAPI.Register(mux)
	tagAPI.Register(mux)
	noteAPI.Register(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  serverReadTO,
		WriteTimeout: serverWriteTO,
		IdleTimeout:  serverIdleTO,
	}

	errCh := make(chan error, 1)
	go func() {
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownT)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case serveErr, ok := <-errCh:
		if !ok {
			return nil
		}
		return serveErr
	}
}

func registerMetaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
}

func loadConfig() (Config, error) {
	loadEnvFile(".env")

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = defaultPort
	}

	return Config{
		DatabaseURL: databaseURL,
		Port:        port,
	}, nil
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}
