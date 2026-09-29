# Plano de negócios — MVP e primeiros clientes

**Decisão posterior, 26/09/2026:** a oferta inclui correção avulsa e assinatura
mensal/anual com 2 pedidos/mês. ICP/preço/escopo vigentes estão na
[oferta](../negocio/oferta/OFERTA_MVP_UMA_PAGINA.md). As hipóteses pontuais e
preços históricos abaixo não devem ser usados em novas propostas.

Referência histórica de planejamento comercial. Para estado atual e fluxo
operacional, consultar [ESTADO_ATUAL.md](ESTADO_ATUAL.md), consolidado em
13/09/2026, e os modelos canônicos de `negocio/juridico/`. Os trechos antigos
sobre aquisição, preços, fornecedores e Check são hipóteses/backlog, não
instruções para ativação automática ou termos já aceitos por um cliente.

> **Atualização de jornada em 26 de agosto de 2026.** O Vexkeep Check está
> pausado. O funil vigente é Pipeline 0.5 gratuito → contratação → Pipeline 1
> integral → onboarding → Pipeline 2 quando aplicável. Para execução vigente,
> usar `PIPELINE_0_5_WEB.md`, `PIPELINE_DE_GERACAO_DE_LEADS.md` e
> `PLANO_VEXKEEP_CHECK_V1.md`. Permanecem válidos aqui o ICP amplo e as regras
> comerciais para projeto pontual.

**Status:** plano inicial de execução comercial  
**Data:** 26 de agosto de 2026  
**Escopo:** negócio, comercialização e operação do MVP. Não cria produto novo, não redefine políticas técnicas e não substitui aconselhamento jurídico, contábil ou tributário.

## 1. Objetivo comercial

Validar, com clientes pagantes, uma oferta de revisão de segurança para aplicações web brasileiras antes de ampliar escopo, automação de correção ou canais de aquisição.

Resultado esperado em até 90 dias:

- 2 pilotos pagos, com escopo autorizado e entrega concluída;
- evidência de que o comprador entende a proposta, aceita o preço e indica valor prático no relatório;
- processo comercial repetível, documentado e com margem positiva por projeto;
- uma decisão baseada em dados: repetir o nicho/oferta, ajustar a oferta ou interromper a tese.

Não é objetivo nesta fase vender monitoramento contínuo, SOC, pentest irrestrito, garantia de ausência de falhas ou correção automática para qualquer stack.

## 2. Tese de mercado e posicionamento

### 2.1 Tese a validar

Empresas brasileiras de software com operação web exposta precisam de uma revisão de segurança compreensível e limitada, mas não querem contratar uma consultoria ampla, lenta ou opaca. A oferta deve transformar exposição observável e testes autorizados em prioridades de correção e evidências úteis para decisão.

A escolha por empresas digitais não é arbitrária: na TIC Empresas 2024, empresas de **informação e comunicação** têm adoção de nuvem superior à média; entre as empresas de 10 a 49 pessoas, 47% já pagam por armazenamento ou banco de dados em nuvem e 20% por plataforma hospedada de desenvolvimento/teste/implantação. Isso torna o segmento um ponto inicial coerente, não uma estimativa de demanda contratada.[^cetic-cloud]

### 2.2 Posicionamento

Use a seguinte mensagem-base, sem promessas absolutas:

> Revisão de segurança de aplicações web com escopo autorizado, evidências rastreáveis e prioridades claras de correção para equipes que precisam reduzir exposição sem paralisar a operação.

O que diferencia a oferta:

- observação inicial de superfície pública e baixo impacto, seguida de verificação ativa somente após autorização explícita;
- achados sustentados por evidência, em vez de listas genéricas de ferramentas ou uma pontuação sem contexto;
- relatório executivo para decisão e material técnico acionável para a equipe;
- correção somente no perfil técnico que o serviço consegue validar com segurança; os demais casos recebem recomendação e orçamento separado, não uma promessa de automação.

Evite no site, em propostas e em conversas as expressões “100% seguro”, “pentest completo”, “garantido”, “varredura grátis” e “corrigimos qualquer aplicação”.

