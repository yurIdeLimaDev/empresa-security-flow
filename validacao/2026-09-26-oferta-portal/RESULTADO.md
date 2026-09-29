# Resultado: oferta inicial, landing e portal

Data: 26/09/2026. Validação local; sem deploy, migração remota, cobrança, consulta de cliente ou teste de alvos externos. Alterações anteriores do worktree foram preservadas. Nenhum commit ou push foi feito neste lote.

## Entregas

- [Oferta de uma página](../../negocio/oferta/OFERTA_MVP_UMA_PAGINA.md), com ICP, problema, entrega, inclusões/exclusões, unidade de pedido, prazo e condições comerciais. PDF de uma página em `negocio/oferta/output/pdf/OFERTA_MVP_UMA_PAGINA.pdf`, gerado da fonte Markdown, extraído e conferido visualmente sem cortes/sobreposições.
- [Pesquisa e memória de preços](../../negocio/oferta/PESQUISA_PRECOS_2026-09-26.md): dez fornecedores/alternativas separados por implementação de correções versus orientação. Base principal Vibe Code Rescue; 15% abaixo de US$ 1.500 e US$ 499/mês, convertidos pela PTAX venda 5,1991 de 25/09/2026. Anual deriva de 12 mensalidades, não de tarifa anual publicada. Fonte da cotação: `ptax.json`.
- Catálogo compartilhado `landing-page/config/commercial-offer.json`: avulsa R$ 6.628,85; mensal R$ 2.205,20; anual R$ 26.462,38. Avulsa até 3 pedidos; assinaturas 2/mês. Hipóteses de oferta, sem checkout ativo.
- Landing com headline profissional e frases rotativas decorativas, pausa explícita e respeito a movimento reduzido/foco. Formulário mantém rótulo estável e confirmação humana.
- Portal de conta com etapa, previsão, atraso, pendências, correções e recibo de entrega. API somente leitura por sessão, migração aditiva D1, projeção mínima por proprietário e preparação offline de atualizações. Nenhum pedido ou saldo fictício no código publicado.
- [Backlog público separado](../../docs/BACKLOG_SEGURANCA_APLICACAO_PUBLICA.md), com evidência de código, prioridade e critérios de aceite para SSRF/rebinding, redirects, IPs privados, quotas, limites e timeouts.
- Documentação, diagrama de oferta/portal, termos públicos e privacidade alinhados à implementação e às limitações reais. Nenhuma aprovação jurídica é presumida.

## Verificações executadas

| Verificação | Resultado |
| --- | --- |
| `npm run build:cloudflare` | Aprovado; sete rotas estáticas geradas |
| `npm run lint` | Aprovado; diretório temporário `.wrangler` excluído do lint |
| `npm exec -- tsc --noEmit` | Aprovado |
| `node --test tests/*.test.mjs` | 26 testes aprovados, zero falhas |
| `npm audit --omit=dev` | Zero vulnerabilidades reportadas na consulta; não equivale a auditoria integral de segurança |
| `pipeline/scripts/check_docs.py` | 63 documentos aprovados em UTF-8, títulos e caminhos de links locais |
| PDF | Uma página; preços e seções presentes; legibilidade e layout conferidos |
| Navegador local | Home, controle Pausar/Retomar, painel/Atualizar e oferta inspecionados; painel e planos legíveis em largura móvel de 390 px |

Os testes cobrem isolamento A/B em SQLite, falta de sessão, origem externa, query/método indevidos, falha de banco versus conta vazia, snapshot adulterado, limite de 50 casos, revisão monotônica, proprietário imutável, escape SQL, cascade e requisitos de revisão/recibo. A autenticação do handler é simulada nos testes unitários; não se afirma homologação OAuth real.

A prévia `preview-server.mjs` usa exclusivamente dados fictícios em loopback, com faixa de aviso e APIs externas/POST bloqueados. Não integra o pacote de deploy nem deve ser exposta na Internet. A captura de página completa com viewport forçado apresentou artefatos de composição do navegador; a conferência visual válida usou as capturas normais e a árvore acessível. Não usar essa captura como evidência visual de produção.

## Pendências explícitas

1. Homologar migração 0002 no D1 de staging com backup, login real A/B, expiração/logout e metadados saneados de um caso real. Só depois aplicar em produção e publicar.
2. Operador publica os marcos pelo procedimento de [PORTAL_CLIENTE](../../landing-page/PORTAL_CLIENTE.md). Integração automática motor→portal e download privado não foram implementados; entrega continua pelo canal seguro combinado.
3. Cobrança, conciliação e ledger de franquia não implementados. Não há saldo fictício nem liberação pelo pagamento isolado.
4. Fechar P0 de egress/DNS rebinding, rate limiting agregado e deadline global antes de ampliar aquisição. Foram documentados, não corrigidos neste lote.
5. Validar custo/capacidade em ensaios cronometrados e fechar contrato de assinatura/cancelamento/restituição antes de cobrar, especialmente anual.

## Reproduzir

Na pasta `landing-page`: executar build, lint, TypeScript e testes acima. Na raiz: executar `pipeline/scripts/check_docs.py`. Para PDF, usar Python com reportlab/pypdf e `validacao/2026-09-26-oferta-portal/render-offer.py`; conferir novamente a imagem renderizada após qualquer edição da oferta. Artefatos temporários ficam em `tmp/` e dependências de consulta em `vendor-types/`, ambos ignorados pelo Git.
