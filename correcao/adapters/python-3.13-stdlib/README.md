# Perfil `python-3.13-stdlib`

Versão 1.0.0. Este é o único perfil de correção executável do primeiro caso.
Ele cobre uma aplicação Python 3.13 sem dependências de terceiros e foi ensaiado
somente no laboratório privado com dados fictícios.

O agente é manual e governado: recebe um patch previamente preparado em
`<patch-root>/<case-id>/<ticket-id>.patch`, exige worktree limpa, valida o
escopo do ticket e executa `git apply --check --whitespace=error-all` antes de
aplicar. Não há rede, instalação de pacote, checkout, reset ou commit pelo
adaptador. O orquestrador faz o commit somente depois de conferir o diff.

Gates obrigatórios:

- `quality-project`: testes funcionais protegidos;
- `retest-property`: propriedade de segurança ligada ao ticket;
- `security-full`: todos os controles protegidos e novo bundle canônico.

Todos usam o mesmo runner com SHA-256 verificado. O runtime é
`python:3.13-slim-bookworm@sha256:00faa2debb87529f9f0764e9491d8ba400a3678976616c3bd7cb193745ac20d1`,
já aprovado em `pipeline/tools.lock.json`, com rede desabilitada, filesystem
read-only, capacidades removidas e limites de CPU/memória/PIDs.

Para gerar a configuração concreta, use `pipeline reference-profile`. O perfil
genérico produzido por onboarding permanece deliberadamente bloqueado; não o
torne executável trocando apenas o hash.