## 3. Cliente ideal inicial (ICP)

### 3.1 Segmento prioritário

Começar exclusivamente com empresas brasileiras B2B de software — SaaS, softwarehouses pequenas ou agências que mantêm aplicações web de clientes — com:

- 10 a 49 pessoas ou uma equipe técnica pequena, porém identificável;
- aplicação web própria em produção e domínio público;
- CTO, fundador técnico, líder de engenharia ou responsável de TI capaz de autorizar o escopo;
- uso de hospedagem, nuvem ou BaaS detectável e uma dor concreta: venda para empresas maiores, exigência de cliente, incidente recente, troca de fornecedor ou necessidade de priorizar dívida de segurança;
- capacidade de contratar um projeto de escopo fechado em até 30 dias.

### 3.2 Quem não atender agora

Recusar ou adiar, de forma educada, quando houver:

- ativo sem responsável capaz de autorizar formalmente;
- pedido de acesso, exploração ou teste fora do escopo contratado;
- ambiente crítico que não permite janela de teste e não possui interlocutor técnico;
- expectativa de garantia total, resposta a incidente 24x7 ou correção de qualquer tecnologia;
- demanda cujo único decisor é um intermediário sem acesso ao proprietário do ativo;
- organização com exigência regulatória, contratual ou de seguro que o MVP ainda não consegue atender documentalmente.

### 3.3 Lista inicial de contas

Criar uma lista manual de 50 contas, com fonte e data de cada informação. Priorizar indicação, rede pessoal e empresas cujo produto web seja claramente público. Para cada conta, registrar somente:

- empresa, domínio público, segmento, tamanho aproximado e cidade;
- decisor técnico ou econômico em canal profissional público;
- hipótese de dor observável e fonte;
- estágio comercial e próxima ação.

Não fazer coleta massiva de dados pessoais, enriquecimento automático ou contato em escala antes de validar mensagem e conversão com a lista inicial.

## 4. Oferta comercial inicial

### 4.1 Produtos vendáveis

| Oferta | O que inclui | Condição de início | Entrega | Limite |
| --- | --- | --- | --- | --- |
| Diagnóstico de superfície | Observação pública de baixo impacto e perfil de exposição disponível | Interesse do prospect e aceite da observação pública | Resumo de exposição, limites e recomendação de próximo passo | Não afirma vulnerabilidade sem evidência suficiente; não executa teste ativo não autorizado |
| Revisão de segurança autorizada | Pipeline 1, escopo de teste aprovado, Pipeline 2 autorizado e validação de achados | Contrato, pagamento e autorização de escopo concluídos | Relatório executivo, achados validados, evidências e plano priorizado | Um domínio/aplicação e período definidos em proposta; extras são aditivo |
| Correção segura guiada | Correções validadas após a revisão | Só quando o perfil técnico e a autorização forem compatíveis | Mudança proposta, evidência de validação e entrega final | No MVP, correção automatizada é restrita ao perfil de referência **Python 3.13 + biblioteca padrão**; para outros perfis, oferecer recomendações ou escopo manual separado |

O primeiro produto pago deve ser a **Revisão de segurança autorizada**. O diagnóstico de superfície funciona como etapa de qualificação e geração de confiança; não deve virar uma varredura gratuita, ampla e automática.

### 4.2 Estrutura de proposta

Cada proposta deve ter uma página principal e anexos operacionais. A página principal precisa responder, sem jargão:

1. qual ativo será revisado;
2. qual resultado de negócio o cliente receberá;
3. o que está incluído e explicitamente excluído;
4. duração e marcos;
5. preço, condição de pagamento e validade;
6. responsabilidades do cliente (autorização, contas de teste, janela, contato técnico);
7. critérios de aceitação da entrega.

Anexos obrigatórios: escopo autorizado, canais de comunicação, dados de acesso solicitados, tratamento de evidências e processo de mudança quando houver correção.

## 5. Precificação: validar antes de escalar

