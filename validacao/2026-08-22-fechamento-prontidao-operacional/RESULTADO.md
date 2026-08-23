# Fechamento da prontidão operacional

Data: 22 de agosto de 2026.

Status: **motor aprovado; deploy condicionado ao caso e ao host escolhido**.

## Pendências tratadas

| Pendência | Resultado |
|---|---|
| 28 revisões `pending` | Fechada: 20 wrappers automáticos `approved`; oito entradas `restricted`; zero `pending`. |
| 14 ferramentas sem container | Fechada sem imagens artificiais: as 20 automáticas possuem runtime por digest; as oito não automáticas são recusadas. |
| Liveness dos scanners | 20/20 runtimes automáticos passaram com pull por digest e perfil endurecido; 8/8 restritos não foram executados. |
| Build limpo/reprodutível | Runner recompilado duas vezes com resultado idêntico; Subfinder, Hadrian, jsluice e sqlmap recompilados de commits exatos e hashes conferidos. |
| Docker/firewall/proxy/canaries | Ensaio integrado aprovado em Linux root descartável; negativo bloqueado, positivo permitido e `RequestLog` emitido. |
| Adapters de correção | Novo `remediation-preflight`; hash zero, executável ausente, hash divergente ou working directory inválido bloqueiam antes do plano e são rechecados na execução. |
| Escala/falhas | Suite cobre 5.000 bundles deduplicados, adulteração de artefato, contrato de saída, limites de log, timeout/cleanup, retenção de `BEST`, tentativa máxima e race detector. |

## Evidências

- `supply-chain-report.json`: `approved=true`, 28 decisões consistentes;
- `runtime-readiness.json`: `approved=true`, Docker 29.6.1, pull solicitado,
  20 `passed` e oito `restricted`;
- `../2026-08-22-isolamento-politicas/linux-root-integration/case-success/`:
  regras de rede, canaries, proxy e checksums;
- `../../docs/REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md`: escopo, decisão
  e limite da revisão;
- `../../pipeline/scripts/rebuild-runtime-artifacts.sh`: reconstrução dos
  quatro artefatos derivados de fonte;
- `../../pipeline/scripts/build-runner-linux.sh`: build duplo determinístico;
- `../../pipeline/scripts/deploy-preflight-linux.sh`: gate obrigatório do host
  real.

## Verificações finais desta versão

- `go test ./... -count=1`: aprovado;
- `go test -race ./... -count=1`: aprovado;
- validação estrita do lock/SBOM: aprovada;
- runtime integrado por digest: aprovado;
- sintaxe dos três scripts Bash: aprovada;
- schema do relatório de prontidão dos adaptadores: aprovado em teste.

O build duplo final do runner linux/amd64 produziu, nas duas execuções,
`00e91c8afdc50ba23c9ff591e4bdfb2897d70db38f00f857531bcb35c8696ea0`.

## Limites e decisão honesta

A revisão das ferramentas é de adoção e da superfície usada, não auditoria
linha a linha. Doze commits upstream não são assinados; o commit exato, a
receita, o digest e o sandbox compensam parcialmente, mas não criam garantia
absoluta. O liveness comprova que o runtime pinado inicia e responde; a
eficácia de detecção continua dependente do alvo, dos insumos e do reteste.

Não foi inventado um adaptador universal de IA/build/teste: isso seria incorreto
entre stacks e poderia alterar o projeto do cliente. Esses adaptadores são
entradas do onboarding, e agora falham fechados antes do plano.

O ensaio Linux root valida o desenho, não um servidor futuro ainda não
selecionado. Por isso o código está pronto para receber um caso, mas a entrega
real só pode ser autorizada depois que políticas/adaptadores reais e o host
escolhido passarem seus preflights. Todo caso precisa usar a cadeia atual antes
da revisão humana final.
