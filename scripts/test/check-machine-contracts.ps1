param(
    [string]$ContractRoot,
    [switch]$SkipNegativeTests
)
#requires -Version 7.0

$ErrorActionPreference = 'Stop'
$scriptPath = $MyInvocation.MyCommand.Path
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
if ([string]::IsNullOrWhiteSpace($ContractRoot)) {
    $ContractRoot = Join-Path $repoRoot 'constras'
}
$ContractRoot = [IO.Path]::GetFullPath($ContractRoot)
$errors = [Collections.Generic.List[string]]::new()

function Fail([string]$message) {
    $errors.Add($message)
}

function Read-Json([string]$path) {
    try {
        return Get-Content -LiteralPath $path -Encoding UTF8 -Raw | ConvertFrom-Json -Depth 100
    } catch {
        Fail "Invalid JSON: ${path}: $($_.Exception.Message)"
        return $null
    }
}

function Require-Properties($object, [string[]]$required, [string[]]$allowed, [string]$label) {
    if ($null -eq $object) { return }
    $actual = @($object.PSObject.Properties.Name)
    foreach ($name in $required) {
        if ($name -notin $actual) { Fail "$label is missing required field '$name'" }
    }
    foreach ($name in $actual) {
        if ($name -notin $allowed) { Fail "$label has unknown field '$name'" }
    }
}

function Require-String($value, [string]$label) {
    if ($value -isnot [string] -or [string]::IsNullOrWhiteSpace($value)) {
        Fail "$label must be a non-empty string"
    }
}

function Require-Bool($value, [string]$label) {
    if ($value -isnot [bool]) { Fail "$label must be boolean" }
}

function Require-StringArray($value, [string]$label, [switch]$NonEmpty) {
    if ($value -isnot [array]) {
        Fail "$label must be an array"
        return
    }
    if ($NonEmpty -and $value.Count -eq 0) { Fail "$label must not be empty" }
    foreach ($item in $value) { Require-String $item "$label item" }
    if (@($value | Sort-Object -Unique).Count -ne $value.Count) {
        Fail "$label must not contain duplicates"
    }
}

function Validate-BehaviorContract([string]$path) {
    $name = Split-Path $path -Leaf
    $contract = Read-Json $path
    if ($null -eq $contract) { return }
    Require-Properties $contract @('version', 'source', 'manualClauses', 'policies') @('version', 'source', 'manualClauses', 'policies') $name
    if ($contract.version -ne 1) { Fail "$name.version must be 1" }
    Require-String $contract.source "$name.source"
    if ($contract.source -ne [IO.Path]::ChangeExtension($name, '.md') -or
        -not (Test-Path -LiteralPath (Join-Path $ContractRoot $contract.source) -PathType Leaf)) {
        Fail "$name source reference is invalid"
    }
    Require-StringArray $contract.manualClauses "$name.manualClauses"
    if ($contract.policies -isnot [array] -or $contract.policies.Count -eq 0) {
        Fail "$name.policies must be a non-empty array"
        return
    }
    $policyIds = [Collections.Generic.HashSet[string]]::new()
    foreach ($policy in $contract.policies) {
        Require-Properties $policy @('id', 'clauses', 'facts', 'allowWhen', 'cases') @('id', 'clauses', 'facts', 'allowWhen', 'cases') "$name policy"
        Require-String $policy.id "$name policy.id"
        if (-not $policyIds.Add($policy.id)) { Fail "$name has duplicate policy id '$($policy.id)'" }
        Require-StringArray $policy.clauses "$name/$($policy.id).clauses" -NonEmpty
        Require-StringArray $policy.facts "$name/$($policy.id).facts" -NonEmpty
        if ($policy.allowWhen -isnot [pscustomobject]) { Fail "$name/$($policy.id).allowWhen must be an object"; continue }
        $factNames = @($policy.facts)
        foreach ($fact in $factNames) {
            if ($null -eq $policy.allowWhen.PSObject.Properties[$fact]) {
                Fail "$name/$($policy.id).allowWhen missing fact '$fact'"
            } else {
                Require-Bool $policy.allowWhen.$fact "$name/$($policy.id).allowWhen.$fact"
            }
        }
        foreach ($property in $policy.allowWhen.PSObject.Properties) {
            if ($property.Name -notin $factNames) { Fail "$name/$($policy.id).allowWhen has unknown fact '$($property.Name)'" }
        }
        if ($policy.cases -isnot [array] -or $policy.cases.Count -lt 2) {
            Fail "$name/$($policy.id).cases must contain at least two cases"
            continue
        }
        $caseNames = [Collections.Generic.HashSet[string]]::new()
        $hasAllow = $false
        $hasDeny = $false
        foreach ($case in $policy.cases) {
            Require-Properties $case @('name', 'input', 'allowed') @('name', 'input', 'allowed') "$name/$($policy.id) case"
            Require-String $case.name "$name/$($policy.id) case.name"
            if (-not $caseNames.Add($case.name)) { Fail "$name/$($policy.id) has duplicate case '$($case.name)'" }
            Require-Bool $case.allowed "$name/$($policy.id)/$($case.name).allowed"
            if ($case.allowed) { $hasAllow = $true } else { $hasDeny = $true }
            if ($case.input -isnot [pscustomobject]) { Fail "$name/$($policy.id)/$($case.name).input must be an object"; continue }
            foreach ($fact in $factNames) {
                if ($null -eq $case.input.PSObject.Properties[$fact]) {
                    Fail "$name/$($policy.id)/$($case.name).input missing fact '$fact'"
                } else {
                    Require-Bool $case.input.$fact "$name/$($policy.id)/$($case.name).input.$fact"
                }
            }
            foreach ($property in $case.input.PSObject.Properties) {
                if ($property.Name -notin $factNames) { Fail "$name/$($policy.id)/$($case.name).input has unknown fact '$($property.Name)'" }
            }
            if ($case.allowed -is [bool] -and $case.input -is [pscustomobject] -and
                @($case.input.PSObject.Properties.Name).Count -eq $factNames.Count) {
                $actualAllowed = @($factNames | Where-Object { $case.input.$_ -cne $policy.allowWhen.$_ }).Count -eq 0
                if ($actualAllowed -ne $case.allowed) {
                    Fail "$name/$($policy.id)/$($case.name).allowed does not match allowWhen"
                }
            }
        }
        if (-not $hasAllow -or -not $hasDeny) { Fail "$name/$($policy.id) must include both allowed and denied cases" }
    }
}

