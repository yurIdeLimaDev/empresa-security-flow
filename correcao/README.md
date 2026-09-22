# Correção automatizada de segurança

Atualização local de 13/09/2026: [kit sintético de avaliação](avaliacao/README.md)
e [estado do fluxo](../docs/ESTADO_FLUXO.md). A entrega deriva o diff cumulativo
baseline → BEST na finalização e o verifica por hash ao empacotar; patches
fornecidos externamente não entram diretamente na entrega.

Esta pasta concentra a documentação e os contratos do fluxo de correção. O
motor reutiliza o modelo canônico e a CLI em `pipeline/`; a landing React
continua isolada em `landing-page/`; evidências de teste ficam somente em
`validacao/`.

## Conteúdo

- `docs/FLUXO_DE_CORRECAO_SEGURA.md`: arquitetura, gates e limites;
- `docs/GERACAO_PATCHES_SEM_PROVEDOR.md`: integração de geração, protocolo,
  automação até revisão final e limites sem fornecedor escolhido;
- `config/agent-generation.disabled.example.json`: objeto `agent` desativado,
  para incorporação futura a uma política concreta;
- `config/remediation.example.json`: política estrita de exemplo;
- `config/final-approval.example.json`: única decisão humana do fluxo.
- `casos/2026-08-22-cer-facil/`: primeira correção real em cópia isolada,
  com fonte Git e patch; o reteste fica em `validacao/`.

Execuções reais não devem gravar artefatos dentro do repositório do cliente.
`output_root` deve apontar para uma raiz externa, exclusiva e protegida. Os
worktrees, tickets, prompts, logs, bundles e decisões são organizados por
`case_id` nessa raiz.

Existem dois caminhos explícitos: o adaptador externo/manual já existente e
`builtin:patch-proposal`, que recebe propostas JSON de um gateway futuro e as
aplica sem executar comandos do modelo. Ambos passam pela governança do motor.
O novo `remediation-run` automatiza somente o segundo caminho. A promoção não
depende da palavra do agente: depende do diff, gates e bundle comparável.
Não há fornecedor/modelo escolhido; o exemplo novo falha fechado e os testes
usam um gerador simulado, sem consumo de IA.

`remediation-preflight` precisa aprovar o agente e todos os gates antes do
plano. Os nomes e hashes zerados do exemplo são intencionalmente recusados;
cada stack deve fornecer comandos reais e pinados no onboarding.

Leia o [fluxo completo](docs/FLUXO_DE_CORRECAO_SEGURA.md) antes de configurar
um caso.

## Primeiro perfil executável

`adapters/python-3.13-stdlib/` continua o único perfil de stack ensaiado. Ele
usa patch manual por caminho exato e três gates pinados. Rascunhos genéricos do
onboarding continuam bloqueados até a seleção de um perfil revisado. A revisão
humana permanece apenas no final; candidato pior que BEST é rejeitado.

O laboratório está em repositório privado separado. Casos de cliente, patches
operacionais, worktrees e chaves não pertencem a este Git.
