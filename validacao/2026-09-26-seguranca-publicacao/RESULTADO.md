# Registro da publicação de oferta e controles públicos

Data do lote: 26/09/2026. Registro histórico consolidado; não representa nova
execução dos testes descritos abaixo.

- Publicação `aea1470c-efd6-4e25-8788-b1a641e81035` com oferta r2, portal,
  cotas D1, deadline e leitura limitada. Migrações 0002–0004 aplicadas após staging.
- Preços publicados: R$ 1.350,00 avulsa, R$ 449,10 mensal e R$ 5.389,20 anual.
- Naquele lote: 32 testes aprovados, build, lint, tipos e audit de produção.
  Testes de portal incluíram sessão sintética com Better Auth e D1 local workerd.
- O smoke test negativo de CAPTCHA não comprovava o caminho positivo. Uma falha
  específica do runtime foi identificada e corrigida posteriormente no mesmo dia.
  Estado vigente: [correção de Turnstile e título](../2026-09-26-turnstile-hero/RESULTADO.md).
- Cobrança permanece desligada; ledger não é integração financeira.
  DNS rebinding/egress e homologação completa de OAuth continuam pendentes.

Rollback de código não deve apagar tabelas novas. Restauração integral por
Time Travel exige avaliar gravações posteriores; não executá-la automaticamente.
