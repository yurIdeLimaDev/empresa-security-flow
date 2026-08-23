# Correção automatizada de segurança

Esta pasta concentra a documentação e os contratos do fluxo de correção. O
motor reutiliza o modelo canônico e a CLI em `pipeline/`; a landing React
continua isolada em `landing-page/`; evidências de teste ficam somente em
`validacao/`.

## Conteúdo

- `docs/FLUXO_DE_CORRECAO_SEGURA.md`: arquitetura, gates e limites;
- `config/remediation.example.json`: política estrita de exemplo;
- `config/final-approval.example.json`: única decisão humana do fluxo.

Casos de clientes e benchmarks ficam fora do repositório do motor.

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