A landing publica uma única hipótese de lançamento: **a partir de R$ 4.750 por projeto**. O valor corresponde a 5% abaixo da referência brasileira pública de R$ 5.000 para uma aplicação web ou API com prazo de até 10 dias.[^pentestai-preco] Ele não substitui o cálculo de custo, a qualificação e a proposta final.

### 5.1 Cálculo do piso por proposta

Antes de enviar qualquer proposta, preencher:

```text
horas de venda + preparação + execução + validação + relatório + reunião final
× taxa-hora mínima sustentável
+ custos variáveis diretos
+ reserva de risco de 20%
= preço mínimo do projeto
```

A taxa-hora mínima sustentável deve ser calculada uma única vez por mês:

```text
(pró-labore mensal desejado + impostos estimados + ferramentas + infraestrutura
  + reserva operacional mensal)
/ horas faturáveis realistas do mês
= taxa-hora mínima sustentável
```

Use no máximo 40% das horas de trabalho como faturáveis até que exista histórico; o restante será consumido por vendas, preparação, suporte e administração. Não conceder desconto que leve o projeto abaixo do piso.

### 5.2 Regra de lançamento

- Fechar os três primeiros projetos como **pilotos pagos de escopo fechado**. Usar R$ 4.750 somente quando o piso calculado não for maior. Se o custo exigir preço superior, ajustar ou recusar o escopo em vez de vender com margem negativa.
- Cobrar integralmente antes do onboarding nos primeiros projetos pequenos, seguindo o fluxo já definido. Não iniciar Pipeline 2, receber credenciais ou aceitar acesso antes de contrato, pagamento e autorização.
- Vender a correção como projeto único, não assinatura. A página mostra a faixa inicial somente depois da prévia pública; o preço final depende de achado confirmado, esforço e perfil elegível.
- Ao concluir três projetos, comparar horas previstas versus reais, margem, objeções e impacto percebido. Só então definir faixas públicas ou pacotes.

Assinatura fica adiada. HostedScan e Beagle Security demonstram que o modelo recorrente é comum em scanners e monitoramento; Cobalt e Intruder também mantêm preço por teste. O MVP da Vexkeep entrega uma intervenção definida com correção e reteste, portanto o projeto único é mais coerente até existir demanda comprovada por repetição.[^hostedscan-preco] [^beagle-preco] [^cobalt-preco] [^intruder-preco]

## 6. Processo de venda

### 6.1 Funil obrigatório

```text
Conta-alvo → contato personalizado → conversa de descoberta → qualificação
→ aceite de observação pública → Pipeline 1 → reunião de devolutiva
→ proposta fechada → contrato + pagamento + autorização
→ onboarding → Pipeline 2 → relatório/correção aplicável → revisão humana final → entrega
```

O Pipeline 1 não é pulado quando não há dados suficientes: ele deve gerar o perfil/evidência de ausência conforme a implementação atual. Pipeline 2 não é uma continuação automática de Pipeline 1; só começa depois da autorização expressa e das condições comerciais concluídas.

### 6.2 Contato inicial

Meta: abrir conversa, não vender tecnicismo.

Modelo de mensagem:

> Olá, [nome]. Vi que a [empresa] opera [produto/ativo público]. Trabalho com revisões de segurança de aplicações web com escopo autorizado e relatório priorizado para a equipe técnica. Estou conversando com empresas desse perfil para entender como hoje avaliam exposição e exigências de clientes. Faz sentido uma conversa de 20 minutos na próxima semana?

Regras:

- personalizar com um fato público verificável; não alegar falha nem mencionar ferramenta;
- usar canal profissional e identidade clara;
- no máximo duas tentativas adicionais, com encerramento respeitoso;
- registrar resposta e motivo de perda; não insistir após recusa;
- não enviar qualquer evidência de segurança sensível por e-mail sem contexto e confirmação do destinatário.

### 6.3 Roteiro da descoberta (20–30 minutos)

Perguntar e registrar:

