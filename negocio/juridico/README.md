# Pacote jurídico-operacional da Vexkeep

**Versão documental:** 1.1, revisada em 11/09/2026.  
**Estado:** minutas e procedimentos para revisão e preenchimento; contratação
digital e conformidade integral não concluídas.  
**Escopo:** contratação pontual de avaliação e correção de segurança de
aplicações web no Brasil.

Este pacote separa três relações diferentes:

1. uso público de `vexkeep.com`, regido pelos Termos de Uso e pela Política de
   Privacidade;
2. contratação do serviço, regida pela proposta, contrato-mestre e SOW;
3. execução técnica, que só pode começar com autorização expressa, escopo
fechado e regras de engajamento íntegras.

O cliente pode revisar e assinar um único pacote com os anexos identificados.
Isso simplifica a experiência sem transformar os Termos de Uso em autorização
genérica de acesso. Os procedimentos novos são documentais; não foi ativado
provedor de assinatura, checkout ou integração automática.

## Ordem obrigatória

1. preencher e aprovar `00_PENDENCIAS_ANTES_DO_PRIMEIRO_CLIENTE.md`;
2. enviar `01_RESUMO_PRE_CONTRATUAL.md` e a proposta comercial;
3. assinar `02_CONTRATO_MESTRE_SERVICOS_SEGURANCA.md`;
4. assinar `03_SOW_ESCOPO_REGRAS_ENGAJAMENTO.md`;
5. assinar `04_AUTORIZACAO_EXPRESSA_TESTES.md`;
6. anexar `05_ANEXO_PROTECAO_DADOS_DPA.md` quando houver dados pessoais do
   cliente no escopo;
7. obter o pagamento conforme a proposta;
8. executar o preflight jurídico-operacional de `14_CHECKLIST_GATE_JURIDICO.md`;
9. executar, entregar e colher o aceite;
10. encerrar acessos e emitir o comprovante de destruição aplicável.

As assinaturas dos passos 3–6 podem ocorrer no mesmo pacote conforme `16`.
Cada documento precisa integrar o conteúdo aceito e a cópia conservada.

Pagamento isolado **não** autoriza teste ativo. Se contrato, SOW, autorização,
política de engajamento ou respectivos hashes divergirem, a execução deve
falhar fechada.

## Documentos

| Arquivo | Função |
| --- | --- |
| `00_PENDENCIAS_ANTES_DO_PRIMEIRO_CLIENTE.md` | bloqueios que impedem uso real |
| `01_RESUMO_PRE_CONTRATUAL.md` | resumo claro antes do aceite/pagamento |
| `01A_PROPOSTA_COMERCIAL.md` | oferta, preço, prazo e próximos passos |
| `02_CONTRATO_MESTRE_SERVICOS_SEGURANCA.md` | regras comerciais e jurídicas gerais |
| `03_SOW_ESCOPO_REGRAS_ENGAJAMENTO.md` | escopo executável de cada projeto |
| `04_AUTORIZACAO_EXPRESSA_TESTES.md` | autorização do titular dos ativos |
| `05_ANEXO_PROTECAO_DADOS_DPA.md` | controlador, operador, dados e instruções |
| `06_ACORDO_CONFIDENCIALIDADE_MUTUA.md` | sigilo bilateral |
| `07_POLITICA_RETENCAO_DESTRUICAO.md` | prazos e descarte por categoria |
| `08_PLANO_COMUNICACAO_INCIDENTE.md` | resposta e comunicação contratual |
| `09_TERMO_ACEITE_ENTREGA.md` | aceite, ressalvas e suporte curto |
| `10_ORDEM_MUDANCA_ADITIVO.md` | alteração controlada de escopo/prazo/preço |
| `11_COMPROVANTE_ENCERRAMENTO_EXCLUSAO.md` | encerramento de acessos e descarte |
| `12_REGISTRO_OPERACOES_TRATAMENTO.md` | registro simplificado LGPD |
| `13_REGISTRO_SUBOPERADORES_TRANSFERENCIAS.md` | fornecedores e transferências |
| `14_CHECKLIST_GATE_JURIDICO.md` | gate antes de qualquer ação ativa |
| `15_BASES_JURIDICAS_E_LIMITES.md` | fontes oficiais e limites deste pacote |
| `TERMOS_DE_USO_SITE.md` | fonte editorial dos termos públicos |
| `16_ASSINATURA_E_CONTRATACAO_DIGITAL.md` | jornada de assinatura, prova, hashes e critérios da integração |
| `17_COBRANCA_ATENDIMENTO_E_CANCELAMENTO.md` | cobrança assistida/online, atendimento, restituição e modelos de comunicação |
| `18_FECHAMENTO_PONTO_3.md` | estado real, informações faltantes e ordem de conclusão |

## Hierarquia documental

O SOW e a autorização fixam o limite técnico; o DPA prevalece em proteção de
dados e o contrato-mestre rege as condições gerais. Aditivos especificam o que
modificam. Divergências com proposta, anúncio ou outros anexos são resolvidas
antes da contratação, sem afastar direitos obrigatórios. Nenhum documento
comercial amplia autorização técnica por inferência.

## Assinatura e integridade

Use o procedimento `16`: preserve arquivo apresentado e final assinado, seus
hashes distintos, método de autenticação, signatários e trilha de auditoria.
Evite dependências circulares entre o hash do pacote e o da configuração que
aponta para ele. Verifique poderes e autorização de ativos separadamente.

Modalidade avançada precisa ter seus requisitos comprovados; não decorre de
chamar um checkbox de assinatura. ICP-Brasil pode ser usada quando adequada.
Nenhum método é descrito como juridicamente incontestável.

O checklist `14` é operacional. Não verifica assinatura criptograficamente nem
substitui integração de provedor. O hash exigido pelo runner não prova sozinho
que o documento foi assinado ou que o signatário possui poderes.

## Limite importante

Os modelos foram construídos para reduzir lacunas operacionais e contratuais,
mas não substituem a revisão de um advogado brasileiro que conheça a entidade,
o município, os clientes e a operação real. Os campos marcados como
`[PREENCHER]` e os gates do arquivo `00` impedem que o pacote seja tratado como
pronto para assinatura antes dessa revisão.

## Sincronização técnica em 13/09/2026

Os modelos jurídicos mantêm a revisão de conteúdo de 11/09 e continuam não
liberados. Esta atualização sincroniza apenas o estado técnico: o ensaio
contratual é fictício, pagamento/assinatura não foram ativados e os hashes do
runner não comprovam poderes ou assinatura do cliente.

A entrega técnica agora deriva um patch cumulativo de baseline até BEST e
confere seu hash com a autorização final. O aceite comercial do documento 09
continua separado da revisão técnica. Referências:
[estado atual](../../docs/ESTADO_ATUAL.md) e
[diagramas](../../docs/DIAGRAMAS_MERMAID.md).
