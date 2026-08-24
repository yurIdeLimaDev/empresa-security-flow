# Runbook — ferramentas manuais

Ferramentas manuais não são chamadas pelo orquestrador. Registre versão, origem,
hash, operador, horário e motivo. Use somente artefato pinado e revisado; não use
`latest` nem instalação ad hoc por npm, pip ou Go.

Para o adaptador manual, coloque exatamente um patch no caminho governado. Não
inclua binário, symlink, dependência ou arquivo fora do ticket. Depois da
execução, confira o resultado hasheado em `agents/`; o orquestrador ainda deve
executar quality, retest e security-full. Limpe o patch externo somente após o
backup criptografado e a entrega serem verificados.

Saídas manuais entram no normalizador como candidatas; sem evidência suficiente
não viram achado confirmado.