1. Qual aplicação/serviço precisa ser protegido e qual mudança ou risco motivou a conversa?
2. Quem é dono técnico e quem aprova compra?
3. Há exigência de cliente, venda, auditoria ou prazo concreto?
4. O que já foi feito e o que seria uma entrega útil para a equipe?
5. É possível definir um domínio/ambiente, janela, contas de teste e responsável para autorização?
6. Qual orçamento e prazo são realistas?

Avançar somente se houver ativo, dor, patrocinador, capacidade de autorizar e cronograma. Se faltar um desses elementos, manter como nutrição ou encerrar.

## 7. Confiança comercial e experiência do cliente

### 7.1 Página e domínio

Ao comprar o domínio:

- escolher nome curto, pronunciável, não dependente de um nicho temporário e alinhado ao nome que será usado na empresa;
- domínio principal decidido e publicado: `vexkeep.com`. Uma eventual compra defensiva de `vexkeep.com.br` deve ser avaliada depois da formalização da empresa; não é requisito para iniciar o MVP;
- registrar o domínio já no titular que deverá possuir o negócio. No Registro.br, o titular pessoa física deve ser a própria pessoa e os contatos de pessoa jurídica devem pertencer à organização; defina isso antes do registro.[^registro-conta]
- manter acesso administrativo, técnico e financeiro em contas controladas pelo fundador e inventariadas; não depender de fornecedor para recuperar o domínio;
- publicar e testar e-mail profissional (`contato@` ou `ola@`) antes de iniciar outbound; configurar autenticação de envio e uma caixa monitorada;
- criar uma landing page de uma página com: promessa, para quem é, como funciona, limites, prova de método, CTA para conversa e contato. Não listar um arsenal de ferramentas nem detalhes que pareçam ataque.

Registro.br informa que domínios sob `.br` exigem contato local com CPF/CNPJ e, ao usar DNS próprio, dois servidores DNS autoritativos respondendo corretamente. Planeje o titular e DNS antes de apontar a landing page.[^registro-regras]

### 7.2 Provas de confiança antes de casos de clientes

Enquanto não houver cases públicos, usar somente provas verdadeiras:

- processo explicável: autorização antes de ações ativas, evidência rastreável e revisão humana final;
- exemplo totalmente sintético/ambiente próprio de como um relatório é estruturado, marcado como demonstrativo;
- política pública curta de escopo responsável e canal de contato;
- perfil profissional do fundador e histórico que possa ser comprovado;
- checklist de onboarding que demonstre previsibilidade, sem expor controles internos sensíveis.

Não usar logos de clientes sem permissão nem fabricar depoimentos, selos ou estatísticas.

### 7.3 Comunicação de entrega

Para cada projeto, estabelecer no kickoff:

- um patrocinador e um contato técnico;
- canal de comunicação, tempo de resposta e rota para incidente;
- marcos: confirmação de escopo, início, bloqueio, prévia executiva, validação final e entrega;
- formato de devolutiva: reunião de 45 minutos com decisor e técnico, seguida de relatório;
- critério de aceite: escopo cumprido, evidências entregues, limitações declaradas e plano de ação priorizado.

## 8. Operação comercial mínima

Criar um CRM único — inicialmente uma planilha privada ou ferramenta simples — com estes estágios: `conta-alvo`, `contatado`, `descoberta`, `qualificado`, `Pipeline 1 autorizado`, `proposta`, `ganho`, `perdido`, `em entrega`, `entregue`, `renovação`.

Campos obrigatórios por oportunidade:

- segmento, origem, decisor, ativo, hipótese de dor, próximo passo e data;
- escopo pretendido, orçamento/faixa discutida e probabilidade;
- autorização da observação pública e, depois, autorização do teste;
- preço, horas estimadas/reais, custos e motivo de ganho/perda;
- feedback pós-entrega e possibilidade de case anonimizado.

Revisão semanal de 30 minutos:

1. nenhuma oportunidade pode ficar sem próximo passo e data;
2. propostas vencidas são encerradas ou reabertas com motivo;
3. comparar capacidade de execução com vendas abertas; não vender mais projetos do que é possível entregar com qualidade;
4. separar dados comerciais mínimos de evidências técnicas e credenciais.

## 9. Indicadores e decisões

