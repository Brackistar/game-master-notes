# Database Setup (Local)

## 1) Create local env file

Copy `.env.example` to `.env` and adjust values if needed.

## 2) Start database

```powershell
./scripts/db.ps1 -Action up
```

## 3) Run migrations

```powershell
./scripts/migrate.ps1 -Action up
```

## 3.1) Run repository integration baseline

```powershell
go test -tags=integration ./src/repository/repos -count=1 -v
```

## 4) Useful commands

```powershell
./scripts/db.ps1 -Action logs
./scripts/db.ps1 -Action down
./scripts/db.ps1 -Action drop
./scripts/migrate.ps1 -Action version
./scripts/migrate.ps1 -Action create -Name add_new_table
```

## 5) CI integration run

GitHub Actions/Linux runners can use:

```bash
bash ./scripts/ci-integration.sh
```

This script:

- boots PostgreSQL container
- waits for health
- applies migrations
- runs integration tests (`-tags=integration`)
- tears down containers/volumes automatically

## Notes

- `docker-compose.db.yml` uses `pgvector/pgvector:pg16`.
- DB init scripts run only on first volume initialization.
- `001_init_phase1` is implemented and validated.
- `002_service_functions` is pending (Phase 2).
