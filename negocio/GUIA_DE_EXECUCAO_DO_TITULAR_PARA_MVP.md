# Guia de execução do titular para o MVP

**Pesquisa e orientação originais: 11/09/2026. Sincronização de estado: 13/09/2026.**

A preparação local dos oito pontos foi concluída; não é mais uma tarefa
pendente do titular. Permanecem suas escolhas, contas, aprovações e a
homologação externa. Roteiros prontos: [aceitação do host](../deploy/linux/ACEITACAO.md),
[POC sem fornecedor](../correcao/avaliacao/README.md) e
[estado atual](../docs/ESTADO_ATUAL.md). Preços/fontes históricas não foram
recotados nesta atualização documental.

## Objetivo e premissas

Este guia transforma os nove bloqueios que dependem do titular em decisões e
ações verificáveis. Ele parte do estado atual do projeto: o motor mantém a
geração de patches desacoplada de fornecedor e desativada até haver uma
configuração aprovada; a validação Linux realizada até aqui foi em laboratório
controlado, não em um host comercial de produção; e a entrega permanece
dependente de revisão humana final.

Premissas usadas para evitar suposições perigosas:

- a operação inicial B2B no Brasil é recomendação deste guia, ainda a confirmar
  pelo titular; vender a CNPJ não afasta por si só as normas de consumo;
- a empresa, o município de estabelecimento, o provedor de IA, o provedor de
  assinatura, o meio de pagamento e o host ainda não foram formalmente
  escolhidos;
- nenhuma aceitação de termos no site substitui contrato, SOW e autorização
  expressa para testar ou corrigir sistemas do cliente;
- este é um roteiro operacional informado por fontes oficiais. A escolha
  tributária e a redação contratual final devem ser validadas por contador e
  advogado habilitados.

O critério de conclusão de cada item é deliberadamente documental. Uma conta
aberta, uma chave criada ou uma tela configurada não prova que a operação está
pronta.

## Ordem recomendada

| Ordem | Decisão/ação do titular | Pode ocorrer em paralelo? | Bloqueia |
|---|---|---:|---|
| 1 | Contratar contador e definir a pessoa jurídica | Sim, com 2 e 3 | faturamento, nota e conta PJ |
| 2 | Contratar advogado e fechar a matriz contratual | Sim, com 1 e 3 | vendas pagas e envio de código a IA |
| 3 | Escolher/provisionar o host Linux de produção | Sim, com 1 e 2 | execução real do motor |
| 4 | Selecionar o provedor de IA por prova de conceito sem dados de cliente | Sim, mas só com dados sintéticos | geração automatizada de patches |
| 5 | Autorizar formalmente o tratamento e a transferência de código | Depois de 2 e 4 | qualquer código de cliente na IA |
| 6 | Abrir assinatura eletrônica, pagamento e conta PJ | Depois de 1; em paralelo com 3 e 4 | contratação e onboarding pago |
| 7 | Definir suporte e canal seguro | Em paralelo | operação com cliente |
| 8 | Fechar o pacote do primeiro cliente | Depois de 1, 2, 6 e 7 | execução no ambiente do cliente |
| 9 | Fazer a revisão humana final e autorizar a entrega | No fim de cada caso | entrega ao cliente |

Não habilite a geração por IA nem aceite o primeiro trabalho pago apenas porque
um item anterior parece “quase pronto”. Cada gate abaixo deve estar concluído.

---

## 1. Escolher o provedor de IA, o modelo e o limite financeiro

### Recomendação

Não escolha um fornecedor pelo nome do modelo ou por preço de tabela. Escolha
após uma POC com código sintético e uma matriz de risco. O motor já aceita um
adaptador de provedor; portanto, a melhor decisão agora é preservar essa
neutralidade e registrar um único provedor/modelo por ambiente, com versão
fixada, orçamento limitado e possibilidade de desligamento imediato.

Há dois perfis razoáveis para a comparação inicial:

