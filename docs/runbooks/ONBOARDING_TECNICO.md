# Runbook — onboarding técnico

1. Colete apenas as entradas explícitas do schema
   `pipeline/config/schemas/onboarding-input.schema.json`.
2. Execute `pipeline onboarding-generate --input <json> --output-dir <novo>`.
3. Confira `status.json`. Sem autorização, isolamento ou perfil revisado,
   `executable` deve permanecer `false`.
4. Valide domínio/base URL, IP:porta, limites, DNS, APIs externas, canaries,
   recursos e identidades fictícias. Não invente SOW, pagamento ou módulo.
5. Gere a configuração concreta pelo perfil aprovado; nunca substitua os
   hashes sentinela do rascunho manualmente.
6. Execute supply-chain, runtime-check e deploy-preflight antes do caso.

HIBP, OAST público, dados pessoais e alvos de terceiros estão fora do primeiro
caso. A ausência de uma entrada encerra a preparação como bloqueada.

## Contratação e geração

Contrato, SOW e autorização devem ser conferidos antes do serviço contratado;
pagamento e login sozinhos não autorizam execução. O checklist jurídico não
é verificação criptográfica de assinatura. O runner recebe entradas explícitas.

No modo de geração, ainda faltam gateway, provedor/modelo/effort, credencial,
autorização de transferência e perfil real de verificadores. A allowlist de
contexto é definida no onboarding, não pelo agente. Configuração incompleta
bloqueia; não substituir pelo simulador do kit. Veja os
[diagramas](../DIAGRAMAS_MERMAID.md) e o
[kit sintético](../../correcao/avaliacao/README.md).
