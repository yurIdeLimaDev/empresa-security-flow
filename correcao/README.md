# Correção automatizada de segurança

Esta pasta concentra a documentação e os contratos do fluxo de correção. O
motor reutiliza o modelo canônico e a CLI em `pipeline/`; a landing React
continua isolada em `landing-page/`; evidências de teste ficam somente em
`validacao/`.

## Conteúdo

- `docs/FLUXO_DE_CORRECAO_SEGURA.md`: arquitetura, gates e limites;
- `config/remediation.example.json`: política estrita de exemplo;
- `config/final-approval.example.json`: única decisão humana do fluxo.
- `casos/2026-08-22-cer-facil/`: primeira correção real em cópia isolada,
  com fonte Git e patch; o reteste fica em `validacao/`.

Execuções reais não devem gravar artefatos dentro do repositório do cliente.
`output_root` deve apontar para uma raiz externa, exclusiva e protegida. Os
worktrees, tickets, prompts, logs, bundles e decisões são organizados por
`case_id` nessa raiz.

O provedor de IA que edita código é um adaptador externo, mas sua execução é
governada pela CLI: binário pinado por SHA-256, timeout, ambiente reduzido,
worktree do ticket, log e commit automático. A promoção da versão não depende
da palavra do agente: depende do diff, dos gates e do bundle canônico
comparável.

`remediation-preflight` precisa aprovar o agente e todos os gates antes do
plano. Os nomes e hashes zerados do exemplo são intencionalmente recusados;
cada stack deve fornecer comandos reais e pinados no onboarding.

Leia o [fluxo completo](docs/FLUXO_DE_CORRECAO_SEGURA.md) antes de configurar
um caso.

## Primeiro perfil executável

`adapters/python-3.13-stdlib/` é o único perfil aprovado nesta iteração. Ele
usa patch manual por caminho exato e três gates pinados. Rascunhos genéricos do
onboarding continuam bloqueados até a seleção de um perfil revisado. A revisão
humana permanece apenas no final; candidato pior que BEST é rejeitado.

O laboratório está em repositório privado separado. Casos de cliente, patches
operacionais, worktrees e chaves não pertencem a este Git.
