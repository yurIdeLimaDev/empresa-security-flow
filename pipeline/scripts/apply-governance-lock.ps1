[CmdletBinding()]
param(
  [string]$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
)

$ErrorActionPreference = 'Stop'
$lockPath = Join-Path $RepositoryRoot 'pipeline\tools.lock.json'
$reviewRelative = '..\docs\REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md'
$reviewPath = Join-Path $RepositoryRoot 'docs\REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md'
$reviewHash = (Get-FileHash -LiteralPath $reviewPath -Algorithm SHA256).Hash.ToLowerInvariant()
$reviewedAt = '2026-08-22T22:30:00Z'

$automatic = @(
  'subfinder', 'amass', 'httpx', 'gitleaks', 'trufflehog', 'testssl', 'nuclei',
  'zap', 'schemathesis', 'jwt_tool', 'sqlmap', 'dalfox', 'osv-scanner', 'trivy',
  'semgrep', 'feroxbuster', 'hibp', 'hadrian', 'jsluice', 'dnsreaper'
)
$manual = @('supabase-rls-checker', 'firepwn')
$infrastructure = @('interactsh-server')

$containers = @{
  'subfinder' = [ordered]@{
    image = 'gcr.io/distroless/static-debian12:nonroot'
    digest = 'sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab'
    entrypoint = '/opt/tool/subfinder'
    runtime_artifacts = @([ordered]@{path='tool-artifacts/linux-amd64/subfinder'; sha256='sha256:f874b5463d1d1e338d3f0312b93464d7ada609dbfd3bc180daaeca1fc24a3ceb'; mount_path='/opt/tool/subfinder'})
  }
  'testssl' = [ordered]@{image='drwetter/testssl.sh:3.2'; digest='sha256:953f4080fdb053b44b4810ff32c01b9f29d8e91867b8da35220c8fbb54d08f6e'}
  'sqlmap' = [ordered]@{
    image = 'python:3.13-slim-bookworm'
    digest = 'sha256:00faa2debb87529f9f0764e9491d8ba400a3678976616c3bd7cb193745ac20d1'
    entrypoint = '/usr/local/bin/python3'
    command_prefix = @('/opt/tool/sqlmap.pyz')
    runtime_artifacts = @([ordered]@{path='tool-artifacts/linux-amd64/sqlmap.pyz'; sha256='sha256:66054f4518c59fefccbdf59ebabc91cafc6d52a62fb6b8542a9cf6e99790e78a'; mount_path='/opt/tool/sqlmap.pyz'})
  }
  'feroxbuster' = [ordered]@{image='epi052/feroxbuster'; digest='sha256:5cab630e43eface8cc79a6d32f8899dbe6f708f1260b0c23f4efda7c1f521042'}
  'hadrian' = [ordered]@{
    image = 'gcr.io/distroless/static-debian12:nonroot'
    digest = 'sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab'
    entrypoint = '/opt/tool/hadrian'
    runtime_artifacts = @([ordered]@{path='tool-artifacts/linux-amd64/hadrian'; sha256='sha256:20345648122283b4ca984181f64a89f525b37b3038148fdc20851bffab47c8da'; mount_path='/opt/tool/hadrian'})
  }
  'jsluice' = [ordered]@{
    image = 'debian:trixie-slim'
    digest = 'sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258'
    entrypoint = '/opt/tool/jsluice'
    runtime_artifacts = @([ordered]@{path='tool-artifacts/linux-amd64/jsluice'; sha256='sha256:49640235e94e96066cbbf13de3f5738331c4048083b134a9d7e68e2931eb8b21'; mount_path='/opt/tool/jsluice'})
  }
}

