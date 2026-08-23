# Resultado — isolamento Linux/root

Data: 22 de agosto de 2026.

Status: aprovado em infraestrutura descartável e controlada.

- bridge exclusiva e regras `DOCKER-USER`/`INPUT` aplicadas;
- canary negativo acessível pelo host e bloqueado pela rede do engagement;
- canary positivo permitido;
- requisição pelo proxy registrada com `scope_decision=allowed_pinned`;
- teardown removeu containers, redes e chains;
- somente `linux-root-integration/case-success/` foi preservado.

A cadeia de ferramentas e o runtime foram revalidados posteriormente em
`../2026-08-22-fechamento-prontidao-operacional/`. O ensaio comprova o desenho,
mas o host selecionado para deploy precisa executar seu próprio preflight.
