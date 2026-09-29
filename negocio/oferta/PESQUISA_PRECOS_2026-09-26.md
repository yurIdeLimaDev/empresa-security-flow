# Pesquisa de mercado e memória de preço

Consulta em 26/09/2026, revisão r2. Benchmark reconfirmado na página oficial nesta revisão. Fontes primárias públicas; sem compra, cadastro comercial ou teste dos serviços. Preço publicado e promessa comercial não comprovam eficácia do fornecedor. A decisão mais recente do titular substitui tanto 5% quanto 15% com câmbio: mesmo número USD em BRL e redução de 10%. Não altera contratos existentes.

## Critério e conclusão

Concorrente direto, conforme a decisão do titular: identifica falhas e implementa correções ou gera patches aplicáveis. Relatório, instrução, prompt ou reteste do que o próprio cliente corrigiu não bastam. Separei serviços executados por especialistas de ferramentas self-service; ambos podem competir, mas não oferecem a mesma responsabilidade de entrega.

**Referência principal escolhida: Vibe Code Rescue.** Atende aplicações pequenas construídas com ajuda de IA, oferece diagnóstico, implementação e orçamento recorrente de correções. Está mais próximo da compra de um resultado pelo fundador do que uma licença por desenvolvedor. A oferta é mais ampla que a Vexkeep (inclui trabalho funcional). A franquia exata do Keep-Alive não é pública; nossos 2 pedidos não são uma franquia copiada nem uma comparação por unidade equivalente.

## Serviços e alternativas avaliados

