# Validação local do MVP

Atualização: 13/09/2026. Este roteiro reaproveita as suítes existentes, acrescenta
matriz de gates, POC sintética e ensaio integrado, e reúne os comandos da CI. Não usa modelo
de IA, fornecedor de assinatura ou conta de cliente. O transporte externo dos
testes do Worker é substituído por respostas controladas; o teste HTTPS do
gerador usa servidor local.

## 1. Preparação

Para reproduzir a bateria completa e guardar somente resumos saneados, executar
na pasta `pipeline/`:

```text
python scripts/collect_local_checks.py --output ../validacao/CHECAGEM-NOVA
```

No Windows, informar `--bash "C:\Program Files\Git\bin\bash.exe"` para a
checagem de sintaxe dos scripts. Isso não executa firewall/Docker e não aprova
Linux. O script não reinstala dependências nem provisiona recursos; executar
`npm ci --ignore-scripts` na landing antes, como descrito abaixo.

- Motor: Go `1.26.5`, Git e Python 3.10 ou superior para a checagem documental.
  O detector de corrida do Go requer compilador C compatível no Windows.
- Landing: Node `24.18.0`, fixado em `landing-page/.node-version`.
- Dependências: `go.mod`/`go.sum` e `package-lock.json` existentes. Não instalar
  ferramentas extras nem atualizar versões automaticamente durante validação.
- Os comandos abaixo usam `python3`; no Windows pode ser necessário chamar
  o caminho do Python instalado. Não há pacote pip a instalar para eles.
- `go test` usa repositórios temporários fictícios. As saídas de build da
  landing ficam em `dist/`, ignorada pelo Git. Credenciais reais não são insumo.

## 2. Motor e documentos

Na pasta `pipeline/`:

```text
go mod verify
go vet ./...
go test ./... -count=1 -timeout=10m
go test -race ./... -count=1 -timeout=10m
go run ./cmd/pipeline supply-chain --lock tools.lock.json --sbom sbom.cdx.json --strict
python3 -m unittest discover -s scripts/tests -p "test_*.py" -v
python3 scripts/check_docs.py
python3 scripts/mvp_acceptance.py --output ../validacao/ENSAIO-NOVO
python3 -m unittest discover -s knowledge/hipporag/tests -p "test_build_corpus.py" -v
python3 knowledge/hipporag/scripts/build_corpus.py
```

`supply-chain` verifica lock, políticas e SBOM; não equivale a executar os
containers. `build_corpus.py` verifica o corpus desta cópia sem chamar modelo
ou rede; não reconfigura o MCP instalado para apontar a esta worktree.

O verificador documental cobre os índices e arquivos atuais enumerados em
`scripts/check_docs.py`, além dos diretórios públicos de docs, runbooks e
modelos em `negocio/juridico/`. Confere UTF-8, título e destinos de links inline
e definições de referências. Ignora exemplos em código, URLs e âncoras; não aprova cláusulas,
documentos preenchidos ou assinatura. Não é uma auditoria de todos os arquivos.

## 3. Landing

Na pasta `landing-page/`, com o Node indicado:

```text
npm ci --ignore-scripts
npm exec -- tsc --noEmit
npm run lint
npm test
npm audit --audit-level=moderate
```

`npm ci` instala o lockfile e exige acesso ao registro. `npm audit` consulta
avisos conhecidos de todas as dependências, inclusive desenvolvimento; não
mede sozinho a segurança de toda a aplicação.
`npm test` faz build e testa o Worker exportado; uma segunda etapa de build
imediatamente depois é redundante. Deploy continua sendo operação separada.

## 4. Ensaio sintético e limites de prova

As verificações abaixo já fazem parte de `go test`. Para investigar somente
esta sequência, usar, dentro de `pipeline/`:

```text
go test ./internal/app -run "Test(OnboardingDrafts|Patch|RemediationEndToEnd|EncryptedBackupRestore|Delivery)" -count=1 -timeout=10m
```

