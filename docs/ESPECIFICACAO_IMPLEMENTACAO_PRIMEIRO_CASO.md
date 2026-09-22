# Especificação técnica — primeiro caso controlado

Nota consolidada em 13/09/2026: o escopo manual deste documento é o registro do primeiro
caso de agosto. Por decisão posterior, foi implementada a camada de geração
independente de fornecedor e execução até revisão final. Provedor/modelo
seguem indefinidos; consulte
[a extensão vigente](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).

## Objetivo

Tornar o motor executável para uma única stack de referência e um laboratório
próprio. Esta entrega não executa alvos de terceiros, não ativa HIBP/OAST e não
altera o fluxo: Pipeline 1 -> onboarding confirmado -> Pipeline 2 -> bundle
validado -> correção sobre BEST -> gate humano final -> entrega.

## Regras invioláveis

- Preservar schemas, isolamento, hashes, gates e a regra de não regressão.
- Não criar bypass para preflight, agente, gates, avaliação ou revisão final.
- Não executar ferramenta fora do executor governado nem baixar dependência em
  runtime.
- Todo binário, container ou ferramenta novo precisa de pin, hash/digest, SBOM,
  revisão de adoção e testes. `latest` é proibido.
- Não registrar credenciais, e-mails, evidência bruta, worktrees, pacotes de
  entrega ou casos de clientes no Git.
- O laboratório vulnerável deve ser repositório privado separado.

## 1. Perfil de correção para uma stack

Criar somente um perfil de referência em `correcao/adapters/<stack-id>/`, com
README, versão, hashes e testes. Não generalizar para múltiplas stacks.

### Adaptador manual governado

Implementar adaptador usado como `remediation-agent`; não pular essa etapa.
O runner atual exige agente pinado e confirma que o commit candidato foi
produzido por ele.

O adaptador recebe `REMEDIATION_CASE_ID`, `REMEDIATION_TICKET_ID`,
`REMEDIATION_PROMPT_PATH`, `REMEDIATION_WORKTREE` e um diretório externo de
patches explicitamente permitido na `environment_allowlist`.

Ele deve:

1. aceitar somente `<patch-root>/<case-id>/<ticket-id>.patch`;
2. resolver caminhos e recusar path traversal ou destino fora de `patch-root`;
3. conferir o worktree recebido;
4. executar `git apply --check`, depois aplicar o patch;
5. registrar SHA-256, caminho relativo e resultado no diretório controlado da
   tentativa;
6. não acessar rede, instalar dependências, fazer checkout/reset ou commit; o
   runner continua responsável pelo commit;
7. falhar se patch estiver ausente, não aplicar limpo ou sair do escopo.

O adaptador precisa passar no `remediation-preflight` com hash real.

### Gates obrigatórios

| Gate | Implementação |
|---|---|
| `quality` | Build, lint e testes em container por digest; dependências apenas pelo lockfile e scripts desabilitados quando compatível; resultado estruturado. |
| `security-full` | Chamar somente o runner e scanners já aprovados para repositório autorizado; normalizar e consolidar `candidate-bundle.json` canônico. |
| `retest` | Reproduzir a propriedade de segurança do ticket: identidades A/B para autorização, request/parâmetro capturado para injeção, estado final para configuração. Hash de resposta isolado não basta. |

Todos os gates devem rejeitar artefato ausente, saída fora do diretório da
tentativa, hash divergente, timeout, cobertura não comparável, achado novo ou
aumento de severidade/rank.

### Testes do perfil

Cobrir patch válido/inválido, path traversal, hash divergente, alteração fora
do escopo, gate falho, reteste insuficiente, achado novo e promoção apenas do
candidato melhor. Sem rede ou segredos.

## 2. Laboratório de ensaio ponta a ponta

Criar em repositório privado separado uma aplicação descartável da stack de
referência, com dados fictícios e contas A/B. Incluir achados intencionais,
reversíveis e representativos: falha de autorização, configuração de dados
incorreta quando aplicável, segredo fictício em bundle, header ausente e regra
de autenticação/token fraca.

O laboratório não pode ficar exposto à internet. Executar o ciclo completo:

1. Pipeline 1 com política do laboratório;
2. onboarding e autorização explícita;
3. Pipeline 2 somente nos módulos/recursos autorizados;
4. normalização, consolidação e relatório;
5. patch pelo adaptador manual;
6. gates por ticket, comparação contra BEST e gate global;
7. aprovação final simulada vinculada a commit/hash;
8. pacote de entrega.

Preservar no Git apenas evidência saneada em `validacao/`. O resultado precisa
ser reproduzível e listar qualquer falha encontrada no fluxo.

## 3. Host Linux e deploy reproduzível

Implementar material de deploy, sem contratar VPS nem executar contra clientes.