Medir desde o primeiro contato, mas não tratar as metas abaixo como verdades universais. Elas são gatilhos de aprendizado para os primeiros 90 dias.

| Momento | Indicador | Decisão associada |
| --- | --- | --- |
| 30 contatos personalizados | respostas e conversas qualificadas | Se não houver conversas, revisar ICP, mensagem e fonte de leads antes de aumentar volume |
| 10 descobertas | dor recorrente, autoridade e objeções | Se a dor não for clara ou não houver decisor, restringir o ICP |
| 5 propostas | taxa de avanço e objeções de preço/escopo | Se todas travarem no mesmo ponto, ajustar produto antes de reduzir preço |
| 3 projetos concluídos | horas reais, margem, satisfação e utilidade do relatório | Só padronizar preço e ampliar divulgação se a entrega for previsível e positiva |
| 1 caso autorizado | prova de resultado | Produzir case anonimizado ou público, conforme autorização, sem revelar dados sensíveis |

Indicadores internos prioritários:

- taxa de conversa por contato;
- taxa de proposta por descoberta;
- taxa de fechamento e ciclo de venda;
- receita contratada, recebida e margem por projeto;
- horas reais versus estimadas;
- entrega no prazo;
- nota de utilidade do relatório (pergunta única de 0 a 10);
- número de ações priorizadas que o cliente decidiu executar.

## 10. Plano de 90 dias

### Dias 1–7: preparar para vender

- definir nome, comprar domínio e configurar e-mail profissional;
- escrever landing page e proposta-base de uma página;
- criar CRM e modelo de descoberta;
- definir taxa-hora mínima, custos e checklist de qualificação;
- confirmar os pré-requisitos operacionais que não podem falhar em cliente real: host Linux dedicado validado, gestão de credenciais/segredos, destinatário de alertas e processo de incidente;
- preparar relatório sintético e página curta de escopo responsável.

### Dias 8–21: descobrir o problema e testar a mensagem

- montar 50 contas-alvo do ICP;
- pedir 10 apresentações a contatos confiáveis e executar 30 contatos personalizados;
- realizar pelo menos 10 conversas de descoberta;
- consolidar objeções, linguagem do comprador, faixa de orçamento e gatilhos de compra;
- ajustar landing page, mensagem e proposta com base nas conversas, não em suposições.

### Dias 22–45: fechar pilotos pagos

- selecionar oportunidades qualificadas e obter aceite para observação pública;
- executar Pipeline 1 somente nas condições previstas;
- conduzir devolutiva e enviar propostas de escopo fechado;
- fechar até dois pilotos pagos sem ampliar escopo para “ganhar” a venda;
- iniciar Pipeline 2 apenas após contrato, pagamento e autorização concluídos.

### Dias 46–90: entregar, aprender e repetir

- concluir os pilotos com relatório e revisão humana final;
- comparar previsão e execução, colher feedback e motivos de decisão;
- pedir autorização para um case anonimizado;
- manter contato de 30 dias para verificar se as prioridades foram executadas;
- decidir: repetir o mesmo ICP, mudar a oferta ou pausar expansão. Não adicionar novos módulos, automações ou canais pagos antes dessa decisão.

## 11. Regras de expansão

Só expandir após três projetos concluídos que comprovem margem, prazo e aceitação do cliente.

- **Novo perfil de correção:** somente se houver demanda repetida, ambiente de validação representativo, política de reversão e prova de que não piora a postura de segurança.
- **Retenção:** oferecer revisão periódica do mesmo escopo antes de criar serviço de monitoramento contínuo.
- **Novos segmentos:** testar um por vez, mantendo a oferta e a métrica comparáveis.
- **Aquisição paga ou escala de outbound:** iniciar somente quando a taxa de conversão e o custo de entrega dos pilotos forem conhecidos.
- **Portal do cliente ou produto SaaS:** adiar até que a operação manual tenha se repetido e os requisitos venham de clientes pagantes.

## 12. Riscos comerciais a controlar agora

