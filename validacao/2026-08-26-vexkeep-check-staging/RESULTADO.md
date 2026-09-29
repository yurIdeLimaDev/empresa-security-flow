# Validação do schema e controles locais do Vexkeep Check

**Data:** 26 de agosto de 2026  
**Escopo:** somente staging e artefatos sintéticos  
**Resultado:** aprovado para continuar desenvolvimento; não aprovado para produção

## D1 de staging

- banco: `vexkeep-check-staging`;
- região informada pelo provedor: `ENAM`;
- migração aplicada: `landing-page/migrations/0001_vexkeep_check.sql`;
- tabelas conferidas: casos, cooldown, eventos, jobs, exclusões, relatórios e recibos;
- colunas novas conferidas: `consecutive_failures` e `heartbeat_at`;
- caso e job sintéticos foram inseridos, lidos e removidos;
- nenhum domínio, e-mail, pagamento, credencial ou dado de cliente foi usado.

## Software

- lint: aprovado;
- TypeScript `--noEmit`: aprovado;
- build Cloudflare: aprovado;
- testes Node: 10/10 aprovados;
- audit de dependências de produção: zero vulnerabilidades;
- dry-run do Wrangler com o binding de staging: aprovado;
- `go vet ./...`: aprovado;
- `go test -race -count=1 ./...`: aprovado.

## Controles incorporados

- cooldown atômico de 24 horas por HMAC da origem;
- expiração do registro de cooldown pelo cron;
- heartbeat do runner;
- stale após três minutos;
- uma hora entre novas tentativas e trava após três falhas consecutivas;
- cancelamento durante execução sinalizado no heartbeat;
- razão de falha restrita a código saneado.

## Limites

Não foram testados runner Linux, política dinâmica, Docker, firewall, pagamento,
reembolso, exclusão física no host ou rollback. O Worker de staging não foi
publicado e nenhum secret foi criado. Esses limites impedem ativação comercial.
