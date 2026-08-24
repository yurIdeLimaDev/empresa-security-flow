[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidateSet('corpus', 'index', 'rebuild', 'search')]
    [string]$Command,
    [Parameter(Position = 1)]
    [string]$Question
)

$ErrorActionPreference = 'Stop'
$integrationRoot = Split-Path -Parent $PSCommandPath
$python = Join-Path $integrationRoot '.runtime\.venv\Scripts\python.exe'
if (-not (Test-Path -LiteralPath $python)) { throw 'Runtime ausente. Execute .\bootstrap.ps1 primeiro.' }
if ($Command -eq 'search' -and [string]::IsNullOrWhiteSpace($Question)) { throw 'Informe a pergunta após search.' }

$cli = Join-Path $integrationRoot 'scripts\project_rag.py'
if ($Command -eq 'search') {
    & $python $cli search $Question
} else {
    & $python $cli $Command
}
exit $LASTEXITCODE
