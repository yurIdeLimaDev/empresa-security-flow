# Fechamento local prioritário do MVP

Data: 11/09/2026. Ambiente: Windows/amd64, Go 1.26.5, Node 24.18.0,
npm 11.16.0. Escopo: melhorias pequenas e verificáveis na cópia local.

## Resultado

O lote reforça o envio de contexto ao futuro gerador, a CI e a documentação.
O motor e a landing passaram nas verificações locais descritas abaixo. Não é
uma liberação comercial nem uma validação do host definitivo ou de IA real.

GraphRAG consultado primeiro, sem delegação a outro modelo. Os trechos do MCP
não continham as mudanças recentes; código, configuração e documentos exatos
foram conferidos diretamente. O corpus desta worktree foi validado localmente;
isso não reconfigura o MCP conectado nem prova que ele consulta esta cópia.

## Alterações

1. `patch_disclosure.go`: checagem local do JSON efetivamente enviado, inclusive
   conteúdo dos arquivos, metadados, critérios e feedback. Bloqueia a própria
   credencial do gateway e padrões comuns de credenciais antes do transporte.
   Não mascara código nem imprime o conteúdo detectado.
2. Testes negativos para fontes/metadados, JSON escapado, texto legítimo,
   preservação dos bytes e erros sem segredo. Ensaio de automação confirma
   zero chamada ao gateway, preservação de BEST e limites de tentativas.
3. Teste adicional de onboarding: somente `payment_status=paid` não concede
   autorização nem gera política ativa. Nenhum pagamento real foi efetuado.
4. `.dev.vars` e variantes privadas ignoradas pelo Git; exemplos explicitamente
   terminados em `.example` continuam disponíveis. Nenhum segredo foi lido,
   removido, revogado ou inserido nesta etapa.
5. CI: Node 24.18.0 compartilhado pelo arquivo `.node-version`, Go 1.26.5,
   credenciais Git não persistidas pelo checkout, cancelamento de CI obsoleta,
   limites de tempo dos testes, `go mod verify` e checagem documental. Removida
   uma segunda etapa de build que repetia o build já feito por `npm test`.
   Actions continuam pinadas por SHA. YAML das três configurações conferido.
6. `check_docs.py`: checagem sem rede de UTF-8, título e destinos de links dos
   documentos canônicos selecionados; com testes de casos válidos, arquivo
   ausente, caminho externo e encoding inválido. Resultado não é aceite legal.
7. Roteiro único de comandos, ensaios existentes, host e integração futura em
   `docs/VALIDACAO_LOCAL_MVP.md`. Índices, arquitetura, roadmap e modelos
   operacionais jurídicos atualizados sem criar outro contrato concorrente.
   Corrigidas referências erradas no guia do titular e a sugestão de segunda
   revisão obrigatória: permanece uma etapa humana técnica final.

## Verificações observadas

| Verificação | Resultado |
| --- | --- |
| `go test ./... -count=1` na entrada do lote | Passou: 99,776 s no pacote app. É a referência anterior à nova proteção. |
| Novos testes de divulgação de credenciais | Passaram: 7,075 s; dados exclusivamente fictícios. |
| `go mod verify` | Todos os módulos conferidos. Sem mudança em go.mod/go.sum. |
| `go vet ./...` | Passou. |
| `go test -race ./... -count=1 -timeout=10m` após proteção | Passou: 142,003 s no pacote app. Inclui os novos testes de divulgação. |
| Novo cenário de pagamento sem autorização | Passou separadamente com race após acrescentado à suíte: 2,669 s. |
| Landing: lint, TypeScript, build exportado e testes | Passaram; 15 testes, sem falha nem skip. |
| `npm audit --omit=dev --audit-level=high --json` | Zero vulnerabilidade reportada nas dependências de produção na consulta. |
| Cadeia de ferramentas em modo estrito | 20 entradas approved e oito restricted; sem baixar/executar containers. |
| Teste e construção do corpus local | Passaram, offline; não alteram a configuração do MCP. |
| YAML da CI e configuração de deploy | Carregados sem erro pelo parser existente no lock da landing. Não executados no GitHub. |
| Ignorar credenciais de desenvolvimento | `.dev.vars` e variante production ignoradas; exemplos continuam não ignorados. |
| Documentação atual | 37 documentos conferidos, nenhum erro de encoding/título/caminho de link; dois testes do verificador passaram. |
| Formatação | `gofmt -l` vazio nos quatro Go envolvidos; `git diff --check` dos arquivos rastreados alterados sem erro, com avisos LF/CRLF. |

Os comandos de reprodução estão no [roteiro](../../docs/VALIDACAO_LOCAL_MVP.md).
O [manifesto desta pasta](MANIFESTO.json) identifica o conteúdo local revisado por SHA-256; não
é assinatura, atestado de segurança, release publicada nem autorização.

## Limites da entrega

- A busca por padrões não detecta todo segredo/dado pessoal; tokens
  codificados, fragmentados e desconhecidos podem escapar. Também pode
  bloquear exemplos e JWTs públicos. Sanear e autorizar contexto continua
  necessário; o detector não altera o achado nem declara uma vulnerabilidade.
- O ensaio sintético reaproveita testes independentes de onboarding, geração,
  correção, revisão final, pacote e backup. Não simula cliente, assinatura ou
  cobrança como se estivessem homologados.
- A checagem documental não revisa cláusulas, não preenche dados jurídicos e
  não verifica assinatura/poderes. Termos públicos permanecem separados de
  contrato/SOW e autorização. O ponto jurídico continua parcialmente preparado.
- A instalação limpa `npm ci` permanece prescrita na CI; esta sessão utilizou
  as dependências locais existentes, com build/testes e auditoria. Não é
  evidência de reinstalação limpa nem de uma nova execução remota de CI.
- Não foi feita auditoria integral de todo o código nem novo ensaio Linux
  root/Docker, scan em terceiro, OAuth real, emissão fiscal ou restore off-host.
- Não foram contratados fornecedores, criadas contas, escolhidos modelo de
  produção, município, entidade, preço ou capacidade de atendimento.
- Não houve commit, push, deploy, assinatura, pagamento ou mensagem a cliente.
  Alterações e exclusões preexistentes da worktree foram preservadas. Nenhum
  arquivo foi excluído; a limpeza limitou-se à etapa duplicada de build.

## Pendências externas mantidas

Host Linux e acesso para validação real; fornecedor/gateway de IA e condições
de tratamento; dados e decisões da prestadora; revisão jurídica/contábil;
assinatura/cobrança em sandbox e produção; perfil representativo do cliente;
sincronização/release remota; revisão humana final de cada entrega.

As regras jurídicas de assinatura foram reconferidas nas fontes já citadas
pelo modelo: [MP 2.200-2, art. 10](https://www.planalto.gov.br/ccivil_03/mpv/antigas_2001/2200-2.htm)
e [STJ sobre assinatura eletrônica](https://www.stj.jus.br/sites/portalp/Paginas/Comunicacao/Noticias/2024/03122024-Falta-de-credenciamento-da-entidade-certificadora-na-ICP-Brasil--por-si-so--nao-invalida-assinatura-eletronica-.aspx).
Autoria, integridade, representação e licitude não se presumem por um checkbox;
nenhuma plataforma torna a validade de toda contratação indiscutível.
