# Validação da geração de patches sem provedor definido

Data: 11/09/2026. Ambiente: Windows/amd64, Go 1.26.5, Git local.
Escopo: ponto 5 tecnológico. Ponto 3 jurídico/negócios permanece pausado.

Registro histórico daquela entrega. A checagem de credenciais e a revalidação
posteriores estão no [fechamento local](../2026-09-11-fechamento-local/RESULTADO.md).
O manifesto desta pasta identifica a versão anterior, não os arquivos depois
desse reforço; não deve ser sobrescrito para aparentar uma revalidação antiga.

## Resultado

**A camada independente de fornecedor foi implementada e passou na suíte
local. A geração por IA real não está ativada nem validada.**

O ensaio usa um gerador programado em teste e verificadores sintéticos
independentes, em subprocessos pinados pelo hash do executável de teste.
Não utiliza modelo de IA, conta/API real, código de cliente ou alvo de terceiro.
Há um teste HTTPS em servidor TLS local; os demais pedidos usam transporte
simulado em memória. Nenhum simulador está disponível como modo de produção.

## Implementação entregue

- `pipeline/internal/app/patch_generation.go`: configuração, seleção de fontes,
  contrato, transporte HTTPS, validação, aplicação e recibos de geração.
- `pipeline/internal/app/remediation_automation.go`: sequência de tentativas,
  dependências, gates, avaliação, limites, locks e parada na revisão final.
- `pipeline/internal/app/remediation.go`: conexão da geração ao agente,
  inspeção do diff antes do commit/gates, feedback e estados de interrupção.
- `pipeline/cmd/pipeline/main.go`: comando `remediation-run`.
- Dois schemas de protocolo e extensão opcional do schema de remediação.
- Exemplo `agent-generation.disabled.example.json`, sem fornecedor/modelo.
- Testes em `pipeline/internal/app/patch_generation_test.go`.
- Documentação de correção, arquitetura, diagramas, README e roadmap atualizada.

Os hashes dos fontes/contratos e lockfiles estão em [MANIFESTO.json](MANIFESTO.json).
O manifesto identifica arquivos, não é assinatura digital nem atestado de segurança.
`go.mod` e `go.sum` não foram alterados; não foi adicionada dependência.

## Verificações executadas

| Verificação | Resultado observado |
|---|---|
| `go test ./... -count=1` | Passou na suíte completa (execução registrada: 94,848 s no pacote app). |
| `go test -race ./... -count=1` | Reexecutada após o último ajuste; passou em 134,575 s no pacote app. Esta é a validação completa sobre a versão final do manifesto. |
| `go vet ./...` | Passou, sem diagnóstico. |
| Teste específico de resultado desconhecido do gateway | Passou; somente uma chamada, uma tentativa contabilizada e bloqueio sem retry automático. |
| `gofmt -l` nos cinco arquivos Go alterados/adicionados | Sem arquivo pendente de formatação. |
| SHA-256 dos arquivos do manifesto | Conferidos contra os arquivos locais. |
| Links locais da documentação afetada | Resolvidos. |
| `git diff --check` | Sem erro; apenas aviso de normalização LF/CRLF no README. Essa verificação cobre arquivos rastreados; a formatação dos Go novos foi conferida separadamente. |

Os tempos não são benchmark do produto. Não houve teste de capacidade, cobrança
real, deploy, CI remota, execução Docker/Linux root ou assinatura de entrega.

## Cenários cobertos

1. Proposta gerada durante o teste a partir do conteúdo recebido, sem arquivo
   `.patch` preparado externamente; candidato passa pelos três verificadores e
   pela validação global, terminando em `awaiting_final_review`.
2. Checkout original preservado; BEST muda somente após avaliação; comando
   não produz aprovação nem autorização de entrega.
3. Resposta de modelo/effort/fornecedor diferente, replay, hash anterior
   divergente, path traversal, alteração de teste, conteúdo binário, duplicidade
   de arquivo, ausência de proposta e excesso de linhas são rejeitados.
4. Alteração que muda comportamento não relacionado à segurança falha no
   verificador independente; não promove o candidato.
5. Rejeições se repetem somente dentro dos limites, preservam BEST e encerram
   o fluxo quando as tentativas acabam.
6. Uma primeira resposta inválida seguida de proposta válida termina na revisão
   final, sem intervenção humana por ticket e com feedback na nova tentativa.
7. Gate global falho mantém a melhor versão, mas não libera entrega.
8. Geração desativada/incompleta é bloqueada antes de worktree/rede. HTTP sem
   TLS, credencial na URL/query, envio não autorizado, contexto protegido e
   credencial compartilhada com gates são recusados.
9. Redirecionamento não é seguido; mensagem de erro não reproduz corpo/token
   retornado pelo servidor; tamanho de resposta e timeout são limitados.
10. JSON ambíguo, campos desconhecidos e conteúdo extra são recusados.
11. Lock impede segunda automação e alteração manual concorrente. Estado
    `generating` interrompido não é repetido automaticamente.
12. Transporte com resultado desconhecido bloqueia em
    `blocked_provider_uncertain`, sem segunda consulta automática.
13. Fonte não rastreada, fonte acima do limite e ausência de contexto editável
    são recusadas. O teste de link simbólico é condicionado à permissão do SO
    para criar symlinks; a rejeição também existe no código de leitura.
14. Modificação do worktree durante a geração invalida a proposta.
15. Os contratos JSON enviados/recebidos pelo gerador de teste são conferidos
    contra os schemas versionados. A credencial fictícia não chega aos gates.

## Ajustes identificados durante a validação

- A amostra inicial comparava um arquivo protegido com LF, mas o Git do Windows
  convertia seu checkout para CRLF. O teste foi corrigido com configuração local
  `core.autocrlf=false` no repositório temporário; o verificador não foi afrouxado.
- A revisão de recuperação identificou que resultados desconhecidos de rede
  não devem disparar nova consulta potencialmente cobrada. Foi adicionado o
  bloqueio específico e um teste que exige apenas uma chamada.

## Não comprovado / pendente

- Qualidade de patches de qualquer IA real, taxa de acerto ou aplicabilidade
  comercial multistack. Os verificadores deste ensaio são sintéticos.
- Tradução do protocolo para o fornecedor futuro, modelo/effort efetivos,
  autenticação real, orçamento monetário/tokens, retenção e tratamento de dados.
- Funcionamento integrado em gateway/host Linux de produção, seus firewalls,
  disponibilidade, isolamento e recuperação operacional.
- Proteção semântica absoluta de toda alteração: diff e testes não substituem
  a revisão final nem a validação de um perfil representativo.

O perfil manual existente não foi removido, dados/casos de clientes não foram
alterados, e nenhuma configuração de produção, landing, conta, pagamento ou
documento jurídico foi ativado/modificado nesta etapa. Não houve commit, push
ou deploy; a entrega é local.

## Reprodução e próximos passos

Rodar os comandos Go acima dentro de `pipeline/`. Os testes criam repositórios
temporários isolados e os removem ao terminar. Para conectar IA de verdade,
seguir [GERACAO_PATCHES_SEM_PROVEDOR.md](../../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).
Não substituir o simulador por evidência de prontidão de um cliente real.
