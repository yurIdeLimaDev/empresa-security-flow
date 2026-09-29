# Correção de Turnstile e título da landing

Data: 26/09/2026. Escopo: validação humana e área inicial indicada pelo titular.

## Causa e correções

- A função `verifyTurnstile` usava `redirect: error`. O runtime workerd rejeitava
  essa opção antes de consultar Siteverify. O widget podia mostrar sucesso,
  mas o servidor falhava. A falha foi reproduzida no navegador e em teste workerd.
- Alterado para `redirect: manual`; respostas 3xx continuam rejeitadas.
  Permanecem checagens de sucesso, action, hostname, timeout e tamanho de resposta.
- Token consumido por tentativa, desafio renovado ao finalizar e envio concorrente
  bloqueado. O botão apresenta spinner e `Verificando...` durante o processamento.
- Animação movida para o H1. O formulário começa diretamente pelo endereço.
  Mantidos pausa, movimento reduzido e reserva de altura para evitar saltos.
- A consulta real ao próprio domínio revelou HTTP 523 por fetch interno da zona.
  Ativada `global_fetch_strictly_public`: a consulta segue a rota pública e passou
  a observar HTTP 200. Não se trata de proteção completa contra DNS rebinding.

## Verificações

- 34 testes aprovados (`node --test tests/*.test.mjs`), build, lint e TypeScript.
- Novo teste executa a função real no workerd: sucesso, redirect, action inválida,
  hostname inválido e token reutilizado. Falhou antes da correção e passou depois.
- Novo teste do formulário cobre falta de token, spinner, envio duplicado,
  erro, renovação e segunda consulta com token novo.
- Testes reais autorizados apenas em `https://vexkeep.com`, sem contratação:
  duas consultas concluídas às 19:38 e 19:39 de Brasília, sem recarregar a página;
  após a correção da rota pública, terceira consulta às 19:40, concluída em 8.861 ms,
  HTTP 200, sete de oito cabeçalhos observados. Nenhum erro de confirmação humana.
- Layout sem overflow horizontal: viewport de 1280 px (conteúdo 1265 px) e
  viewport padrão de 423 px (conteúdo 408 px). Título permanece dentro da área.
- Não houve alteração de preços, cobrança, dados de autenticação ou secrets.

## Publicação

- Worker: `vexkeep-landing`.
- Versão final: `2cce29f1-11d5-4c15-87ac-8524518e6361`.
- Horário: `2026-09-26T22:40:26.228224Z`.
- SHA-256 do bundle minificado:
  `52E741CF59FAD31C95068E4CA519428A715CD2EBFCA14373BFD23AD8B680AD18`.
- Configuração: data de compatibilidade `2026-08-28`, `nodejs_compat` e
  `global_fetch_strictly_public`. D1, bindings secretos e cron preservados.
- Artefatos temporários e capturas ficam em `tmp/`, ignorado pelo Git.

## Limites

Não é uma auditoria integral nem certificação de segurança. DNS rebinding/egress
com IP fixado continua pendente no backlog. Não foi homologado OAuth completo,
pagamento ou entrega contratada. CAPTCHA não foi desligado nem simulado em produção.
Regressão de código deve ser corrigida com novo deploy: retornar à versão anterior
a este lote reintroduz o problema do Turnstile.

Referência: [Cloudflare — global fetch strictly public](https://developers.cloudflare.com/workers/configuration/compatibility-flags/#global-fetch-strictly-public).