- **API direta de modelo**: tende a reduzir a complexidade operacional. Por
  exemplo, a OpenAI permite projetos isolados, contas de serviço e limites de
  gasto por projeto, mas a documentação informa retenção padrão de até 30 dias
  para logs de monitoramento de abuso; controles de retenção zero/modificada
  exigem elegibilidade e aprovação. Isso precisa entrar no contrato, não ficar
  presumido. [Projetos e limites](https://help.openai.com/en/articles/9186755),
  [controles de dados](https://platform.openai.com/docs/models/default-usage-policies-by-endpoint).
- **Plataforma de nuvem com IAM**: faz sentido se o cliente ou a futura
  operação exigir controle de identidade, rede e auditoria centralizados. No
  Amazon Bedrock, a AWS afirma que provedores de modelo não acessam prompts e
  conclusões no ambiente Bedrock; ainda assim, retenção, região de inferência e
  comportamento variam por modelo/configuração e devem ser conferidos no
  momento da contratação. [Proteção de dados](https://docs.aws.amazon.com/bedrock/latest/userguide/data-protection.html),
  [retenção](https://docs.aws.amazon.com/bedrock/latest/userguide/data-retention.html).

Não há uma opção universalmente “melhor” antes de saber quais dados poderão
ser enviados, em que países poderão ser processados e qual volume de patches
será necessário. Para o MVP, a escolha deve favorecer: contrato/DPA adequado,
retenção compatível, região conhecida, controle de acesso por conta de
serviço, custo previsível e API compatível com saída estruturada.

### Passo a passo do titular

1. Defina por escrito o orçamento máximo mensal e por caso, incluindo impostos
   e margem para testes. Defina também o responsável que pode aumentar esse
   teto.
2. Monte uma matriz com três candidatos no máximo e os critérios: contrato de
   tratamento/DPA, uso para treinamento, retenção, regiões, suboperadores,
   logs, conta de serviço, limite de gasto, fixação de versão do modelo,
   disponibilidade, preço de entrada/saída e cancelamento.
3. Abra a conta **em nome da futura pessoa jurídica** quando ela existir. Até
   lá, use uma conta de avaliação separada, sem dados reais de cliente e sem
   misturar cobrança pessoal e operacional.
4. Crie um projeto exclusivo, por exemplo `patch-generation-prod`, e uma conta
   de serviço exclusiva. Não reutilize chave pessoal, de desenvolvimento ou
   de outro produto.
5. Restrinja permissões ao mínimo necessário e configure orçamento, alertas de
   gasto, rate limit e procedimento de revogação. Salve o segredo apenas no
   cofre aprovado; nunca em Git, documentação, variáveis de CI expostas ou
   tickets.
6. Execute uma POC com 10 a 20 casos sintéticos representativos: correção
   simples, pedido fora de escopo, código que contém um falso segredo, saída
   inválida, mudança funcional indevida e correção que piora a segurança.
7. Meça taxa de patch válido, taxa de bloqueio correto, custo por caso e
   reprodutibilidade. Não avalie apenas se o texto do patch “parece bom”.
8. Escolha um modelo e um identificador/versionamento que possam ser
   registrados por caso. A troca de modelo é mudança controlada: nova POC,
   nova aprovação e atualização dos documentos.
9. Só então entregue ao responsável técnico a identificação do provedor, do
   modelo, do projeto e o mecanismo de segredo para configurar o adaptador.
   A chave secreta deve ser inserida pelo titular no cofre/host, não enviada
   por chat.

### Evidência de conclusão

- matriz assinada ou aprovada, com data de revisão;
- contrato/termos e política de dados arquivados na versão aceita;
- projeto, conta de serviço, orçamento e alertas configurados;
- POC sintética com resultado, custo e casos bloqueados;
- registro de provedor/modelo permitido, responsável e data de expiração da
  decisão;
- teste de revogação da chave concluído.

### Regra de parada

Se o fornecedor não oferecer condições contratuais e técnicas compatíveis com
o envio de código, não envie código de cliente. Mantenha o fluxo de patches
externos/manual e reavalie outro fornecedor ou uma solução local.

---

## 2. Autorizar o envio limitado de código ao provedor de IA

### Recomendação

Trate isto como uma decisão de dados e de contrato, não como um checkbox
interno. A autorização deve ser específica a cada cliente e a cada fornecedor
que poderá receber trechos de código. Se houver tratamento internacional, o
controlador deve verificar a hipótese legal e o mecanismo aplicável; o
operador deve auxiliá-lo. A resolução da ANPD exige ainda observância de base
legal e dos princípios da LGPD. [Resolução CD/ANPD nº 19/2024](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-19-de-23-de-agosto-de-2024),
 [transferências internacionais](https://www.gov.br/anpd/pt-br/assuntos/assuntos-internacionais/transferencia-internacional-de-dados).

O padrão seguro é **mínimo contexto necessário, sem segredos, sem evidências
brutas, sem URLs de produção e sem acesso do modelo ao repositório ou à rede do
cliente**. A finalidade é propor um patch; a IA não executa comandos nem
aprova a própria alteração.

### Passo a passo do titular

1. Peça ao advogado que classifique os papéis de controlador, operador e
   suboperador para o serviço, em especial quando código, configurações ou
   logs puderem conter dados pessoais.
2. Defina uma política interna de transferência de código com: finalidade,
   classes permitidas/proibidas, provedores aprovados, regiões, retenção,
   suboperadores, prazo de guarda, destruição e resposta a incidente.
3. Atualize contrato-mestre, SOW e anexo de proteção de dados para declarar de
   modo claro: fornecedor de IA, finalidade, tipo de fragmento enviado,
   possibilidade de transferência internacional, medidas de segurança,
   exclusões, direito de recusa do cliente e alternativa sem IA.
4. Para cada cliente, obtenha autorização expressa no SOW/anexo **antes** de
   qualquer envio. A autorização precisa identificar o ambiente, o período e
   o provedor; uma autorização genérica de “usar IA” não basta.
5. Configure o caso para aceitar apenas arquivos da allowlist de contexto.
   Exclua `.env`, chaves, credenciais, dumps, arquivos de produção,
   dependências, testes não necessários, diretórios ocultos e qualquer arquivo
   marcado como confidencial pelo cliente.
6. Aplique redatores/validadores locais antes da transmissão. Se forem
   encontrados segredo, dado pessoal desnecessário ou ambiguidade, bloqueie o
   pedido e encaminhe para decisão humana; não tente “mascarar” sem registrar.
7. Use somente IDs opacos do caso e hashes no pedido. Não anexe relatório,
   URLs, tokens, usuários, cookies ou credenciais.
8. Registre hash do contexto, provedor, modelo, data, política aplicada e
   decisão de envio. Preserve apenas o necessário para auditoria e elimine
   conforme o prazo contratual.
9. Faça um teste com uma amostra sintética contendo segredos falsos e dados
   fictícios; a transmissão deve ser bloqueada e a evidência deve registrar o
   motivo sem imprimir o conteúdo sensível.

### Evidência de conclusão

- parecer jurídico para os papéis e transferência;
- política aprovada e anexo contratual revisado;
- autorização específica assinada para cada caso que usar IA;
- log de contexto por hash e resultado do teste de bloqueio;
- mecanismo documentado de recusa/execução sem IA.

### Regra de parada

Sem autorização específica, sem classificação jurídica ou diante de contexto
que contenha segredo/dado fora da política, a geração por IA fica desligada
para aquele caso. Isso não impede a correção manual nem a revisão humana.

---

## 3. Contratar e configurar o host Linux real

### Recomendação

O teste de integração em laboratório não substitui um host real: rede,
firewall, Docker, proxy, backups, monitoramento e permissões mudam no ambiente
de produção. Contrate um VPS/KVM em conta da empresa, com região e suporte
documentados. Para o MVP, comece com um host de produção e um ambiente de
staging separado ou um segundo VPS temporário para testes de atualização. Não
use o laptop pessoal como host de produção.

Como base operacional, use Ubuntu LTS suportado e Docker do repositório
oficial, com versão deliberadamente escolhida. A própria documentação do
Docker alerta que portas publicadas podem contornar regras do UFW/firewalld e
que regras de filtragem precisam considerar a cadeia `DOCKER-USER`.
[Docker Engine no Ubuntu](https://docs.docker.com/engine/install/ubuntu/).

### Passo a passo do titular

1. Compare no máximo três provedores com estes critérios: região declarada,
   SLA/suporte, IPv4, política de backup e exclusão, 2FA, logs de acesso,
   capacidade de desligamento, exportação de dados, custo de renovação e
   contrato de tratamento de dados. Não decida pelo preço promocional.
2. Contrate em conta corporativa, ative 2FA e cadastre um contato de recuperação
   que não seja apenas um e-mail no mesmo domínio do servidor.
3. Provisione um Ubuntu LTS suportado com recursos suficientes para executar
   os containers e armazenar temporariamente evidências cifradas. Registre
   região, IP, plano, data de renovação e responsável em inventário privado.
4. Crie uma conta administrativa nominal com SSH por chave; desabilite login
   remoto de `root` e autenticação por senha após validar o acesso alternativo.
   Guarde uma chave de recuperação em cofre, com procedimento de acesso de
   emergência.
5. Aplique atualização inicial, sincronização de horário, atualizações de
   segurança, endurecimento SSH e firewall de entrada com regra explícita para
   cada porta necessária. Não exponha painéis de Docker, bancos, métricas ou
   portas de depuração.
6. Instale Docker pelo repositório oficial, fixe as versões aprovadas e
   documente quando serão revisadas. Antes de publicar containers, implemente
   as regras de saída e entrada na `DOCKER-USER`, além da política do host.
7. Configure armazenamento cifrado/segregado conforme o provedor permitir,
   retenção local curta e backup off-host cifrado. Faça um restore com dados
   sintéticos; backup sem restore testado não é aceito.
8. Instale monitoramento de espaço, memória, expiração de certificados,
   resultado de backup e disponibilidade do serviço. Alertas não devem
   conter URL de cliente, segredos, evidências nem dados pessoais.
9. Execute como `root` os preflights e as validações Linux fornecidos pelo
   projeto, incluindo firewall, Docker, proxy e canaries. Guarde as saídas
   saneadas em diretório de validação separado.
10. Faça uma simulação de incidente: revogar uma chave, bloquear egress,
    restaurar backup, reiniciar um container e comprovar que nenhum caso de
    cliente é apagado ou exposto.

### Evidência de conclusão

- contrato/conta do host, inventário e contatos de recuperação;
- resultado aprovado de preflight/drift e das validações de firewall, Docker,
  proxy e canaries no host contratado;
- relatório de portas expostas, versões fixadas e regras de saída;
- backup cifrado com restore sintético verificado;
- teste de alerta e procedimento de incidente.

### Regra de parada

Não execute casos de cliente em host sem validação root real. O resultado do
laboratório é útil para desenvolvimento, mas não prova o ambiente contratado.

---

## 4. Formalizar a empresa, tributação e emissão fiscal

### Recomendação

Comece por contador, não por formulário. Para um serviço B2B de segurança com
potencial responsabilidade contratual, a hipótese inicial mais prudente é uma
sociedade limitada unipessoal (SLU), sujeito à validação do contador e do
advogado. Não abra MEI supondo que qualquer atividade de TI é permitida, nem
escolha CNAE ou alíquota por texto de internet.

O portal Redesim exige consulta de viabilidade, inscrição/registro e licenças
locais. Na abertura, a opção pelo Simples Nacional precisa ser tratada no
momento correto: a Receita informa que empresa em início de atividade deve
formalizar a intenção na inscrição pelo módulo tributário da Redesim, e a
escolha perdida pode postergar a adesão. [Abrir CNPJ](https://www.gov.br/empresas-e-negocios/pt-br/redesim/abrir-cnpj),
 [opção pelo Simples](https://www.gov.br/pt-br/servicos/optar-pelo-simples-nacional),
 [orientação da Receita](https://www8.receita.fazenda.gov.br/SimplesNacional/Servicos/Grupo.aspx?area=1&grp=t).

### Passo a passo do titular

1. Contrate um contador que atenda empresas de serviços B2B e entregue uma
   proposta escrita com abertura, mensalidade, folha/pró-labore, obrigações
   municipais e responsabilidade por prazos.
2. Entregue ao contador a descrição real do serviço: verificação e correção de
   segurança autorizadas, com venda B2B, evidências e eventual uso de IA.
   Peça uma nota técnica com: natureza jurídica, CNAEs principal/secundários,
   regime tributário, inscrição municipal, exigência de certificado e rotina
   de NFS-e.
3. Defina razão social, nome fantasia, endereço da sede, administrador,
   capital social e atividade de fato. Valide previamente a viabilidade do
   endereço e do nome na Redesim.
4. Faça a inscrição/registro com o ato constitutivo revisado e acompanhe o
   protocolo até obter CNPJ e inscrições aplicáveis. A Redesim descreve o
   fluxo de consulta de viabilidade, inscrição e licenças. [Serviço de
   inscrição](https://www.gov.br/pt-br/servicos/inscrever-no-cnpj?id=23662&origem=servico).
5. Faça a opção tributária no prazo correto. Não conclua que o Simples foi
   aceito apenas por ter solicitado; salve o comprovante de deferimento e as
   pendências regularizadas.
6. Abra conta bancária PJ, obtenha o certificado/credencial necessários e
   cadastre a emissão de NFS-e conforme município e regime. O padrão nacional
   atende empresas em municípios conveniados; confirme a situação municipal
   antes da primeira nota. [NFS-e nacional](https://www.gov.br/pt-br/servicos/emitir-nota-fiscal-de-servico-eletronica).
7. Monte um calendário mensal: impostos, pró-labore, declarações, emissão de
   notas, conciliação de pagamentos e guarda de documentos. O contador deve
   confirmar cada obrigação e o responsável.

### Evidência de conclusão

- parecer do contador com natureza, CNAEs, regime e obrigações;
- CNPJ, contrato social/ato constitutivo e inscrições aplicáveis;
- opção tributária comprovada;
- conta PJ e emissor de NFS-e testados com a primeira nota quando cabível;
- calendário de obrigações e responsável por cada uma.

### Regra de parada

Não publique preço final, não aceite pagamento recorrente e não prometa nota
fiscal antes de a estrutura fiscal aplicável estar confirmada pelo contador.

---

## 5. Contratar advogado e fechar os documentos comerciais

### Recomendação

Para o primeiro MVP, venda apenas a pessoas jurídicas até que um advogado
decida de forma diferente. Isso reduz a mistura entre contrato empresarial,
termos públicos do site e regras de consumo. Não use modelo público copiado nem
trate os atuais documentos como prontos para assinatura sem a revisão
profissional.

Selecione advogado brasileiro com experiência comprovável em contratos de
tecnologia, LGPD e testes/correção de segurança. Valide inscrição na OAB pela
seccional competente antes de contratar e fixe o escopo e entregáveis por
escrito.

Assinaturas eletrônicas podem ser juridicamente válidas, mas o nível de prova
importa: documentos ICP-Brasil têm presunção legal; outros meios dependem da
aceitação das partes e da preservação da autoria e integridade. A Lei nº
14.063/2020 define requisitos da assinatura avançada; a MP 2.200-2/2001 admite
outros meios aceitos pelas partes. [Lei nº 14.063/2020](https://www.presidencia.gov.br/ccivil_03/_ato2019-2022/2020/lei/l14063.htm),
 [orientação do ITI](https://www.gov.br/iti/pt-br/acesso-a-informacao/perguntas-frequentes/certificacao-digital).

### Passo a passo do titular

1. Prepare um briefing de uma página: serviço, etapas públicas e autorizadas,
   o que é excluído, como dados/evidências são tratados, uso opcional de IA,
   prazo de guarda, forma de pagamento e objetivo B2B.
2. Avalie três profissionais/escritórios no máximo e contrate por escopo
   fechado. Peça exemplos anonimizados de atuação em contratos de TI e
   privacidade, não apenas certificados.
3. Entregue ao advogado os modelos existentes e solicite, no mínimo:
   contrato-mestre, SOW, regras de engajamento/autorização, DPA/anexo LGPD,
   termo de uso público, política de privacidade, política de retenção e
   destruição, política de IA/transferência, proposta comercial e termo de
   aceite/entrega.
4. Peça que o advogado preencha as lacunas que não podem ser inventadas:
   identificação da empresa, foro, limites de responsabilidade, seguros se
   aplicáveis, prazos, preço, reembolso/cancelamento, propriedade intelectual,
   obrigações de cooperação do cliente e regras de contratação B2B.
5. Exija cláusulas objetivas de autorização: dono dos ativos, ativos exatos,
   período, técnicas permitidas, exclusões, taxas, contatos de emergência,
   encerramento imediato, credenciais de teste e responsabilidade de terceiros.
6. Faça a revisão específica do uso de IA e de transferências internacionais;
   não reutilize a mesma redação sem verificar o provedor escolhido.
7. Defina a hierarquia de documentos: contrato-mestre, SOW, anexos e política
   de engajamento. Preveja que nenhuma condição comercial ou alteração de
   escopo seja válida só por conversa informal.
8. Versione os documentos, calcule hash do anexo de escopo e preserve a versão
   assinada e a trilha de auditoria. Toda alteração precisa de aditivo assinado.

### Evidência de conclusão

- contratação e parecer do advogado;
- conjunto de documentos revisado, versionado e com campos operacionais
  preenchíveis;
- matriz de autorização e regras de engajamento;
- decisão documentada sobre o público inicial e avaliação de consumo aplicável;
- procedimento de assinatura e guarda do pacote probatório.

### Regra de parada

Termos de uso do site não autorizam teste intrusivo nem correção em ativo de
cliente. Sem contrato/SOW/autorização específicos, limite o site a informação
e, se existir, à leitura pública explicitamente descrita pela política.

---

## 6. Escolher assinatura eletrônica, pagamento e faturamento

### Recomendação

Para o MVP, use dois fluxos separados:

1. **assinatura documental** para contrato, SOW, anexos e aceite; e
2. **cobrança** para pagamento e conciliação.

Não condicione execução apenas ao clique de termos ou a uma notificação de
pagamento. A liberação ocorre somente quando houver: documentos corretos
assinados, escopo confirmado, autorização válida e pagamento no estado definido
na política comercial.

Como candidato, a Clicksign tem API para envelopes, autenticação de signatário,
webhooks e validação de HMAC. Isso atende à necessidade técnica, mas a escolha
final depende de preço, contrato e método de identificação aceito pelo
advogado. [Documentação Clicksign](https://developers.clicksign.com/).
Para cobrança, o Asaas oferece sandbox, cobrança Pix/boleto/cartão e webhooks;
o estado da cobrança deve ser acompanhado por webhook, pois criar a cobrança
não confirma pagamento. [Guia de cobranças](https://docs.asaas.com/docs/guia-de-cobrancas),
 [criar cobrança](https://docs.asaas.com/reference/criar-nova-cobranca),
 [eventos](https://docs.asaas.com/docs/webhook-para-cobrancas).

### Passo a passo do titular

1. Peça ao advogado o nível de assinatura necessário para seus contratos e os
   elementos de prova a manter. Prefira solução que preserve PDF final,
   certificado/trilha de auditoria, autenticação usada, data/hora e mecanismo
   verificável de integridade.
2. Crie a conta corporativa do provedor de assinatura, ative 2FA e separe os
   papéis de administrador e remetente quando houver mais de uma pessoa.
3. Assine um contrato de teste entre você e uma conta de teste. Baixe o PDF e
   a trilha de auditoria; valide que o documento e a prova podem ser
   recuperados sem depender da tela do fornecedor.
4. Escolha o provedor de pagamento após CNPJ e validação de KYC. Configure
   sandbox primeiro; não use chave de produção em ambiente local nem no Git.
5. Modele estados internos imutáveis: `proposta`, `aguardando_assinaturas`,
   `assinado`, `aguardando_pagamento`, `pago`, `onboarding`, `encerrado`.
   Evento repetido de webhook deve ser idempotente e nunca duplicar um caso.
6. Valide a autenticidade do webhook segundo a documentação do fornecedor,
   guarde o evento mínimo necessário e compare ID de cobrança, valor, moeda e
   referência do caso antes de alterar o status.
7. Defina com contador e advogado a emissão de NFS-e, estorno, cancelamento,
   desconto, inadimplência e reembolso. Teste o fluxo de uma cobrança em
   sandbox e uma conciliação manual.
8. Somente depois de assinatura, escopo e pagamento confirmados crie o caso
   autorizado. O fluxo público de leitura não deve criar obrigação de serviço
   pago.

### Evidência de conclusão

- contas empresariais com 2FA e responsáveis definidos;
- teste de assinatura com trilha de auditoria recuperável;
- testes em sandbox de cobrança, evento duplicado, pagamento recusado,
  pagamento recebido e estorno/cancelamento;
- política de liberação e conciliação aprovada;
- rotina de NFS-e confirmada pelo contador.

### Regra de parada

Não configure entrega automática após evento de pagamento. Cobrança confirmada
não conserta contrato ausente, escopo ambíguo ou autorização vencida.

---

## 7. Definir suporte, responsável e horários

### Recomendação

No MVP, não prometa suporte 24/7. Ofereça um canal profissional, horário
comercial definido, prazo de confirmação de recebimento e uma rota de
emergência autenticada somente para o contato indicado no SOW. O conteúdo
sensível deve permanecer no repositório/portal de evidências aprovado, não no
corpo de e-mail ou mensageria comum.

### Passo a passo do titular

1. Escolha um provedor de e-mail corporativo que envie e receba de forma
   confiável no domínio da empresa. Configure MFA, contas nominais e SPF,
   DKIM e DMARC; faça testes de envio e recebimento com domínio externo.
2. Crie ao menos `suporte@`, `privacidade@` e um endereço pessoal nominal.
   Não use encaminhamento para e-mail pessoal como canal definitivo.
3. Nomeie o responsável principal, o substituto e o contato de emergência.
   Se não há substituto, declare essa limitação em vez de simular cobertura.
4. Publique internamente e no SOW: horário, prazo de confirmação, prazo de
   resposta inicial, critério de incidente, canal de emergência, exigência de
   autenticação do solicitante e o que fica fora do suporte.
5. Defina uma ferramenta simples de controle de casos que armazene apenas
   metadados: ID do caso, empresa, responsável, prazo, estado e próximo passo.
   Evidências, código e credenciais não entram nela sem avaliação específica.
6. Crie roteiro de incidente: identificar o contato pelo número/e-mail
   registrado, registrar hora, suspender o caso quando necessário, preservar
   evidência e comunicar sem divulgar dados sensíveis.
7. Simule três casos: solicitação normal, pedido de paralisação pelo contato
   autorizado e solicitação por pessoa não autorizada. O terceiro deve ser
   bloqueado e registrado.

### Evidência de conclusão

- e-mail corporativo funcional com autenticação de domínio;
- política de suporte, responsável, substituto e janela de atendimento;
- teste do canal de emergência e dos três cenários;
- canal de privacidade para titulares, como exigido quando aplicável. Agentes
  de pequeno porte podem ter obrigações simplificadas, mas continuam sujeitos
  à LGPD e devem disponibilizar canal de comunicação. [Resolução ANPD nº
  2/2022](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-2-de-27-de-janeiro-de-2022).

### Regra de parada

Sem responsável e canal autenticável, não aceite trabalho que dependa de
paralisação urgente ou contato de segurança do cliente.

---

## 8. Fechar o primeiro cliente com autorização e escopo reais

### Recomendação

O primeiro contrato é uma execução controlada, não uma experiência de venda.
Escolha um cliente com responsável técnico acessível, ativos que ele possui ou
administra e baixa dependência de terceiros. O objetivo não é demonstrar o
maior número possível de técnicas: é provar que a operação consegue cumprir
escopo, produzir evidência e entregar correção sem degradar a segurança.

O NIST recomenda planejamento, condução, análise e mitigação em testes de
segurança, e o OWASP WSTG é uma referência pública para estruturar testes web.
Eles são metodologia; não substituem a autorização contratual.
[NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final),
 [OWASP WSTG](https://wstg.owasp.org/v4.2/2-Introduction/).

### Passo a passo do titular

1. Qualifique o lead: confirme razão social, representante, vínculo com os
   ativos e objetivo do trabalho. Rejeite domínio de terceiro, subdomínio sem
   autorização ou cliente que não indica contato de emergência.
2. Emita proposta curta com serviço, resultado esperado, premissas, duração,
   preço, responsabilidades do cliente e validade. Não prometa que o sistema
   ficará “seguro”.
3. Assine contrato-mestre, SOW, regras de engajamento, anexo LGPD e, quando
   aplicável, anexo de uso de IA. O SOW deve identificar ativos por domínio,
   IP, aplicação, repositório ou ambiente, e listar explicitamente o que está
   fora.
4. Obtenha a autorização de quem tem poder para concedê-la e confirme o
   escopo por segundo canal previamente acordado. Guarde identidade, data,
   assinatura e hash dos anexos.
5. Registre janela, limites de taxa, técnicas permitidas, interrupção/kill
   switch, contatos técnico e executivo, prioridade de comunicação e fluxo
   para incidente inesperado.
6. Crie contas de teste separadas quando a avaliação exigir autenticação.
   Nunca receba senha pessoal, conta de produção compartilhada ou token por
   mensagem comum. Defina entrega segura de credenciais e remoção posterior.
7. Emita cobrança, valide pagamento conforme política e gere a NFS-e no fluxo
   indicado pelo contador. Só então crie o caso de execução.
8. Gere os arquivos de escopo/política requeridos pelo motor a partir dos
   documentos assinados. Compare os hashes antes de iniciar e arquive a versão
   assinada no dossiê cifrado.
9. Faça uma reunião curta de kickoff: confirme escopo, horário, responsáveis,
   parada de emergência e como o cliente receberá achados críticos.
10. Termine com uma reunião de encerramento, aceite, prazo de retenção e
    confirmação de destruição/devolução quando vencer o prazo.

### Evidência de conclusão

- dossiê do cliente com proposta, contrato, SOW, autorização e anexos;
- prova de que o signatário controla os ativos ou tem autorização do titular;
- pagamento e documento fiscal conforme política;
- contatos de emergência testados e política de parada;
- arquivos de escopo gerados dos documentos assinados e hashes conferidos.

### Regra de parada

Escopo ambíguo, autorização ausente, ativo de terceiro, contato de emergência
inexistente ou divergência entre arquivos do motor e SOW são bloqueios. Pare,
corrija a documentação e obtenha novo aceite antes de executar.

---

## 9. Fazer a revisão humana final e autorizar a entrega

### Recomendação

Esta responsabilidade não deve ser automatizada nem transferida ao provedor de
IA. A automação pode preparar patch, executar gates independentes e reter a
melhor versão, mas só uma pessoa responsável pode decidir que a evidência é
suficiente para comunicar um achado e entregar uma alteração ao cliente.

Mantenha uma etapa humana técnica no final, como definido no projeto. O revisor
pode solicitar uma segunda opinião quando não tiver condições de decidir;
isso é exceção, não aprovação intermediária por ticket nem nova exigência
para os três primeiros clientes. Não aprove mudança que não consiga explicar,
conferir e reverter.

### Passo a passo do titular/revisor designado

1. Confirme identidade do caso, escopo assinado, autorização vigente, janela
   respeitada e ausência de mudanças de escopo não documentadas.
2. Leia cada achado contra sua evidência primária. Diferencie observação,
   hipótese, candidato e achado confirmado; elimine afirmações que não possam
   ser reproduzidas ou explicadas.
3. Verifique que relatórios e mensagens não exponham segredos, dados pessoais
   desnecessários, tokens, detalhes exploráveis sem necessidade ou ativos fora
   do escopo.
4. Para cada patch, revise o diff: ele deve tratar uma questão de segurança,
   manter-se nos arquivos autorizados, não introduzir dependência indevida e
   não alterar comportamento de negócio sem autorização do cliente.
5. Confirme os gates independentes: testes relevantes, análise estática,
   verificação de regressão de segurança, reversibilidade e evidência de que a
   nova versão não é pior que a melhor versão anterior.
6. Releia o resumo executivo como o cliente o lerá. Declare limite do escopo,
   data da observação, impacto real, prioridade, correção proposta/aplicada,
   pendências e o que não foi verificado. Nunca afirme garantia de segurança.
7. Decida formalmente: `aprovar entrega`, `devolver para correção`, `descartar
   patch/manter melhor versão` ou `escalar ao cliente`. Registre nome,
   data/hora, motivo e versão/hash da entrega.
8. Entregue pelo canal definido, peça confirmação de recebimento e arquive o
   pacote entregue. Aplique a política de retenção e destruição ao término.

### Evidência de conclusão

- checklist de revisão final preenchido e assinado/registrado;
- versões/hashes do relatório e das alterações entregues;
- decisão explícita de aprovação ou bloqueio e justificativa;
- registro de entrega, aceite/recebimento e prazo de retenção.

### Regra de parada

Se houver evidência insuficiente, regressão, alteração fora de escopo,
ambiguidade técnica relevante ou risco não explicado, não entregue como
concluído. Preserve a melhor versão conhecida e devolva o caso para correção
ou escale a decisão ao cliente.

---

## Checklist mínimo de prontidão para vender o primeiro caso pago

Marque apenas com evidência disponível:

- [ ] Pessoa jurídica, regime tributário, conta PJ e NFS-e definidos pelo
  contador.
- [ ] Contrato-mestre, SOW, regras de engajamento, LGPD, IA e aceite revisados
  pelo advogado.
- [ ] Assinatura eletrônica testada com trilha de auditoria recuperável.
- [ ] Cobrança em sandbox e conciliação por webhook testadas; rotina fiscal
  aprovada.
- [ ] Host Linux real passou preflight e testes de firewall, Docker, proxy,
  canaries, backup e restore.
- [ ] Canal de suporte, privacidade e parada de emergência testados.
- [ ] Provedor de IA escolhido apenas após POC sintética, com limite de gasto
  e política de retenção/transferência aprovadas.
- [ ] Uso de IA de cliente condicionado a autorização específica e validação
  local de contexto.
- [ ] Primeiro cliente possui ativos comprovados, escopo, assinatura,
  pagamento, contatos e autorização válidos.
- [ ] Responsável pela revisão final foi nomeado e conhece o checklist de
  entrega.

## Fontes consultadas

Fontes primárias consultadas em 11 de setembro de 2026. Características,
preços, elegibilidade e termos de fornecedores devem ser reconfirmados antes
da contratação.

- [ANPD — Resolução CD nº 19/2024](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-19-de-23-de-agosto-de-2024)
  e [transferência internacional de dados](https://www.gov.br/anpd/pt-br/assuntos/assuntos-internacionais/transferencia-internacional-de-dados).
- [ANPD — Resolução CD nº 2/2022](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-2-de-27-de-janeiro-de-2022).
- [Redesim — abrir CNPJ](https://www.gov.br/empresas-e-negocios/pt-br/redesim/abrir-cnpj),
  [opção pelo Simples](https://www.gov.br/pt-br/servicos/optar-pelo-simples-nacional)
  e [NFS-e padrão nacional](https://www.gov.br/pt-br/servicos/emitir-nota-fiscal-de-servico-eletronica).
- [Lei nº 14.063/2020](https://www.presidencia.gov.br/ccivil_03/_ato2019-2022/2020/lei/l14063.htm)
  e [ITI — certificação digital](https://www.gov.br/iti/pt-br/acesso-a-informacao/perguntas-frequentes/certificacao-digital).
- [Docker Engine para Ubuntu](https://docs.docker.com/engine/install/ubuntu/).
- [NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final) e
  [OWASP WSTG v4.2](https://wstg.owasp.org/v4.2/2-Introduction/).
- [OpenAI — projetos/contas de serviço/limites](https://help.openai.com/en/articles/9186755)
  e [controles de dados](https://platform.openai.com/docs/models/default-usage-policies-by-endpoint).
- [AWS Bedrock — proteção de dados](https://docs.aws.amazon.com/bedrock/latest/userguide/data-protection.html)
  e [retenção](https://docs.aws.amazon.com/bedrock/latest/userguide/data-retention.html).
- [Clicksign — documentação da API](https://developers.clicksign.com/) e
  [Asaas — guia de cobranças](https://docs.asaas.com/docs/guia-de-cobrancas).

## Relação com o repositório

Este plano considera que o ambiente isolado e o fluxo de execução exigem
validação de rede/Docker e que o onboarding contém requisitos de escopo e
autorização. A prova Linux já disponível é de laboratório controlado; a prova
de produção deve ser gerada no host efetivamente contratado. Fontes internas
consultadas: `pipeline/internal/app/isolation.go`,
[onboarding](../pipeline/config/schemas/onboarding-input.schema.json) e
[ensaio de isolamento](../validacao/2026-08-22-isolamento-politicas/RESULTADO.md).

O reforço executável e os limites da validação atual estão no
[roteiro local](../docs/VALIDACAO_LOCAL_MVP.md). As etapas acima são plano de
execução; assinatura, recebimento, infraestrutura e fornecedor não foram
habilitados por este documento.
