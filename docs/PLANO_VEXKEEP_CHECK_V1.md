# Vexkeep Check v1

Estado: pausado e retirado da oferta pública em 26 de agosto de 2026.

## Decisão vigente

O Vexkeep Check e o Pipeline 1 gratuito não fazem parte do MVP público atual.
A landing executa apenas o Pipeline 0.5 efêmero. Se a pessoa contratar o
serviço, o Pipeline 1 integral é executado no ambiente governado antes do
onboarding e de qualquer Pipeline 2 aplicável.

O código do Check e a migração do D1 de staging foram preservados para uma
eventual retomada, mas não representam produto ativo. Produção não define
`CHECK_PUBLIC_ENABLED`; por isso:

- `/api/check/*` retorna 404;
- `/internal/check-*` retorna 404;
- o scheduler do Check encerra sem criar trabalho;
- a landing não apresenta ativação, prova de domínio ou acompanhamento mensal;
- `/resultado` é somente uma página informativa.

## Condições para uma futura retomada

Uma nova decisão explícita deve preceder qualquer ativação. Depois disso ainda
será necessário validar política dinâmica, runner Linux isolado, secrets,
pagamento autenticado e idempotente, retenção, exclusão e matriz controlada
completa. Nenhuma dessas condições altera a decisão atual.

## Fonte atual

A jornada vigente, cobertura e limites estão em
[`PIPELINE_0_5_WEB.md`](PIPELINE_0_5_WEB.md). A arquitetura geral continua em
[`PIPELINE_DE_GERACAO_DE_LEADS.md`](PIPELINE_DE_GERACAO_DE_LEADS.md).
