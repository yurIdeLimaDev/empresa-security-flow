# Resultado — ensaio integrado Linux/root

Data: 22 de agosto de 2026.

Status: aprovado em infraestrutura descartável e controlada.

- Host de execução: Docker-in-Docker Linux root, Docker Engine 29.6.2,
  iniciado a partir de `docker:29.6-dind` por digest;
- dois serviços HTTP BusyBox foram criados em redes privadas separadas, sem
  credenciais nem dados de terceiros;
- o canary negativo respondeu quando chamado pelo host e foi bloqueado quando
  chamado pela rede do engajamento;
- o canary positivo foi permitido;
- uma requisição HTTP atravessou o proxy de auditoria e gerou `RequestLog`
  com `tool_run_id=run_isolation_smoke`, status 200 e
  `scope_decision=allowed_pinned`;
- teardown removeu os containers, as redes e as chains do ensaio.

A evidência final preservada é `case-success/`; tentativas exploratórias foram
descartadas.

Correções descobertas pelo ensaio:

1. `network=host` não aceita o sysctl de IPv6 no Docker; o canary de host não
   o usa;
2. as chains `DOCKER-USER` e `INPUT` agora aceitam somente tráfego de retorno
   `ESTABLISHED,RELATED`, antes do `DROP`; sem isso os retornos do target e do
   proxy eram descartados.
