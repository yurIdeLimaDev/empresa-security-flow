# Runner dos Pipelines 1 e 2

Estado local em 13/09/2026: [consolidação](../docs/ESTADO_ATUAL.md).
Novos ensaios em `scripts/mvp_acceptance.py`; aceite de evidências do host em
`scripts/host_acceptance.py`. A entrega recebe `approved-security.patch` gerado
na finalização e ligado por hash à autorização; omitir `--patch-root` no
comando `delivery-package`. Patches do adaptador manual continuam sendo insumos
da correção, nunca fonte direta do pacote final.

O runner é escrito em Go. Planejamento não gera tráfego. Execução real exige
Linux root, Docker e todas as políticas/supply chain aprovadas.

Validação local e comandos reproduzíveis de CI:
[VALIDACAO_LOCAL_MVP.md](../docs/VALIDACAO_LOCAL_MVP.md). A checagem de conteúdo
antes do gateway complementa a autorização de envio; não escolhe fornecedor,
não anonimiza código e não substitui revisão de fontes.

## Modelo de decisão

As decisões humanas são registradas antes da execução:

- `engagement-policy.json`: destino, IP, porta, taxa, budget, autorização e
  canaries;
- `tool-policy.json`: modo, pin, duração, proibições e flags;
- `test-matrix.json`: recursos/identidades permitidos para teste ativo de
  dados;
- `scope.json` ou `lead.json`: entrada do caso e caminhos das políticas.

O gate válido vem exclusivamente dos arquivos front-loaded; as antigas flags
de confirmação, que não alteravam a decisão, foram removidas.

## Controles

1. schema Draft 2020-12 estrito e validação sem campos desconhecidos;
2. rede Docker exclusiva, IPv6 desligado, `DOCKER-USER` com egress `DROP`;
3. chain `INPUT` bloqueando o acesso do container ao host, exceto as portas
   efêmeras do proxy de auditoria abertas pelo runner;
4. canary negativo controlado pelo operador e canary positivo no escopo;
5. executor Docker-only por digest, revisão, SBOM, lock e timeout;
6. ScopeTransport/proxy com host, porta e IP pinados;
7. log de execução com hashes de políticas e insumos; artefatos sensíveis em
   arquivos `0600`.

O proxy registra HTTP e túneis HTTPS, mas não substitui o firewall. Em HTTPS
sem interceptação TLS, ele vê o `CONNECT`, não cada request interno.

## Automação vigente

Pipeline 1:

- Subfinder, Amass passivo, httpx, testssl, jsluice e Gitleaks;
- TruffleHog apenas em filesystem local, com `--no-verification` injetado;
- dnsReaper automaticamente sobre a lista curada de Subfinder/Amass, com
  `--parallelism 10`; toda saída continua candidata e nunca há claim;
- HIBP automaticamente apenas em `/breachedaccount/{email}` para até 20
  e-mails corporativos publicados nas páginas configuradas. A conta é
  pseudonimizada no artefato; busca por domínio é proibida;
- keyleak-detector desligado.

Pipeline 2, sempre condicionado a escopo/autorização/configuração:

- feroxbuster, Amass ativo, Schemathesis, Hadrian, Nuclei, Gitleaks,
  OSV-Scanner, Semgrep, Trivy, ZAP, jwt_tool, Dalfox e sqlmap;
- Supabase-RLS-Checker e firepwn permanecem manuais/GUI;
- SupaShield, AuthProbe, keyleak ativo e rlsgate permanecem desligados;
- scenario runner próprio somente para crédito/billing com leitura de estado
  antes e depois.

“Automático” descreve o wrapper. A execução continua bloqueada enquanto a
entrada correspondente não passar na cadeia de ferramentas.

## Uso

```powershell
cd pipeline
go vet ./...
go test ./...
go build -o bin/pipeline.exe ./cmd/pipeline

# Sem rede
.\bin\pipeline.exe pipeline1 --config config/examples/lead.json
.\bin\pipeline.exe pipeline2 --config config/examples/scope.json

# Execução: somente em Linux root e com arquivos reais aprovados
./bin/pipeline pipeline1 --config config/lead-real.json --execute
./bin/pipeline pipeline2 --config config/scope-real.json --execute

# Ensaio sem scanner: somente infraestrutura Linux/root descartável e
# endpoints controlados pelo operador. case-dir deve ser novo e vazio.
./bin/pipeline isolation-smoke --engagement-policy config/engagement-smoke.json --case-dir /var/tmp/empresa-security-smoke

# Cadeia, integridade e modelo canônico
.\bin\pipeline.exe supply-chain --lock tools.lock.json --sbom sbom.cdx.json --strict
.\bin\pipeline.exe runtime-check --lock tools.lock.json --sbom sbom.cdx.json --pull --strict --output runtime-readiness.json
.\bin\pipeline.exe verify --case-dir ..\artifacts\CLIENT-001\pipeline-2
.\bin\pipeline.exe normalize --tool hadrian --input hadrian.json --output hadrian.bundle.json --engagement ENG-001 --scope SCOPE-001
.\bin\pipeline.exe consolidate --input hadrian.bundle.json --input zap.bundle.json --output consolidated.json --report consolidated.md
.\bin\pipeline.exe compare --previous anterior.json --current atual.json --output comparacao.json
```

