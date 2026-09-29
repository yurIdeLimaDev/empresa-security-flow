# Cobrança, atendimento e cancelamento

Versão documental: 1.1, de 11/09/2026.

**Estado: procedimento preparado, operação não comprovada.** Preencher os
responsáveis e prazos com o proprietário, advogado e contador. Não divulgar
como serviço ativo antes de testar canais e capacidade de atendimento.

## 1. Cadastro operacional obrigatório

| Item | Informação a confirmar |
| --- | --- |
| Prestadora | `[RAZÃO SOCIAL/IDENTIFICAÇÃO FISCAL E ENDEREÇO]` |
| Responsável por contratação | `[NOME/CARGO/CANAL]` |
| Responsável financeiro/contador | `[NOME/CANAL/MUNICÍPIO E PROCESSO FISCAL]` |
| Recebimento | `[INSTITUIÇÃO, TITULAR E PROCEDIMENTO; SEM SENHAS]` |
| Assinatura | `[PROVEDOR/PLANO/ADMINISTRADOR]` |
| Atendimento | `contato@vexkeep.com`, condicionado a teste de recebimento e resposta |
| Dias/horários/fuso | `[PREENCHER CONFORME CAPACIDADE REAL]` |
| Resposta inicial e acompanhamento | `[PRAZOS COMPATÍVEIS COM A LEI E O CONTRATO]` |
| Emergência durante execução | `[CONTATO PRINCIPAL E SUBSTITUTO]` |
| Responsável por privacidade | `[NOME/CANAL]` |

Não inventar cidade, código fiscal, alíquota, CNPJ, foro, prazo de suporte ou
assinatura. Não inserir dados reais de clientes, documentos ou dados bancários
completos neste modelo versionado.

## 2. Cobrança de projeto pontual

1. Aprovar escopo, preço total, tributos, despesas e marcos antes de enviar a
   versão final para assinatura.
2. Resolver divergências entre o que foi anunciado e o contratado antes do
   pagamento. Não vender uma falha como confirmada com base na prévia pública.
3. Conferir assinatura de todos os participantes e cópia entregue às partes.
4. Emitir cobrança pela instituição escolhida; confirmar nome do beneficiário,
   valor, vencimento e vínculo com a proposta.
5. Confirmar recebimento na instituição, não apenas por comprovante remetido
   pelo cliente. Registrar ID, valor, data, status e responsável.
6. Emitir o documento fiscal segundo a orientação do contador e o procedimento
   aplicável à empresa/município. Não presumir emissão em cada recebimento ou
   determinado código municipal sem confirmação.
7. Concluir autorização técnica, onboarding e preflight antes de executar.

Conciliação assistida pode atender os primeiros pilotos, mas é procedimento
manual. API, webhook e checkout não estão implementados por este documento.
Se forem automatizados, usar o mecanismo oficial do provedor, idempotência e
conferência de vínculo/valor/estado. Pagamento não aprova autorização técnica.

## 3. Registro mínimo de atendimento

Cada demanda deve ter protocolo, data/hora, solicitante, tipo, resumo mínimo,
responsável, próximo prazo, ações e conclusão. Usar os tipos:

- proposta/contratação;
- correção cadastral ou contratual;
- cobrança/nota;
- cancelamento/reembolso;
- revogação da autorização técnica;
- desconformidade da entrega/suporte;
- direitos sobre dados pessoais;
- incidente ou suspeita de comprometimento.

Não incluir credenciais ou evidência bruta no protocolo. Encaminhar material
sensível somente por canal protegido e para quem precisa recebê-lo.