| Etapa | Prova local | Não comprovado por este teste |
| --- | --- | --- |
| Onboarding | Sem autorização fica bloqueado; mudar somente pagamento para `paid` não libera nem gera política ativa. | Pagamento real, assinatura, titularidade e poderes. |
| Geração | Pedido válido chega ao gerador programado; fonte/metadado com possível credencial bloqueia antes do transporte. | Qualidade de uma IA real e detecção de todos os dados sensíveis. |
| Correção | Gates separados, limites finitos, candidato rejeitado não substitui BEST. | Correção de qualquer aplicação ou todas as classes de falhas. |
| Revisão final | Automação termina em `awaiting_final_review`; aprovação final é uma entrada distinta. | Aprovação de uma pessoa real para cliente real. |
| Entrega | Pacote vinculado a hashes; adulteração recusada. | Recebimento pelo cliente e canal externo de entrega. |
| Recuperação | Backup age com round-trip e integridade local. | Backup off-host, custódia da chave e restore no VPS. |

O novo `TestMVPFirstCustomerRehearsal` conecta os modelos contratuais fictícios,
SOW, autorização e pagamento declarados como simulação, plano de verificação,
bundle sintético, geração, revisão simulada, entrega e backup/restore.
Scanners, serviços externos e decisões reais não são simulados como homologação.
O [kit de avaliação](../correcao/avaliacao/README.md) mede também a rejeição de
regressões e compara resultados da mesma suíte, sem escolher/custear IA.

## 5. CI e publicação

- CI agregada: `.github/workflows/ci.yml`; CI da landing separada:
  `landing-page/.github/workflows/ci.yml`. Incluem audit completo, SBOM npm,
  schemas/exemplos, ensaio sintético e sintaxe dos scripts Linux.
- As actions continuam pinadas por SHA; checkout não mantém credencial Git.
- Node é lido do mesmo arquivo em validação/deploy. Go foi fixado na versão
  testada. Atualizar esses pins requer nova validação.
- CI de uma revisão substituída é cancelada; deploy de produção continua
  manual e serializado pela configuração existente.
- A checagem documental e `go mod verify` foram incorporados à CI agregada.
- Nada neste roteiro cria CI remota aprovada, commit de release ou deploy.

## 6. Passos que dependem do ambiente real

Use os procedimentos existentes, sem criar outro baseline:

- [baseline do host](../deploy/linux/README.md);
- [aceitação e coleta saneada](../deploy/linux/ACEITACAO.md);
- [operação do host e restore](runbooks/OPERACAO_HOST.md);
- [onboarding técnico](runbooks/ONBOARDING_TECNICO.md);
- [emergência](runbooks/EMERGENCIA.md);
- [assinatura e contratação](../negocio/juridico/16_ASSINATURA_E_CONTRATACAO_DIGITAL.md);
- [integração futura do gerador](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).

O host Linux root, seus canaries/firewalls, o fornecedor de IA e a jornada
contratual real continuam pendentes. O ensaio Windows não substitui esses
aceites. Testes devem preservar a etapa humana técnica apenas no final.

## 7. Evidência deste lote

Resultado, limites e arquivos alterados:
[validacao/2026-09-13-lote-completo](../validacao/2026-09-13-lote-completo/RESULTADO.md).

A atualização documental posterior, incluindo inventário ampliado e parser
Mermaid isolado do produto, está em
[validação dos documentos e diagramas](../validacao/2026-09-13-documentacao/RESULTADO.md).
O [índice atual](INDICE_DOCUMENTACAO.md) identifica guias, minutas e históricos.
Não substituir evidência anterior por hashes recalculados de documentos novos.

A API de versão por arquivo usada na CI foi conferida na documentação de
[setup-node](https://github.com/actions/setup-node) e a fixação do Go na de
[setup-go](https://github.com/actions/setup-go). As versões das actions não
foram trocadas por tags móveis.
