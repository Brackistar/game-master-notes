param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("up", "down", "version", "force", "create")]
  [string]$Action,

  [int]$Version,
  [string]$Name
)

if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
  Write-Error "'migrate' CLI not found. Install golang-migrate first: https://github.com/golang-migrate/migrate"
  exit 1
}

if (-not (Test-Path .env)) {
  Write-Error "Missing .env file. Copy .env.example to .env first."
  exit 1
}

Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $parts = $_ -split '=', 2
  if ($parts.Count -eq 2) {
    [System.Environment]::SetEnvironmentVariable($parts[0].Trim(), $parts[1].Trim())
  }
}

if (-not $env:DATABASE_URL) {
  Write-Error "DATABASE_URL is not set in .env"
  exit 1
}

switch ($Action) {
  "up"      { migrate -path migrations -database $env:DATABASE_URL up }
  "down"    { migrate -path migrations -database $env:DATABASE_URL down 1 }
  "version" { migrate -path migrations -database $env:DATABASE_URL version }
  "force"   {
    if ($PSBoundParameters.ContainsKey('Version') -eq $false) {
      Write-Error "Version is required for action=force"
      exit 1
    }
    migrate -path migrations -database $env:DATABASE_URL force $Version
  }
  "create"  {
    if ([string]::IsNullOrWhiteSpace($Name)) {
      Write-Error "Name is required for action=create"
      exit 1
    }
    migrate create -ext sql -dir migrations -seq $Name
  }
}