function Validate-IdentityContract([string]$path) {
    $name = Split-Path $path -Leaf
    $contract = Read-Json $path
    if ($null -eq $contract) { return }
    Require-Properties $contract @('version', 'clauses', 'cases') @('version', 'clauses', 'cases') $name
    if ($contract.version -ne 1) { Fail "$name.version must be 1" }
    Require-StringArray $contract.clauses "$name.clauses" -NonEmpty
    $source = Join-Path $ContractRoot '02-access-and-evidence.md'
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        Fail "$name has missing clause source"
    } else {
        $sourceText = Get-Content -LiteralPath $source -Encoding UTF8 -Raw
        foreach ($clause in @($contract.clauses)) {
            if ($sourceText -notmatch ("(?m)^- ``" + [regex]::Escape($clause) + "``")) {
                Fail "$name has unknown clause '$clause'"
            }
        }
    }
    if ($contract.cases -isnot [array] -or $contract.cases.Count -eq 0) { Fail "$name.cases must be a non-empty array"; return }
    $caseNames = [Collections.Generic.HashSet[string]]::new()
    $allowedResults = @('allowed', 'role_denied', 'inactive', 'scope_denied', 'invalid_identity')
    foreach ($case in $contract.cases) {
        Require-Properties $case @('name', 'member', 'target', 'expected') @('name', 'member', 'target', 'expected') "$name case"
        Require-String $case.name "$name case.name"
        if (-not $caseNames.Add($case.name)) { Fail "$name has duplicate case '$($case.name)'" }
        if ($case.target -isnot [string]) { Fail "$name/$($case.name).target must be a string" }
        if ($case.expected -notin $allowedResults) { Fail "$name/$($case.name).expected has invalid value" }
        Require-Properties $case.member @('userId', 'organizationId', 'organizationKind', 'active', 'roles') @('userId', 'organizationId', 'organizationKind', 'active', 'roles') "$name/$($case.name).member"
        if ($case.member.userId -isnot [string]) { Fail "$name/$($case.name).member.userId must be a string" }
        if ($case.member.organizationId -isnot [string]) { Fail "$name/$($case.name).member.organizationId must be a string" }
        Require-String $case.member.organizationKind "$name/$($case.name).member.organizationKind"
        Require-Bool $case.member.active "$name/$($case.name).member.active"
        if ($case.member.roles -isnot [array]) {
            Fail "$name/$($case.name).member.roles must be an array"
        } else {
            foreach ($role in $case.member.roles) { Require-String $role "$name/$($case.name).member.roles item" }
        }
    }
}

