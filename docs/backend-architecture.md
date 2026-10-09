# Backend Architecture Roadmap (Go)

This is the source-of-truth roadmap for the Go backend.

## 1) Scope

- Core model definitions
- PostgreSQL schema and migrations
- Repository layer
- Service layer
- HTTP/JSON + gRPC transport integration

## 2) Locked Decisions

- External API: `HTTP + JSON` (`chi`)
- Internal backend communication: `gRPC`
- API prefix: `/api/v1`
- Auth: none in Phase 1
- Single-user context in Phase 1
- Soft delete by default (`deleted_at`), hidden unless explicitly included
- Strict delete behavior (`not found` if missing/already deleted)
- ULID in app layer
- UTC timestamps
- Optimistic locking with `version` (checked in service layer)
- SQL-first approach: `pgx + sqlc + golang-migrate`
- DB multi-table operations should be moved to `fn_*` PostgreSQL functions in Phase 2

## 3) Data Layer Strategy

### SQL + Codegen

- Query sources: `backend/go/src/repository/postgres/queries/*.sql`
- Generated code: `backend/go/src/repository/postgres/generated/*.go`
- Concrete repos: `backend/go/src/repository/repos/*.go`
- Interfaces: `backend/go/src/repository/interfaces/*.go`
- Shared repository errors: `backend/go/src/repository/error/errors.go`
- Stable DB function token contract (`GMN_*`): `backend/go/db/GMN_ERROR_CONTRACT.md` and `backend/go/src/repository/error/gmn_contract.go`

### Migrations

- `001_init_phase1`: complete schema with enums/tables/indexes/trigger
- `002_service_functions`: implemented (Phase 2)

## 4) Current Repository Coverage

Implemented and integration-tested repositories:

- Base entities:
  - `WorldRepository`
  - `PlaneRepository`
  - `CampaignRepository`
  - `PlayerRepository`
  - `SessionRepository`
  - `NoteRepository`
  - `TagRepository`
- Supporting relations:
  - `CampaignPlayerRepository`
  - `NoteOwnerRepository`
  - `NoteTagRepository`
  - `NoteAssetRepository`
  - `MapNotePlacementRepository`
  - `NoteLinkRepository`

All above have:

- interface contract
- SQL query file
- sqlc-generated bindings
- concrete repo adapter

## 5) Project Structure (Current)

```text
backend/go/
|-- migrations/
|-- scripts/
|-- db/
|-- src/
|   |-- model/
|   |-- repository/
|   |   |-- error/
|   |   |   `-- errors.go
|   |   |-- interfaces/
|   |   |-- postgres/
|   |   |   |-- queries/
|   |   |   `-- generated/
|   |   `-- repos/
|   `-- service/   # pending implementation
`-- cmd/api/       # pending real app wiring
```

## 6) Testing Status

- Repository compilation: passing
- Integration tests (build tag `integration`): passing for all implemented repositories
- CI script for integration pipeline exists:
  - `backend/go/scripts/ci-integration.sh`
- Service-layer testing policy:
  - Service tests are unit tests only (no DB integration in `src/service` tests)
  - All service dependencies are injected and mocked/faked in tests
  - DB integration coverage lives in repository integration tests

## 7) Known Pending Work (Start of Phase 2)

1. Implement service layer per domain with explicit methods.
2. Move multi-table business mutations to service methods backed by repository `fn_*` wrappers.
3. Add service error mapping from repository errors + SQLSTATE/token semantics into typed service errors.
4. Add service-focused tests (happy path + conflict/not-found/validation edge cases) as primary safety net.
5. Begin communication layer wiring (HTTP/JSON + gRPC) once service contracts stabilize.

## 8) Service Layer Plan (Phase 2)

- One service per domain
- Explicit methods (no generic CRUD-only naming)
- Services return model structs
- DTO mapping belongs only to transport layer
- Pagination params: `offset`, `limit`, `include_deleted`
- Reads hide deleted by default
- Session default list order: `played_on DESC, created_at DESC, id DESC`

## 9) Phase 2 Handoff Checklist

When a new session starts:

1. Verify DB is up:
   - `./scripts/db.ps1 -Action up`
2. Verify migrations:
   - `./scripts/migrate.ps1 -Action version`
3. Run integration suite baseline:
   - `go test -tags=integration ./src/repository/repos -count=1 -v`
4. Create `002_service_functions` migration.
   - `./scripts/migrate.ps1 -Action up` enforces post-run version >= 2, confirming function migration runs after schema init.
5. Add repository wrappers for `fn_*` operations and validate with integration tests.
6. Start service-layer vertical implementation (recommended first: campaign-player add/remove) with service-focused tests.

## 10) Decision Change Log

```md
## Decision Update - YYYY-MM-DD
- Changed:
- Reason:
- Impacted areas:
```
