[CmdletBinding()]
param(
  [string]$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
)

$ErrorActionPreference = 'Stop'
$lockPath = Join-Path $RepositoryRoot 'pipeline\tools.lock.json'
$sbomPath = Join-Path $RepositoryRoot 'pipeline\sbom.cdx.json'
$lock = Get-Content -LiteralPath $lockPath -Raw | ConvertFrom-Json
$sbom = Get-Content -LiteralPath $sbomPath -Raw | ConvertFrom-Json
$sbom.metadata.timestamp = '2026-08-22T22:30:00Z'
$sbom.metadata.component.version = '0.5.0'

$byReference = @{}
foreach ($tool in $lock.tools) {
  $byReference[$tool.sbom_component_ref] = $tool
}

foreach ($component in $sbom.components) {
  if (-not $byReference.ContainsKey($component.'bom-ref')) {
    continue
  }
  $tool = $byReference[$component.'bom-ref']
  $properties = @(
    [ordered]@{name='empresa-security:execution-class'; value=$tool.execution_class},
    [ordered]@{name='empresa-security:review-status'; value=$tool.code_review.status},
    [ordered]@{name='empresa-security:source-commit'; value=$tool.release.commit}
  )
  if ($tool.container) {
    $properties += [ordered]@{name='empresa-security:container'; value=($tool.container.image + '@' + $tool.container.digest)}
    foreach ($artifact in @($tool.container.runtime_artifacts)) {
      $properties += [ordered]@{name='empresa-security:runtime-artifact'; value=($artifact.path + '@' + $artifact.sha256)}
    }
  }
  $component | Add-Member -NotePropertyName properties -NotePropertyValue $properties -Force
}

$json = $sbom | ConvertTo-Json -Depth 30
[IO.File]::WriteAllText($sbomPath, $json + [Environment]::NewLine, [Text.UTF8Encoding]::new($false))
