[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$integrationRoot = Split-Path -Parent $PSCommandPath
$runtimeRoot = Join-Path $integrationRoot '.runtime'
$sourceRoot = Join-Path $runtimeRoot 'HippoRAG'
$venvRoot = Join-Path $runtimeRoot '.venv'
$upstream = 'https://github.com/OSU-NLP-Group/HippoRAG.git'
$commit = '2f52a86dd04e4633703bd2fb3bb6a37683ac3cfb'

if (-not (Get-Command git -ErrorAction SilentlyContinue)) { throw 'Git não encontrado.' }
if (-not (Get-Command py -ErrorAction SilentlyContinue)) { throw 'Python Launcher (py) não encontrado.' }

New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
if (Test-Path -LiteralPath $sourceRoot) {
    $current = (& git -C $sourceRoot rev-parse HEAD).Trim()
    if ($current -ne $commit) { throw "Fonte HippoRAG incompatível em $sourceRoot. Remova apenas esse diretório e execute novamente." }
} else {
    & git clone --no-checkout $upstream $sourceRoot
    if ($LASTEXITCODE -ne 0) { throw 'Falha ao clonar HippoRAG.' }
    & git -C $sourceRoot checkout --detach $commit
    if ($LASTEXITCODE -ne 0) { throw 'Falha ao fixar o commit HippoRAG.' }
    $current = (& git -C $sourceRoot rev-parse HEAD).Trim()
    if ($current -ne $commit) { throw 'A origem HippoRAG não corresponde ao commit fixado.' }
}

$python = Join-Path $venvRoot 'Scripts\python.exe'
if (-not (Test-Path -LiteralPath $python)) {
    & py -3.12 -m venv $venvRoot
    if ($LASTEXITCODE -ne 0) { throw 'Falha ao criar o ambiente Python 3.12.' }
}

& $python -m pip install --disable-pip-version-check --require-hashes -r (Join-Path $integrationRoot 'requirements.lock')
if ($LASTEXITCODE -ne 0) { throw 'Falha ao instalar dependências fixadas.' }
& $python -m pip install --disable-pip-version-check --no-build-isolation --no-deps $sourceRoot
if ($LASTEXITCODE -ne 0) { throw 'Falha ao instalar HippoRAG a partir do commit fixado.' }
& $python (Join-Path $integrationRoot 'scripts\verify_runtime.py')
if ($LASTEXITCODE -ne 0) { throw 'A verificação do runtime falhou.' }
