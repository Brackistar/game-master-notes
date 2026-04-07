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

## 4) Useful commands

```powershell
./scripts/db.ps1 -Action logs
./scripts/db.ps1 -Action down
./scripts/db.ps1 -Action drop
./scripts/migrate.ps1 -Action version
./scripts/migrate.ps1 -Action create -Name add_new_table
```

## Notes

- `docker-compose.db.yml` uses `pgvector/pgvector:pg16`.
- DB init scripts run only on first volume initialization.