## Correção segura

O fluxo de correção cria um ticket por achado validado e trabalha sempre sobre
a melhor versão retida. A tentativa ocorre em worktree isolado, passa por gates
de qualidade, segurança e reteste, e só é promovida se a cobertura for
comparável e não houver regressão. Tentativas piores são registradas e
descartadas. Existe uma única revisão humana, depois do gate global.

Comandos disponíveis:

```powershell
.\bin\pipeline.exe remediation-preflight --config ..\correcao\config\remediation.json --output remediation-readiness.json --strict
.\bin\pipeline.exe remediation-plan --config ..\correcao\config\remediation.json
.\bin\pipeline.exe remediation-worktree --config ..\correcao\config\remediation.json --plan PLAN --ticket REM-0123456789AB
.\bin\pipeline.exe remediation-agent --config ..\correcao\config\remediation.json --plan PLAN --ticket REM-0123456789AB
.\bin\pipeline.exe remediation-gates --config ..\correcao\config\remediation.json --plan PLAN --ticket REM-0123456789AB --phase candidate --candidate-ref COMMIT --worktree WORKTREE
.\bin\pipeline.exe remediation-evaluate --config ..\correcao\config\remediation.json --plan PLAN --ticket REM-0123456789AB --candidate-ref COMMIT --candidate-bundle BUNDLE --gate-run GATE_RUN
.\bin\pipeline.exe remediation-finalize --config ..\correcao\config\remediation.json --plan PLAN --approval APPROVAL --output DELIVERY
```

O contrato completo está em
`../correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md`. Os nomes de adaptadores no JSON
de exemplo são placeholders deliberados: o preflight os recusa, e
`remediation-plan` repete esse gate antes de criar qualquer artefato.

As 20 ferramentas automáticas estão revisadas, no SBOM, pinadas por digest e
passaram no liveness integrado; oito entradas não automáticas ficam
explicitamente restritas. Antes de um caso real ainda é obrigatório trocar
placeholders/IPs, registrar o SOW e executar no host Linux escolhido:

```bash
./scripts/build-runner-linux.sh
./scripts/deploy-preflight-linux.sh /etc/empresa-security/baseline.env /etc/empresa-security/engagement-policy.json /var/lib/empresa-security/preflight-$(date -u +%Y%m%dT%H%M%SZ)
```

O ensaio Linux/root controlado está documentado em
`../validacao/2026-08-22-isolamento-politicas/linux-root-integration/`; o
fechamento da cadeia está em
`../validacao/2026-08-22-fechamento-prontidao-operacional/`.

## Comandos do primeiro caso

- `onboarding-generate`: rascunhos por schema, fail-closed;
- `reference-profile`: perfil Python concreto;
- `remediation-apply-patch`: adaptador manual governado;
- `remediation-run`: geração por gateway configurado, tentativas, gates e
  comparação sobre BEST até a revisão humana final; não aprova nem entrega;
- `reference-gate`: quality, retest ou security-full;
- `delivery-package`: HTML, PDF, manifesto, ZIP e age;
- `backup-create` / `backup-restore`: backup e restauração verificada.

O empacotador roda offline em distroless por digest. Consulte o
[runbook de revisão e entrega](../docs/runbooks/ENTREGA_E_REVISAO.md)
antes da operação, incluindo entradas do renderer e migração das autorizações
antigas para o patch cumulativo vinculado por hash.

## Geração independente de fornecedor — 11/09/2026

O modo `agent.command=["builtin:patch-proposal"]` usa o próprio runner pinado e
`agent.generation`. O gateway/modelo permanecem indefinidos e desativados no
exemplo. Nenhum fornecedor é chamado nos testes; o simulador existe somente
na suíte de testes, nunca como fallback de produção.

Configuração, protocolo, ativação futura e limitações estão em
[GERACAO_PATCHES_SEM_PROVEDOR.md](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).
