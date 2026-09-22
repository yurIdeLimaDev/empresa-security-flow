# Estado do fluxo de verificação e correção

Atualizado em 22/09/2026. Este documento acompanha a publicação separada do
motor técnico. O site React e o pacote comercial completo permanecem fora do
repositório do motor. Quatro minutas não preenchidas são usadas somente como
insumos do ensaio sintético de contratação e autorização.

## Implementado

- Pipeline 1 integral obrigatório antes da seleção de módulos do Pipeline 2;
- políticas de escopo, autorização, janela e ferramentas, com bloqueio por padrão;
- normalização de achados, evidências, comparação e relatórios;
- execução governada com isolamento, controle de rede, proxy e canaries;
- correção serial sobre a melhor versão, com tentativas finitas e verificadores
  independentes;
- contrato de geração de patches independente de fornecedor e ensaio sintético;
- revisão humana final, patch cumulativo com hash e entrega vinculada;
- CI, scripts de preflight e coleta de aceite do host.

O [relatório do lote local](../validacao/2026-09-13-lote-completo/RESULTADO.md)
registra os testes realizados e seus limites. A
[geração de patches](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md) ainda usa
respostas programadas na validação; nenhum provedor ou modelo de IA foi ativado.

## Pendente para um cliente real

- Aprovar host Linux definitivo com Docker, firewall, canaries, proxy, backup
  externo e restore efetivo;
- escolher e validar provedor/modelo de IA com casos representativos e regras
  de tratamento de código autorizadas;
- obter contratação, escopo, autorização e pagamento verificados para cada caso;
- executar verificações e correções no perfil concreto do cliente;
- concluir revisão humana de cada entrega.

Os testes sintéticos e o código publicado não representam aprovação do host
nem autorização para atuar em ativos de terceiros.
