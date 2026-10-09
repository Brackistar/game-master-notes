package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("startup_begin", "component", "backend")

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("startup_failed_config", "error", err)
		return err
	}
	slog.Info("config_loaded", "port", cfg.Port, "database_url_set", cfg.DatabaseURL != "")

	slog.Info("db_connect_start")
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db_connect_failed", "error", err)
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()
	slog.Info("db_connect_end")

	slog.Info("db_ping_start")
	if err := pool.Ping(ctx); err != nil {
		slog.Error("db_ping_failed", "error", err)
		return fmt.Errorf("ping db: %w", err)
	}
	slog.Info("db_ping_end")

	idGenerator := serviceshared.NewOklogULIDGenerator()
	slog.Info("dependency_created", "name", "id_generator")

	slog.Info("repository_wiring_start")
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
	slog.Info("repository_wiring_end")

	slog.Info("service_wiring_start")
	worldService := service.NewWorldService(worldRepo, idGenerator)
	playerService := service.NewPlayerService(playerRepo, idGenerator)
	campaignService := service.NewCampaignService(campaignRepo, campaignPlayerRepo, idGenerator)
	planeService := service.NewPlaneService(planeRepo, idGenerator)
	sessionService := service.NewSessionService(sessionRepo, idGenerator)
	tagService := service.NewTagService(tagRepo, idGenerator)
	noteService := service.NewNoteService(noteRepo, noteOwnerRepo, noteTagRepo, noteLinkRepo, noteAssetRepo, mapPlacementRepo, idGenerator)
	slog.Info("service_wiring_end")

	slog.Info("api_wiring_start")
	worldAPI := api.NewWorldAPI(worldService)
	playerAPI := api.NewPlayerAPI(playerService)
	campaignAPI := api.NewCampaignAPI(campaignService)
	planeAPI := api.NewPlaneAPI(planeService)
	sessionAPI := api.NewSessionAPI(sessionService)
	tagAPI := api.NewTagAPI(tagService)
	noteAPI := api.NewNoteAPI(noteService)
	slog.Info("api_wiring_end")

	mux := http.NewServeMux()
	slog.Info("routes_registration_start")
	registerMetaRoutes(mux)
	registerSwaggerRoutes(mux)
	worldAPI.Register(mux)
	playerAPI.Register(mux)
	campaignAPI.Register(mux)
	planeAPI.Register(mux)
	sessionAPI.Register(mux)
	tagAPI.Register(mux)
	noteAPI.Register(mux)
	slog.Info("routes_registration_end")

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      auditHTTPMiddleware(mux),
		ReadTimeout:  serverReadTO,
		WriteTimeout: serverWriteTO,
		IdleTimeout:  serverIdleTO,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http_server_start", "addr", server.Addr)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			slog.Error("http_server_runtime_error", "error", serveErr)
			errCh <- serveErr
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown_begin", "reason", "context_done")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownT)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("shutdown_failed", "error", err)
			return err
		}
		slog.Info("shutdown_end")
		return nil
	case serveErr, ok := <-errCh:
		if !ok {
			slog.Info("http_server_stopped")
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
	slog.Info("config_load_start")
	loadEnvFile(".env")

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		slog.Error("config_missing", "field", "DATABASE_URL")
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
	slog.Info("env_file_load_start", "path", path)
	file, err := os.Open(path)
	if err != nil {
		slog.Warn("env_file_load_skipped", "path", path, "error", err)
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
		slog.Info("env_var_loaded", "key", key)
	}
	slog.Info("env_file_load_end", "path", path)
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func auditHTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		slog.Info("http_request_start",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		slog.Info("http_request_end",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
