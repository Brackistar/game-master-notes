param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("up", "down", "drop", "logs", "ps")]
  [string]$Action
)

$composeFile = "docker-compose.db.yml"

switch ($Action) {
  "up"   { docker compose -f $composeFile --env-file .env up -d }
  "down" { docker compose -f $composeFile --env-file .env down }
  "drop" { docker compose -f $composeFile --env-file .env down -v }
  "logs" { docker compose -f $composeFile --env-file .env logs -f postgres }
  "ps"   { docker compose -f $composeFile --env-file .env ps }
}
