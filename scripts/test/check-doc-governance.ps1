param()

$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$docs = Join-Path $root 'docs'
$contracts = Join-Path $root 'constras'
$errors = [Collections.Generic.List[string]]::new()
$required = @(
    'AGENTS.md',
    'docs/README.md',
    'docs/GOVERNANCE.md',
    'docs/DECISIONS.md',
    'docs/OPEN.md',
    'docs/implementation/README.md',
    'docs/research/market-handover.md',
    'docs/ssot/01-product.md',
    'docs/ssot/02-authority.md',
    'docs/ssot/03-site-identity.md',
    'docs/ssot/04-access.md',
    'docs/ssot/05-workflows.md',
    'docs/ssot/06-technology.md',
    'constras/README.md',
    'constras/01-scope-and-identity.md',
    'constras/02-access-and-evidence.md',
    'constras/03-rules-and-records.md',
    'constras/04-subscription.md',
    'constras/05-platform.md',
    'constras/01-scope-and-identity.json',
    'constras/02-access-and-evidence.json',
    'constras/03-rules-and-records.json',
    'constras/04-subscription.json',
    'constras/05-platform.json'
)
foreach ($path in $required) {
    if (-not (Test-Path -LiteralPath (Join-Path $root $path) -PathType Leaf)) {
        $errors.Add("Missing required entry: $path")
    }
}