| Risco | Controle de negócio |
| --- | --- |
| Escopo cresce sem receita | Proposta com ativo, prazo, exclusões e aditivo para qualquer expansão |
| Compra por medo, mas expectativa irreal | Linguagem sem garantias e devolutiva com limites claros |
| Projeto tecnicamente complexo demais | Qualificação e recusa quando o perfil não cabe no MVP |
| Tempo de entrega consome toda a margem | Registro de horas reais; parar e rever preço/escopo após cada piloto |
| Confiança insuficiente em empresa nova | Processo transparente, exemplo sintético e contato direto do fundador |
| Dependência de um único canal | Priorizar indicações e outreach manual, mas registrar origem e diversificar após validação |
| Dados e credenciais tratados de forma improvisada | Não iniciar onboarding sem fluxo operacional aprovado; minimização de dados e separação de evidência técnica/comercial |

## 13. Decisões que este plano fixa

1. O MVP é **serviço especializado**, não plataforma de segurança genérica.
2. O cliente inicial é B2B de software com aplicação web e interlocutor técnico, não qualquer pequeno negócio.
3. A primeira venda é uma revisão autorizada de escopo fechado; diagnóstico público serve para qualificar e demonstrar contexto, não para criar um produto gratuito de massa.
4. Correção automatizada só é vendida dentro do perfil de referência já validado; fora dele, a resposta é recomendação ou escopo manual separado.
5. O negócio cresce por pilotos pagos e evidência de margem, não por volume de scans, ferramentas ou leads.
6. A passagem para Pipeline 2 continua condicionada a contrato, pagamento e autorização; ausência de dados no Pipeline 1 não cria atalho.

## 14. Pesquisa que embasa as escolhas

[^pentestai-preco]: [PentestAI, pentest por R$ 5.000](https://pentestai.com.br/), consultado em 26 ago. 2026. É preço publicado por concorrente e serve como referência de posicionamento, não como fonte neutra de custos.

[^hostedscan-preco]: [HostedScan Pricing](https://hostedscan.com/pricing), consultado em 26 ago. 2026.

[^beagle-preco]: [Beagle Security Pricing](https://beaglesecurity.com/pricing), consultado em 26 ago. 2026.

[^cobalt-preco]: [Cobalt Pricing](https://www.cobalt.io/platform/pricing), consultado em 26 ago. 2026.

[^intruder-preco]: [Intruder Pentest Pricing](https://www.intruder.io/pentest-pricing), consultado em 26 ago. 2026.

[^cetic-cloud]: [Cetic.br, TIC Empresas 2024 — indicador B18](https://cetic.br/pt/tics/pesquisa/2024/empresas/B18/expandido/), publicado em 2025. Pesquisa nacional com empresas formais de 10 ou mais pessoas ocupadas. Os números justificam priorizar empresas digitais de 10–49 pessoas; não provam intenção de compra.

[^registro-conta]: [Registro.br — gerenciamento de conta](https://registro.br/ajuda/gerenciamento-de-conta/). Consultado em 24 ago. 2026.

[^registro-regras]: [Registro.br — regras para registro de domínios](https://registro.br/dominio/regras/) e [registro de novos domínios](https://registro.br/ajuda/registro-de-novos-dominios/). Consultados em 24 ago. 2026.

Referências complementares para a mensagem de valor e avaliação de fornecedores:

- [CISA — recursos para pequenas e médias empresas](https://www.cisa.gov/small-and-medium-sized-business-resources): o material reforça a necessidade de práticas e avaliações proporcionais, mas não é base para promessa comercial no Brasil.
- [CISA — avaliação de fornecedores por PMEs](https://www.cisa.gov/resources-tools/resources/assisting-small-and-medium-sized-businesses-assess-vendors-and-suppliers-fact-sheet): reforça que compradores de serviços críticos precisam avaliar fornecedor e acesso; por isso o plano prioriza escopo, controles e prova verificável.

## 15. Próxima ação única

Com `vexkeep.com` publicado, escolher e configurar o provedor de envio de e-mail profissional antes de iniciar outreach. Em seguida, executar os itens dos dias 1–7 sem criar novas ofertas ou ferramentas.
