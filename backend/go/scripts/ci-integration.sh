#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="docker-compose.db.yml"
POSTGRES_CONTAINER="gm_notes_postgres"
MIGRATE_VERSION="${MIGRATE_VERSION:-v4.19.1}"
TEST_TARGET="${TEST_TARGET:-./...}"

cleanup() {
  echo "[ci-integration] Cleaning up database containers..."
  docker compose -f "$COMPOSE_FILE" --env-file .env down -v
}
trap cleanup EXIT

echo "[ci-integration] Preparing environment..."
if [[ ! -f ".env" ]]; then
  cp .env.example .env
fi

echo "[ci-integration] Starting PostgreSQL..."
docker compose -f "$COMPOSE_FILE" --env-file .env up -d

echo "[ci-integration] Waiting for PostgreSQL health..."
for i in {1..60}; do
  status="$(docker inspect --format='{{.State.Health.Status}}' "$POSTGRES_CONTAINER" 2>/dev/null || true)"
  if [[ "$status" == "healthy" ]]; then
    echo "[ci-integration] PostgreSQL is healthy."
    break
  fi
  if [[ "$i" -eq 60 ]]; then
    echo "[ci-integration] PostgreSQL did not become healthy in time."
    docker logs "$POSTGRES_CONTAINER" || true
    exit 1
  fi
  sleep 2
done

if ! command -v migrate >/dev/null 2>&1; then
  echo "[ci-integration] Installing migrate $MIGRATE_VERSION..."
  BIN_DIR="${RUNNER_TEMP:-/tmp}/migrate-bin"
  mkdir -p "$BIN_DIR"
  curl -fsSL "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${MIGRATE_VERSION#v}.linux-amd64.tar.gz" \
    | tar -xz -C "$BIN_DIR"
  chmod +x "$BIN_DIR/migrate"
  export PATH="$BIN_DIR:$PATH"
fi

echo "[ci-integration] Loading DATABASE_URL from .env..."
set -a
# shellcheck disable=SC1091
source .env
set +a

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "[ci-integration] DATABASE_URL is not defined in .env"
  exit 1
fi

echo "[ci-integration] Applying migrations..."
migrate -path migrations -database "$DATABASE_URL" up

echo "[ci-integration] Running integration tests..."
go test -tags=integration "$TEST_TARGET" -count=1 -v

echo "[ci-integration] Integration tests completed successfully."
