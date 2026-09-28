param([switch]$DocsOnly)
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
function Run([string]$directory, [string]$tool, [string[]]$arguments) {
    Push-Location (Join-Path $root $directory)
    try {
        & $tool @arguments
        if ($LASTEXITCODE -ne 0) { throw "$tool failed with exit code $LASTEXITCODE" }
    } finally { Pop-Location }
}
Run '.' 'powershell.exe' @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', 'scripts/test/check-doc-governance.ps1')
$contract = Get-Content (Join-Path $root 'constras/platform/live.openapi.json') -Raw | ConvertFrom-Json
if ($contract.openapi -ne '3.1.0' -or $contract.paths.'/health/live'.get.operationId -ne 'getLiveness') {
    throw 'Liveness contract is missing or inconsistent'
}
if ($DocsOnly) {
    Write-Output 'PARTIAL: document checks only; project checks skipped.'
    exit 0
}
Run 'backend' 'go' @('test', './...')
Run 'backend' 'go' @('vet', './...')
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