$liveness = @{
  'subfinder' = [ordered]@{command=@('-version'); expected_pattern='(?i)2\.15\.0'}
  'amass' = [ordered]@{command=@('-version'); expected_pattern='(?i)5\.1\.1'}
  'httpx' = [ordered]@{command=@('-version'); expected_pattern='(?i)1\.10\.0'}
  'gitleaks' = [ordered]@{command=@('version'); expected_pattern='(?i)8\.30\.1'}
  'trufflehog' = [ordered]@{command=@('--version'); expected_pattern='(?i)3\.97\.0'}
  'testssl' = [ordered]@{command=@('--version'); expected_pattern='(?i)3\.2\.4'}
  'nuclei' = [ordered]@{command=@('-version'); expected_pattern='(?i)3\.11\.1'}
  'zap' = [ordered]@{command=@('-cmd', '-silent', '-version'); expected_pattern='(?i)2\.17\.0'}
  'schemathesis' = [ordered]@{command=@('--version'); expected_pattern='(?i)4\.24\.3'}
  'jwt_tool' = [ordered]@{command=@('-h'); expected_pattern='(?i)jwt[_ -]?tool|usage'}
  'sqlmap' = [ordered]@{command=@('--version'); expected_pattern='(?i)1\.10'}
  'dalfox' = [ordered]@{command=@('--version'); expected_pattern='(?i)3\.2\.1'}
  'osv-scanner' = [ordered]@{command=@('--version'); expected_pattern='(?i)2\.5\.1'}
  'trivy' = [ordered]@{command=@('--version'); expected_pattern='(?i)0\.74\.0'}
  'semgrep' = [ordered]@{command=@('--version', '--disable-version-check', '--metrics=off'); expected_pattern='(?i)1\.173\.0'}
  'feroxbuster' = [ordered]@{command=@('--version'); expected_pattern='(?i)2\.13\.1'}
  'hibp' = [ordered]@{command=@('--version'); expected_pattern='(?i)curl 8\.15\.0'}
  'hadrian' = [ordered]@{command=@('version'); expected_pattern='(?i)1\.0\.0'}
  'jsluice' = [ordered]@{command=@('--help'); expected_pattern='(?i)jsluice'}
  'dnsreaper' = [ordered]@{command=@('--help'); expected_pattern='(?i)dnsreaper|usage'}
}

$entrypointOverrides = @{
  'zap' = '/zap/zap.sh'
  'semgrep' = '/usr/bin/semgrep'
  'dalfox' = '/app/dalfox'
}

$lock = Get-Content -LiteralPath $lockPath -Raw | ConvertFrom-Json
$lock.schema_version = '1.1.0'
$lock.generated_at = $reviewedAt

foreach ($tool in $lock.tools) {
  if ($automatic -contains $tool.name) {
    $class = 'automatic'
    $status = 'approved'
    $scope = 'adoption review: pinned source, dependencies, install/update paths, governed adapter, network/destructive modes and runtime liveness'
  } elseif ($manual -contains $tool.name) {
    $class = 'manual'
    $status = 'restricted'
    $scope = 'adoption review: manual-only item; executor prohibition and documented confirmation requirements'
  } elseif ($infrastructure -contains $tool.name) {
    $class = 'infrastructure'
    $status = 'restricted'
    $scope = 'adoption review: separately operated infrastructure; not executable by the scanner runner'
  } else {
    $class = 'disabled'
    $status = 'restricted'
    $scope = 'adoption review: disabled item; executor prohibition and documented reason'
  }

  $tool | Add-Member -NotePropertyName execution_class -NotePropertyValue $class -Force
  $tool.code_review = [ordered]@{
    status = $status
    reviewer = 'Codex assisted adoption review'
    reviewed_at = $reviewedAt
    reviewed_commit = $tool.release.commit
    reference = $reviewRelative.Replace('\', '/')
    reference_sha256 = "sha256:$reviewHash"
    scope = $scope
  }
  if ($status -eq 'approved') {
    $tool | Add-Member -NotePropertyName liveness -NotePropertyValue $liveness[$tool.name] -Force
  } elseif ($tool.PSObject.Properties.Name -contains 'liveness') {
    $tool.PSObject.Properties.Remove('liveness')
  }
  if ($containers.ContainsKey($tool.name)) {
    $tool | Add-Member -NotePropertyName container -NotePropertyValue $containers[$tool.name] -Force
  }
  if ($entrypointOverrides.ContainsKey($tool.name)) {
    $tool.container | Add-Member -NotePropertyName entrypoint -NotePropertyValue $entrypointOverrides[$tool.name] -Force
  }
  if ($tool.name -eq 'semgrep') {
    $tool.container | Add-Member -NotePropertyName environment -NotePropertyValue ([ordered]@{SEMGREP_SEND_METRICS='off'; EIO_BACKEND='posix'}) -Force
  }
}

$json = $lock | ConvertTo-Json -Depth 30
[IO.File]::WriteAllText($lockPath, $json + [Environment]::NewLine, [Text.UTF8Encoding]::new($false))
