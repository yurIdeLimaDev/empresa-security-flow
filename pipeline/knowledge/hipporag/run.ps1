[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidateSet('corpus', 'index', 'rebuild', 'ask')]
    [string]$Command,
    [Parameter(Position = 1)]
    [string]$Question
)

$ErrorActionPreference = 'Stop'
$integrationRoot = Split-Path -Parent $PSCommandPath
$python = Join-Path $integrationRoot '.runtime\.venv\Scripts\python.exe'
if (-not (Test-Path -LiteralPath $python)) { throw 'Runtime ausente. Execute .\bootstrap.ps1 primeiro.' }
if ($Command -eq 'ask' -and [string]::IsNullOrWhiteSpace($Question)) { throw 'Informe a pergunta após ask.' }
if ($Command -ne 'corpus' -and [string]::IsNullOrWhiteSpace($env:OPENAI_API_KEY)) { throw 'Defina OPENAI_API_KEY somente nesta sessão antes de indexar, reconstruir ou consultar.' }

$cli = Join-Path $integrationRoot 'scripts\project_rag.py'
if ($Command -eq 'ask') {
    & $python $cli ask $Question
} else {
    & $python $cli $Command
}
exit $LASTEXITCODE
