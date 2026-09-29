param([switch]$DocsOnly, [switch]$Database)
#requires -Version 7.0
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
function Run([string]$directory, [string]$tool, [string[]]$arguments) {
    Push-Location (Join-Path $root $directory)
    try {
        & $tool @arguments
        if ($LASTEXITCODE -ne 0) { throw "$tool failed with exit code $LASTEXITCODE" }
    } finally { Pop-Location }
}
Run '.' 'pwsh' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', 'scripts/test/check-doc-governance.ps1')
Run '.' 'pwsh' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', 'scripts/test/check-machine-contracts.ps1')
if ($DocsOnly) {
    Write-Output 'PARTIAL: document and machine structure checks only; OpenAPI specification, real responses and project checks skipped.'
    exit 0
}
Run 'scripts/contracts' 'npm.cmd' @('run', 'check')
Run 'backend' 'go' @('test', './...')
Run 'backend' 'go' @('vet', './...')
if ($Database) {
    Run '.' 'pwsh' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', 'scripts/test/check-database.ps1')
} else {
    Write-Output 'NOTE: real PostgreSQL integration skipped; use -Database with POSTGRES_PASSWORD.'
}
Run 'apps/admin-web' 'npm.cmd' @('run', 'build')
Run 'apps/admin-web' 'npm.cmd' @('test')
foreach ($platform in @('android', 'ios')) {
    if (-not (Test-Path (Join-Path $root "apps/mobile/$platform"))) {
        throw "Missing native Flutter scaffold: $platform"
    }
}
if (-not (Test-Path (Join-Path $root 'apps/mobile/pubspec.lock'))) {
    throw 'Flutter dependencies are not locked. Complete flutter pub get first.'
}
Run 'apps/mobile' 'flutter.bat' @('analyze', '--no-pub')
Run 'apps/mobile' 'flutter.bat' @('test', '--no-pub')
Write-Output 'PASS: engineering preflight (not production readiness).'