Para relações de consumo abrangidas pelo Decreto 7.962/2013, há exigências de
confirmação imediata de recebimento das demandas e manifestação em até cinco
dias. A capacidade real de atendimento deve cumprir o prazo aplicável; uma
resposta automática de recebimento não é solução da demanda. Não presumir
que vender para CNPJ afasta automaticamente o CDC.
[Decreto oficial](https://www.planalto.gov.br/ccivil_03/_ato2011-2014/2013/decreto/d7962.htm).

## 4. Cancelamento e revogação: fluxos distintos

**Revogação de autorização técnica:** interromper novas ações prontamente,
preservar o mínimo necessário e registrar o ocorrido. Não aguardar solução de
disputa financeira para parar. Retomada exige autorização válida e preflight.

**Cancelamento do contrato:** registrar pedido, interromper futuras atividades
conforme o caso, confirmar recebimento e apurar serviços executados, valores,
devolução e encerramento. Não exigir login ou dificultar o pedido por canal
diferente daquele necessário para contratar.

**Arrependimento quando aplicável:** respeitar o art. 49 do CDC, inclusive
prazo e restituição cabíveis. Não impor multa ou retenção de trabalho como
regra para esse exercício. Solicitar início imediato não equivale a renunciar
ao direito. Na dúvida, não iniciar dentro do prazo antes de definir o
procedimento com advogado. [CDC](https://www.planalto.gov.br/ccivil_03/leis/l8078compilado.htm).

**Cancelamento empresarial fora de hipótese de consumo aplicável:** apurar
proporcionalidade e despesas autorizadas, conforme contrato validado. Havendo
falha da Vexkeep, aplicar os remédios contratuais e legais, sem retenção
automática de todo o valor.

## 5. Reembolso e contestação

1. Conferir identidade de forma proporcional e localizar cobrança/contrato.
2. Registrar motivo, base contratual/legal, saldo e decisão fundamentada.
3. Pedir restituição pelo provedor/meio apropriado; não coletar dados bancários
   completos se a própria transação permitir estorno seguro.
4. Informar valor e previsão real do processamento, sem inventar SLA do banco.
5. Confirmar conclusão e reconciliar taxa, reembolso e documento fiscal com o
   contador. Duplicidade de webhook não pode gerar múltiplos estornos.
6. Guardar comprovante mínimo e manter prazo de retenção aplicável.

Contestação não autoriza reter credenciais ou dados do cliente como garantia,
publicar achados, alterar o sistema ou apagar prova necessária de forma
retaliatória. Aplicar suspensão segura e resolver pelos canais contratuais.

## 6. Suporte e desconformidade

Separar nova solicitação de defeito da entrega. Comparar SOW, commit entregue,
reteste, descrição e evidência. Informar conclusão e, quando cabível, correção,
reexecução, abatimento, restituição ou encaminhamento acordado.

Não exigir que o cliente prove tecnicamente tudo sozinho para receber
atendimento. Pedir somente material necessário e auxiliar a reprodução segura.
Mudança ou patch novo continua sujeito aos mesmos controles técnicos e à
revisão final, sem alterações oportunistas de produto.

Suporte contratual não reduz garantia legal obrigatória. O prazo precisa ser
igual no resumo, proposta, SOW, contrato e página comercial aplicável.

## 7. Modelos de comunicação

### Recebimento

> Recebemos sua solicitação, protocolo [ID], em [DATA/HORA/FUSO]. O responsável
> é [NOME/CANAL]. Retornaremos até [PRAZO APLICÁVEL]. Não envie senhas ou dados
> de produção por e-mail. Se o pedido for parar uma execução em andamento,
> use também [CANAL DE EMERGÊNCIA].

### Contrato concluído

> A assinatura do contrato [ID/VERSÃO] foi concluída. Sua cópia e o comprovante
> estão disponíveis em [CANAL PRIVADO]. O serviço só começa depois das
> condições de pagamento, autorização, onboarding e janela acordadas.

### Cancelamento

> Registramos seu pedido de cancelamento em [DATA/HORA/FUSO]. O estado das
> atividades é [ESTADO REAL]. A apuração é [CRITÉRIO] e retornaremos até
> [PRAZO]. O pedido não depende de criar conta nem afasta direitos aplicáveis.

### Encerramento

> O projeto [ID] foi encerrado. Foram entregues [ITENS], com [RESSALVAS]. Os
> acessos e dados foram tratados conforme [COMPROVANTE], que informa o que foi
> eliminado e eventuais retenções. O canal de suporte é [CANAL/PRAZO].

Enviar apenas depois de conferir os fatos. Modelos não são mensagens enviadas.

## 8. Teste antes do primeiro cliente

- [ ] mensagem externa recebida e resposta autenticada entregue;
- [ ] cobertura do atendimento e emergência confirmada pelo responsável;
- [ ] assinatura de teste e recuperação de cópia exercitadas;
- [ ] cobrança e cancelamento/reembolso testados em sandbox quando disponível;
- [ ] conferência manual registrada se o piloto usar cobrança assistida;
- [ ] fluxo fiscal confirmado pelo contador, sem documento fiscal fictício;
- [ ] pedido sem login consegue atendimento;
- [ ] revogação técnica interrompe execução independentemente do pagamento;
- [ ] demanda e decisão registradas sem informação sensível desnecessária.