function Validate-ResponseSchema($schema, [string]$label, [string]$expectedConst) {
    Require-Properties $schema @('type', 'required', 'additionalProperties', 'properties') @('type', 'required', 'additionalProperties', 'properties') $label
    if ($schema.type -ne 'object') { Fail "$label.type must be object" }
    if (@($schema.required).Count -ne 1 -or $schema.required[0] -ne 'status') { Fail "$label.required must contain only status" }
    if ($schema.additionalProperties -ne $false) { Fail "$label.additionalProperties must be false" }
    if ($null -eq $schema.properties -or $null -eq $schema.properties.status) { Fail "$label.properties.status is required"; return }
    Require-Properties $schema.properties.status @('const') @('type', 'const') "$label.properties.status"
    if ($schema.properties.status.const -ne $expectedConst) { Fail "$label status const must be '$expectedConst'" }
}

function Validate-OpenApi([string]$path, [string]$expectedPath, [string]$expectedOperation, [hashtable]$expectedResponses, [string]$expectedTitle) {
    $name = Split-Path $path -Leaf
    $contract = Read-Json $path
    if ($null -eq $contract) { return }
    Require-Properties $contract @('openapi', 'info', 'paths') @('openapi', 'info', 'paths') $name
    if ($contract.openapi -ne '3.1.0') { Fail "$name.openapi must be 3.1.0" }
    Require-Properties $contract.info @('title', 'version') @('title', 'version') "$name.info"
    if ($contract.info.title -ne $expectedTitle) { Fail "$name.info.title is unexpected" }
    Require-String $contract.info.version "$name.info.version"
    Require-Properties $contract.paths @($expectedPath) @($expectedPath) "$name.paths"
    $pathItem = $contract.paths.PSObject.Properties[$expectedPath].Value
    Require-Properties $pathItem @('get') @('get') "$name.$expectedPath"
    $operation = $pathItem.get
    Require-Properties $operation @('operationId', 'responses') @('operationId', 'description', 'security', 'responses') "$name.$expectedPath.get"
    if ($operation.operationId -ne $expectedOperation) { Fail "$name operationId must be '$expectedOperation'" }
    $responseNames = @($operation.responses.PSObject.Properties.Name)
    foreach ($code in $expectedResponses.Keys) {
        if ($code -notin $responseNames) { Fail "$name missing response '$code'"; continue }
        $response = $operation.responses.PSObject.Properties[$code].Value
        Require-Properties $response @('description', 'content') @('description', 'content') "$name response $code"
        Require-Properties $response.content @('application/json') @('application/json') "$name response $code.content"
        $schema = $response.content.'application/json'.schema
        Validate-ResponseSchema $schema "$name response $code schema" $expectedResponses[$code]
    }
    foreach ($code in $responseNames) {
        if (-not $expectedResponses.ContainsKey($code)) { Fail "$name has unknown response '$code'" }
    }
}

foreach ($relative in @(
    '01-scope-and-identity.json', '02-access-and-evidence.json', '03-rules-and-records.json',
    '04-subscription.json', '05-platform.json', 'identity/authorization.json',
    'platform/live.openapi.json', 'platform/ready.openapi.json'
)) {
    if (-not (Test-Path -LiteralPath (Join-Path $ContractRoot $relative) -PathType Leaf)) {
        Fail "Missing machine contract: $relative"
    }
}
foreach ($file in @(Get-ChildItem -LiteralPath $ContractRoot -Filter '*.json' -File)) {
    Validate-BehaviorContract $file.FullName
}
Validate-IdentityContract (Join-Path $ContractRoot 'identity/authorization.json')
Validate-OpenApi (Join-Path $ContractRoot 'platform/live.openapi.json') '/health/live' 'getLiveness' @{ '200' = 'alive' } 'Process liveness'
Validate-OpenApi (Join-Path $ContractRoot 'platform/ready.openapi.json') '/health/ready' 'getReadiness' @{ '200' = 'ready'; '503' = 'not_ready' } 'Database readiness'