- Baseline versionado para host Linux dedicado: SSH por chave, usuário
  operacional, atualizações de segurança, Docker, iptables e raiz exclusiva de
  casos.
- Script idempotente de verificação de drift.
- Execução obrigatória de `pipeline/scripts/deploy-preflight-linux.sh` antes de
  cada caso, com evidência em diretório novo.
- Testes de bloqueio para ausência de root, Docker, chain `DOCKER-USER`, runner
  ou política válida.
- Stop/teardown preserva logs e evidência antes de remover containers, redes e
  chains.
- Armazenamento e backup criptografados; chave fora do host; teste documentado
  de restauração. Backup sem restauração testada não é concluído.

O ensaio Linux existente prova o desenho; todo host escolhido precisa executar
seu próprio preflight.

## 4. Onboarding técnico

Implementar CLI ou gerador schema-driven que produza rascunhos de `lead`,
`scope`, `engagement-policy`, `tool-policy`, `test-matrix` e configuração de
remediação. Ele deve pedir e validar domínio/base URL, IP:porta, taxa, download,
APIs externas, DNS, canaries, recursos/identidades de teste, comandos e hashes
de adaptadores.

Não gerar valores fictícios, não confirmar autorização, não preencher SOW e não
habilitar módulos sem entrada explícita. Todo rascunho deve passar pelos schemas
existentes e permanecer bloqueado sem escopo/autorização requeridos.

## 5. Relatório e entrega determinísticos

Implementar a partir do bundle canônico validado:

- HTML versionado e PDF em container por digest;
- escopo, limitações, cobertura, achados saneados, evidência, recomendação,
  reteste, commit, hash do bundle e não regressão;
- SHA-256 do PDF e manifesto;
- pacote com PDF, bundle permitido, patches e `SHA256SUMS`;
- criptografia autenticada por chave pública do destinatário (`age`, PGP ou
  CMS). 7z com senha é apenas fallback documentado, por canal separado.

Testar renderização, determinismo do manifesto e rejeição de bundle, commit ou
hash incompatíveis.

## 6. Runbooks e observabilidade mínima

Criar `ONBOARDING_TECNICO.md`, `EMERGENCIA.md`, `FERRAMENTAS_MANUAIS.md` e
`OPERACAO_HOST.md`. Devem cobrir entradas, preflight, stop, preservação de
evidência, teardown, restauração, rotação de segredo e limpeza de ferramentas
manuais. Executá-los no laboratório.

Monitorar disco, falha de backup, containers residuais e falha de preflight.
Alertas não podem conter segredo, e-mail, URL privada, evidência ou case ID.

## Critérios de aceite

1. Um perfil possui agente manual e três gates pinados/aprovados.
2. Preflight rejeita adaptador/gate ausente, hash zero ou divergente.
3. O laboratório executa o ciclo inteiro sem rede pública nem dados de
   terceiros.
4. Patch válido só é promovido se gates e comparação canônica provarem ausência
   de regressão; patch inválido, reteste insuficiente ou achado novo preserva
   BEST.
5. Relatório/pacote vêm do bundle canônico e todos os hashes conferem.
6. Preflight, backup e restauração foram exercitados em Linux de ensaio.
7. Testes passam localmente e em CI sem segredos ou rede de terceiros.
8. Arquitetura e decisões documentam somente o que foi implementado.

## Fora de escopo

- VPS, domínio, DNS, e-mail, marketing, preço, contrato e base legal.
- HIBP, OAST público ou chamadas contra terceiros.
- Múltiplas stacks, agente LLM remoto e mudanças na landing page.
- Dados, casos ou artefatos de clientes no repositório.

## Estado da implementação — 2026-08-24

Implementado: perfil único `python-3.13-stdlib`, adaptador manual governado,
gates quality/retest/security-full, onboarding fail-closed, baseline e drift
Linux, preservação/teardown, monitoramento, backup/restauração age, relatório
HTML/PDF e pacote determinístico criptografado.

O laboratório privado executou cinco correções fictícias, gate global sem
regressão, aprovação final simulada e entrega. Testes Go, race detector, vet,
testes Python e CI privada passaram. A evidência saneada está em
`validacao/2026-08-24-primeiro-caso-controlado/`.

Condição de deploy: repetir o preflight como root no Linux dedicado escolhido.
O ensaio em contêiner validou runner, empacotamento e restauração, mas não
substitui a inspeção de SSH, updates e firewall desse host.

## Evolução posterior do contrato de entrega

Esta especificação preserva o primeiro perfil manual. O `patch-root` daquele
adaptador é entrada de proposta, não entrada do empacotador. Desde 13/09,
a entrega usa `approved-security.patch` derivado de baseline até BEST, com
hash na autorização final. Consulte o [estado atual](ESTADO_FLUXO.md) para
uso vigente; não executar instrução histórica de entrega externa.