| Fornecedor / fonte oficial | Preço encontrado | Implementa? | Uso na decisão |
| --- | --- | --- | --- |
| [Vibe Code Rescue](https://vibecoderescue.dev/pricing) | Audit US$ 299; Sprint US$ 1.500; Keep-Alive US$ 499/mês | Sprint corrige problemas do audit; mensal inclui orçamento de correções | Benchmark principal. Os US$ 299 são creditados no Sprint: não somar 299 + 1.500. Anual não publicado. |
| [Done Wright Software](https://donewrightsoftware.com/ai-project-rescue) | Security hardening US$ 2.400–8.000 | Auditoria + correção de achados altos/críticos | Direto para serviço pontual; escopo maior, sem mensal/anual comparável publicado. |
| [Teamseven](https://www.teamseven.ltd/code-audit) | US$ 8.000–12.000 | Inclui correções críticas de segurança/estabilidade | Direto, mas mais amplo e distante do MVP enxuto. |
| [Glue Studio](https://gluestudio.co.uk/services/website-security-audit/) | £495, pontual | Auditoria WordPress + correções padrão; trabalho maior separado | Direto no nicho WordPress, não benchmark para SaaS customizado. |
| [Aikido](https://www.aikido.dev/pricing) | Developer gratuito, 10 AutoFixes/mês; Basic mostra 300/mês; tabela dinâmica por moeda/ciclo | AutoFix abre PRs para revisão/merge | Direto tecnológico. Não usar 300 sem confirmar seletor de moeda/ciclo. O gratuito impede alegar “mais barato que todos”. Não entrega a mesma operação humana gerenciada. |
| [Fixier](https://fixier.co/) | Sob orçamento | Audit + Fix implementa no repositório e reavalia | Direto qualitativo; sem número público utilizável. |
| [CyberBit](https://cyberbitsolutions.com/pricing) | Cleanup US$ 750; sprint desde US$ 1.500 | Implementa correções focadas | Comparável para configuração/infra. Snapshot US$ 199 não inclui implementação; mensal Watch não equivale a franquia de correção. |
| [VibeArmor](https://vibearmor.ai/) | Report US$ 499; Continuous US$ 999/mês | Página oferece instruções/prompts e guia; não confirma execução de patches nesses planos | Excluído do cálculo: encontrar e orientar não é realizar a correção. |
| [SaaS Security](https://www.saassecurity.io/) | Essential US$ 249 por scan; Professional US$ 2.499 | Cliente corrige com orientação; re-scan | Excluído do cálculo. |
| [Valletta](https://valletta-software.com/vibe-coding-cleanup) | Auditoria desde US$ 199; cleanup orçado depois | Correção existe, mas não pelo preço do audit | Direto qualitativo; US$ 199 não é preço de encontrar + corrigir. |

Não localizei, nesta amostra, uma oferta brasileira de SaaS customizado com execução de correções, franquia comparável e preços mensal/anual públicos suficientemente claros. Isso não prova ausência de concorrentes. Não utilizei redação escolar, antivírus, limpeza exclusiva de malware ou SEO como comparação.

## Cálculo reproduzível

Não se aplica câmbio. A PTAX registrada na validação anterior é apenas evidência histórica da hipótese substituída, não entrada do catálogo atual.

Fórmula: preço Vexkeep em centavos = arredondar(número USD do benchmark × 0,90 × 100). Exemplo solicitado: US$ 100 → R$ 100 → R$ 90.

| Modalidade | Número publicado em USD | Vexkeep BRL após 10% |
| --- | --- | --- |
| Avulsa | 1.500 | 1.350,00 |
| Mensal | 499/mês | 449,10/mês |
| Anual | 499 × 12 = 5.988 | 5.389,20 |

**Limite importante:** anual é normalização de 12 mensalidades, não tarifa anual publicada pelo concorrente. Não existe desconto anual adicional. Os números são política comercial determinada pelo titular, não comparação econômica entre moedas. Impostos/encargos próprios e capacidade precisam ser fechados antes da primeira venda.

Não publicar “10% mais barato que o mercado”: há moedas, escopos e ofertas gratuitas diferentes. Não vender “validação avulsa” como se incluísse correção sem descrevê-la; o SKU adotado é **avaliação focada + correção avulsa**. A prévia por URL continua gratuita.

## Limites comerciais adotados

- ICP: fundador/responsável técnico de pequeno SaaS B2B, uma aplicação própria e um repositório. Não aceitar stack sem verificador funcional compatível; React na nossa landing não prova capacidade universal de corrigir React de clientes.
- Avulsa: até 3 pedidos, máximo estimado de 4 horas técnicas cada; avaliação de até 20 rotas/2 papéis no fluxo aceito. Escopo maior é outro orçamento, com aceite antes de cobrar.
- Assinatura: 2 pedidos/mês, um ativo por vez, até 4 horas cada; sem acúmulo e sem taxa de setup no padrão. Avaliação contextual e testes fazem parte de cada pedido. Não inclui uma auditoria irrestrita mensal.
- Unidade: uma causa raiz, não cada endpoint repetindo a falha. Reteste e retrabalho do próprio patch não consomem franquia adicional. Se a triagem recusar, a franquia permanece. Trabalho aceito que atravessa o mês não é contado duas vezes.
- Contagem: reserva interna atômica após autorização; início e restituição separados. Primitivas transacionais implementadas em `landing-page/shared/correction-credits.mjs`, sem endpoint público, criação automática de ciclos ou conexão a pagamento. Não oferecer saldo fictício no painel.
- Prazo: triagem até 2 dias úteis; até 5 dias úteis previstos por pedido a partir do início confirmado; até 10 dias úteis para 2 pedidos sequenciais. Avulsa 7–10 dias úteis. Horário comercial de Brasília, feriados conforme contrato. Pausa precisa de motivo, data e novo prazo acordado; não apagar atraso silenciosamente.
- Mensal: cancelar novas renovações sem retirar entregas já aceitas. Anual: 12 ciclos, sem renovação automática. Política de rescisão e restituição proporcional deve ser fechada no contrato antes de cobrança. Nenhum texto desta oferta substitui revisão jurídica.
- Aceite: evidência confirmada, patch/config de segurança, testes/reteste, revisão humana, pacote e instrução de reversão. Falha não resolvida não pode ser vendida como corrigida; registrar e negociar encaminhamento/restituição, não consumir crédito por falha nossa.

## Teste de viabilidade antes de cobrar

**Risco de margem elevado:** R$ 449,10 para até 8 horas técnicas/mês corresponde a R$ 56,14/h de receita bruta antes de qualquer custo adicional. Exemplo orçamentário, não custo medido: 8 h × R$ 120 + R$ 250 de operação + reserva de 20% da receita = R$ 1.299,82; prejuízo de R$ 850,72. A reserva não é alíquota tributária. Fazer dois ensaios cronometrados e medir custos antes de cobrar; se inviável, titular deve aprovar mudança de preço ou franquia. Não reduzi os 2 pedidos/mês nem a revisão humana silenciosamente.

Antes de habilitar cobrança: entidade/contratos, capacidade, host, homologação de login, canal de entrega, catálogo/quotas e conciliação aprovados. Anual não deve ser cobrado antes de provar a capacidade recorrente. A modalidade está definida e visível para proposta, não para checkout imediato.
