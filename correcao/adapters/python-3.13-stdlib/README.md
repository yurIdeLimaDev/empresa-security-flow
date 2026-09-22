# Perfil `python-3.13-stdlib`

Versão 1.0.0. Este é o único perfil de correção executável do primeiro caso.
Ele cobre uma aplicação Python 3.13 sem dependências de terceiros e foi ensaiado
somente no laboratório privado com dados fictícios.

Atualização de 11/09/2026: existe uma camada independente de fornecedor para
geração automática, documentada em
[GERACAO_PATCHES_SEM_PROVEDOR.md](../../docs/GERACAO_PATCHES_SEM_PROVEDOR.md).
Este perfil não foi alterado/ativado para IA: seu adaptador padrão continua
manual. Validar a camada com simulador não certifica este perfil com IA real.

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

## Entrega após o perfil manual

Atualização documental de 13/09/2026: o `patch-root` acima permanece entrada
do adaptador, não do empacotador. Após gates globais e revisão humana, a
finalização produz `approved-security.patch` cumulativo de baseline até BEST.
O empacotador recebe esse arquivo vinculado à autorização por SHA-256 e recusa
`--patch-root`. Veja o [runbook de entrega](../../../docs/runbooks/ENTREGA_E_REVISAO.md).
