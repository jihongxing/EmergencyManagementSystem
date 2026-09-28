param()
#requires -Version 7.0
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
function Invoke-Checked([string]$tool, [string[]]$arguments) {
    & $tool @arguments
    if ($LASTEXITCODE -ne 0) { throw "$tool failed ($LASTEXITCODE)" }
}
if (-not $env:POSTGRES_PASSWORD) { throw 'Set POSTGRES_PASSWORD for the local development database.' }
$port = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { '55472' }
$name = 'ems_test_' + [Guid]::NewGuid().ToString('N')
$old = $env:TEST_DATABASE_URL
$oldPersisted = $env:TEST_PERSISTED_DATABASE_URL
Push-Location $root
try {
    Invoke-Checked podman @('compose', '-f', 'infra/compose.yaml', 'up', '-d')
    $ready = $false
    for ($i=0; $i -lt 30; $i++) {
        & podman compose -f infra/compose.yaml exec -T postgres pg_isready -U emergency_dev -d emergency_dev
        if ($LASTEXITCODE -eq 0) { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) { throw 'PostgreSQL did not become ready.' }
    Invoke-Checked podman @('compose','-f','infra/compose.yaml','exec','-T','postgres','createdb','-U','emergency_dev',$name)
    $password = [Uri]::EscapeDataString($env:POSTGRES_PASSWORD)
    $env:TEST_DATABASE_URL = "postgres://emergency_dev:${password}@127.0.0.1:${port}/${name}?sslmode=disable"
    Push-Location backend
    try { Invoke-Checked go @('test','./internal/database','-run','TestIntegration','-count=1','-v') }
    finally { Pop-Location }
    Invoke-Checked podman @('compose','-f','infra/compose.yaml','stop','postgres')
    Invoke-Checked podman @('compose','-f','infra/compose.yaml','start','postgres')
    $ready = $false
    for ($i=0; $i -lt 30; $i++) {
        & podman compose -f infra/compose.yaml exec -T postgres pg_isready -U emergency_dev -d $name
        if ($LASTEXITCODE -eq 0) { $ready = $true; break }
        Start-Sleep -Seconds 1
    }
    if (-not $ready) { throw 'PostgreSQL did not recover after restart.' }
    $version = & podman compose -f infra/compose.yaml exec -T postgres psql -U emergency_dev -d $name -Atc 'SELECT version_id FROM goose_db_version ORDER BY id DESC LIMIT 1'
    if ($LASTEXITCODE -ne 0 -or ($version -join '').Trim() -ne '1') { throw 'Migration state did not survive restart.' }
    $env:TEST_PERSISTED_DATABASE_URL = $env:TEST_DATABASE_URL
    Push-Location backend
    try { Invoke-Checked go @('test','./internal/database','-run','TestPersistedReadiness','-count=1','-v') }
    finally { Pop-Location }
    Write-Output "PASS: database integration and restart persistence; retained dedicated test database $name"
} finally {
    $env:TEST_DATABASE_URL = $old
    $env:TEST_PERSISTED_DATABASE_URL = $oldPersisted
    Pop-Location
}