if (-not $SkipNegativeTests) {
    $fixtureRoot = Join-Path ([IO.Path]::GetTempPath()) ('ems-machine-contracts-' + [guid]::NewGuid().ToString('N'))
    try {
        Copy-Item -LiteralPath $ContractRoot -Destination $fixtureRoot -Recurse
        $mutations = @(
            @{ Name = 'unknown field'; Expected = "unknown field 'unknown'"; Path = '01-scope-and-identity.json'; Mutate = { param($x) Add-Member -InputObject $x -MemberType NoteProperty -Name unknown -Value $true -Force } },
            @{ Name = 'missing required field'; Expected = "missing required field 'version'"; Path = '01-scope-and-identity.json'; Mutate = { param($x) $x.PSObject.Properties.Remove('version') } },
            @{ Name = 'incorrect case result'; Expected = 'allowed does not match allowWhen'; Path = '01-scope-and-identity.json'; Mutate = { param($x) $x.policies[0].cases[0].allowed = $false } },
            @{ Name = 'unknown identity field'; Expected = "unknown field 'unknown'"; Path = 'identity/authorization.json'; Mutate = { param($x) Add-Member -InputObject $x.cases[0].member -MemberType NoteProperty -Name unknown -Value $true -Force } },
            @{ Name = 'wrong operation id'; Expected = "operationId must be 'getLiveness'"; Path = 'platform/live.openapi.json'; Mutate = { param($x) $x.paths.'/health/live'.get.operationId = 'wrongOperation' } },
            @{ Name = 'wrong response status'; Expected = "missing response '503'"; Path = 'platform/ready.openapi.json'; Mutate = { param($x) $x.paths.'/health/ready'.get.responses.PSObject.Properties.Remove('503') } },
            @{ Name = 'unknown clause reference'; Expected = "unknown clause 'C02-99'"; Path = 'identity/authorization.json'; Mutate = { param($x) $x.clauses[0] = 'C02-99' } },
            @{ Name = 'response schema mismatch'; Expected = "status const must be 'alive'"; Path = 'platform/live.openapi.json'; Mutate = { param($x) $x.paths.'/health/live'.get.responses.'200'.content.'application/json'.schema.properties.status.const = 'dead' } }
        )
        foreach ($mutation in $mutations) {
            $fixture = Join-Path $fixtureRoot $mutation.Path
            $value = Get-Content -LiteralPath $fixture -Encoding UTF8 -Raw | ConvertFrom-Json -Depth 100
            & $mutation.Mutate $value
            $value | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $fixture -Encoding UTF8
            $result = & pwsh -NoProfile -ExecutionPolicy Bypass -File $scriptPath -ContractRoot $fixtureRoot -SkipNegativeTests 2>&1
            if ($LASTEXITCODE -eq 0 -or ($result -join "`n").IndexOf($mutation.Expected, [StringComparison]::Ordinal) -lt 0) {
                Fail "Negative fixture did not report expected error: $($mutation.Name)"
            }
            Copy-Item -LiteralPath (Join-Path $ContractRoot $mutation.Path) -Destination $fixture -Force
        }
        $fixture = Join-Path $fixtureRoot 'identity/authorization.json'
        Set-Content -LiteralPath $fixture -Encoding UTF8 -Value '{'
        $result = & pwsh -NoProfile -ExecutionPolicy Bypass -File $scriptPath -ContractRoot $fixtureRoot -SkipNegativeTests 2>&1
        if ($LASTEXITCODE -eq 0 -or ($result -join "`n").IndexOf('Invalid JSON:', [StringComparison]::Ordinal) -lt 0) {
            Fail 'Negative fixture did not report invalid JSON'
        }
        Copy-Item -LiteralPath (Join-Path $ContractRoot 'identity/authorization.json') -Destination $fixture -Force
        $fixture = Join-Path $fixtureRoot 'platform/ready.openapi.json'
        Remove-Item -LiteralPath $fixture
        $result = & pwsh -NoProfile -ExecutionPolicy Bypass -File $scriptPath -ContractRoot $fixtureRoot -SkipNegativeTests 2>&1
        if ($LASTEXITCODE -eq 0 -or ($result -join "`n").IndexOf('Missing machine contract: platform/ready.openapi.json', [StringComparison]::Ordinal) -lt 0) {
            Fail 'Negative fixture did not report missing contract'
        }
    } finally {
        if (Test-Path -LiteralPath $fixtureRoot) { Remove-Item -LiteralPath $fixtureRoot -Recurse -Force }
    }
}

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Output $_ }
    exit 1
}
Write-Output 'PASS: machine contract structure, references, response schemas, and negative fixtures.'