$files = @(Get-ChildItem -LiteralPath $docs -Filter '*.md' -Recurse)
$files += @(Get-ChildItem -LiteralPath $contracts -Filter '*.md')
$files += Get-Item -LiteralPath (Join-Path $root 'AGENTS.md')
$linkCount = 0
$tableCount = 0
$contractIds = [Collections.Generic.HashSet[string]]::new()
foreach ($file in $files) {
    # Historical snapshots retain original syntax and sources.
    if ($file.FullName.StartsWith((Join-Path $docs 'archive') + [IO.Path]::DirectorySeparatorChar)) {
        continue
    }
    $content = Get-Content -LiteralPath $file.FullName -Encoding UTF8 -Raw
    foreach ($match in [regex]::Matches($content, '\[[^\]]*\]\(([^)]+)\)')) {
        $target = $match.Groups[1].Value
        if ($target -match '^[a-zA-Z][a-zA-Z0-9+.-]*:' -or $target.StartsWith('#')) {
            continue
        }
        $path = ($target -split '#', 2)[0]
        $resolved = [IO.Path]::GetFullPath((Join-Path $file.DirectoryName $path))
        $linkCount++
        if (-not (Test-Path -LiteralPath $resolved)) {
            $errors.Add("Broken link in $($file.Name): $target")
        }
    }
    $expected = 0
    $lineNumber = 0
    foreach ($line in ($content -split '\r?\n')) {
        $lineNumber++
        if ($line.StartsWith('|')) {
            if ($line.Contains('\|')) {
                $errors.Add("Escaped pipe requires manual review: $($file.Name):$lineNumber")
            }
            $count = ([regex]::Matches($line, '\|')).Count
            if ($expected -eq 0) {
                $expected = $count
                $tableCount++
            } elseif ($expected -ne $count) {
                $errors.Add("Table width mismatch: $($file.Name):$lineNumber")
            }
        } else {
            $expected = 0
        }
    }
    if ($file.DirectoryName -eq (Join-Path $docs 'ssot') -and $file.Name -match '^0[1-6]-.*\.md$' -and $content -notmatch '`FROZEN`') {
        $errors.Add("Missing FROZEN status: $($file.Name)")
    }
    if ($file.DirectoryName -eq $contracts) {
        if ($content -notmatch 'DERIVED CONTRACT') {
            $errors.Add("Missing DERIVED CONTRACT status: $($file.Name)")
        }
        if ($file.Name -match '^(0[1-5])-.*\.md$') {
            $prefix = 'C' + $Matches[1] + '-'
            if ($content -notmatch '\.\./docs/ssot/') {
                $errors.Add("Missing SSOT source: $($file.Name)")
            }
            $ids = [regex]::Matches($content, '`(C\d{2}-\d{2})`')
            if ($ids.Count -eq 0) {
                $errors.Add("Missing contract clauses: $($file.Name)")
            }
            foreach ($id in $ids) {
                $value = $id.Groups[1].Value
                if (-not $value.StartsWith($prefix)) {
                    $errors.Add("Wrong contract prefix: $($file.Name): $value")
                }
                if (-not $contractIds.Add($value)) {
                    $errors.Add("Duplicate contract clause: $value")
                }
            }
        }
    }
}
$policyCount = 0
$caseCount = 0
foreach ($file in @(Get-ChildItem -LiteralPath $contracts -Filter '*.json')) {
    $expectedSource = [IO.Path]::ChangeExtension($file.Name, '.md')
    $sourcePath = Join-Path $contracts $expectedSource
    if (-not (Test-Path -LiteralPath $sourcePath -PathType Leaf)) {
        $errors.Add("Machine contract has no Markdown source: $($file.Name)")
        continue
    }
    try {
        $contract = Get-Content -LiteralPath $file.FullName -Encoding UTF8 -Raw | ConvertFrom-Json
    } catch {
        $errors.Add("Invalid JSON contract: $($file.Name): $($_.Exception.Message)")
        continue
    }
    if ($contract.version -ne 1 -or $contract.source -cne $expectedSource) {
        $errors.Add("Invalid contract version or source: $($file.Name)")
    }
    $sourceText = Get-Content -LiteralPath $sourcePath -Encoding UTF8 -Raw
    $sourceIds = @([regex]::Matches($sourceText, '(?m)^- `(C\d{2}-\d{2})`') | ForEach-Object { $_.Groups[1].Value })
    $covered = [Collections.Generic.HashSet[string]]::new()
    $prefix = 'C' + $file.Name.Substring(0, 2) + '-'
    foreach ($id in @($contract.manualClauses)) {
        if ($id -isnot [string] -or -not $id.StartsWith($prefix) -or -not $covered.Add($id)) {
            $errors.Add("Invalid or duplicate manual clause: $($file.Name): $id")
        }
    }
    $policyIds = [Collections.Generic.HashSet[string]]::new()
    if (@($contract.policies).Count -eq 0) {
        $errors.Add("No policies in $($file.Name)")
    }
    foreach ($policy in @($contract.policies)) {
        $policyCount++
        if ($policy.id -isnot [string] -or $policy.id -notmatch '^[a-z][a-z0-9-]+$' -or -not $policyIds.Add($policy.id)) {
            $errors.Add("Invalid or duplicate policy id in $($file.Name)")
        }
        if (@($policy.clauses).Count -eq 0) {
            $errors.Add("No clauses for policy $($policy.id)")
        }
        foreach ($id in @($policy.clauses)) {
            if ($id -isnot [string] -or -not $id.StartsWith($prefix) -or $id -notin $sourceIds) {
                $errors.Add("Unknown clause for $($policy.id): $id")
            } else {
                [void]$covered.Add($id)
            }
        }
        $facts = @($policy.facts)
        if ($facts.Count -eq 0 -or $null -eq $policy.allowWhen) {
            $errors.Add("Missing facts or allowWhen for $($policy.id)")
            continue
        }
        if (@($facts | Select-Object -Unique).Count -ne $facts.Count) {
            $errors.Add("Duplicate facts for $($policy.id)")
        }
        foreach ($fact in $facts) {
            if ($fact -isnot [string] -or $fact -cnotmatch '^[a-z][a-zA-Z0-9]*$' -or
                $null -eq $policy.allowWhen.PSObject.Properties[$fact] -or
                $policy.allowWhen.$fact -isnot [bool]) {
                $errors.Add("Invalid fact or condition in $($policy.id): $fact")
            }
        }
        foreach ($condition in $policy.allowWhen.PSObject.Properties) {
            if ($condition.Name -cnotin $facts -or $condition.Value -isnot [bool]) {
                $errors.Add("Unexpected condition for $($policy.id): $($condition.Name)")
            }
        }
        $hasAllow = $false
        $hasDeny = $false
        if (@($policy.cases).Count -lt 2) {
            $errors.Add("Policy needs positive and negative cases: $($policy.id)")
        }
        $caseNames = [Collections.Generic.HashSet[string]]::new()
        foreach ($testCase in @($policy.cases)) {
            $caseCount++
            if ($testCase.name -isnot [string] -or -not $caseNames.Add($testCase.name) -or
                $testCase.allowed -isnot [bool] -or $null -eq $testCase.input) {
                $errors.Add("Invalid test case in $($policy.id)")
                continue
            }
            $validInput = $true
            foreach ($fact in $facts) {
                if ($null -eq $testCase.input.PSObject.Properties[$fact] -or $testCase.input.$fact -isnot [bool]) {
                    $validInput = $false
                }
            }
            foreach ($property in $testCase.input.PSObject.Properties) {
                if ($property.Name -cnotin $facts) {
                    $validInput = $false
                }
            }
            if (-not $validInput) {
                $errors.Add("Invalid input for $($policy.id): $($testCase.name)")
                continue
            }
            $actual = $true
            foreach ($fact in $facts) {
                if ($testCase.input.$fact -cne $policy.allowWhen.$fact) {
                    $actual = $false
                }
            }
            if ($actual -ne $testCase.allowed) {
                $errors.Add("Failed contract case $($policy.id): $($testCase.name)")
            }
            if ($testCase.allowed) { $hasAllow = $true } else { $hasDeny = $true }
        }
        if (-not $hasAllow -or -not $hasDeny) {
            $errors.Add("Policy lacks allow or deny case: $($policy.id)")
        }
    }
    foreach ($id in $sourceIds) {
        if (-not $covered.Contains($id)) {
            $errors.Add("Uncovered Markdown clause: $($file.Name): $id")
        }
    }
    foreach ($id in $covered) {
        if ($id -notin $sourceIds) {
            $errors.Add("Unknown machine clause: $($file.Name): $id")
        }
    }
}
foreach ($source in @(Get-ChildItem -LiteralPath $contracts -Filter '0*.md')) {
    if (-not (Test-Path -LiteralPath ([IO.Path]::ChangeExtension($source.FullName, '.json')) -PathType Leaf)) {
        $errors.Add("Missing machine contract for $($source.Name)")
    }
}
if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Output $_ }
    exit 1
}
Write-Output "PASS: $($required.Count) required entries, $linkCount local links, $tableCount tables, $policyCount parsed policies, $caseCount executed cases."
