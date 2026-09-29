# Revisão documental da contratação

Data: 11/09/2026.

**Resultado: coerência documental conferida; ponto 3 permanece parcialmente
preparado e bloqueado para liberação comercial.** Não é parecer jurídico,
certificação, aprovação de advogado ou validação de integração em produção.

## Escopo e método

- GraphRAG consultado primeiro. Os trechos não cobriam adequadamente o pacote
  jurídico; foram lidos diretamente os modelos e fontes exatas necessárias.
- Utilizados os modelos Markdown existentes. Não houve criação de DOCX/PDF
  assinável, impressão, assinatura simulada ou envelope em provedor.
- Consultadas fontes oficiais de legislação, STJ e ANPD, além de documentação
  oficial de provedores para avaliar o desenho de assinatura. Referências em
  `negocio/juridico/15_BASES_JURIDICAS_E_LIMITES.md` e procedimento `16`.
- Conferidos `scope.schema.json`, onboarding e validação de SOW no runner para
  não confundir SHA-256 com verificação de assinatura/autoria.
- Mantida esta sessão sem delegação a outro modelo/agente.

## Alterações documentais

- 17 dos 19 documentos originais revisados, incluindo índice, contrato,
  proposta/resumo, SOW, autorização, DPA, retenção, incidente, aceite,
  encerramento, registros, aditivo, checklist e bases.
- Preservados os textos do NDA e dos Termos de Uso públicos: não foram
  convertidos em autorização genérica para o serviço pago.
- Acrescentados três arquivos: procedimento de assinatura, operação de
  cobrança/atendimento e fechamento do ponto 3. Total: 22 arquivos no pacote.
- Atualizados índice de negócio, referência no README geral e roadmap.
- Corrigidos teto automático de responsabilidade, exclusão geral de perda de
  dados, ambiguidade de suporte e referências circulares de hash.
- Registradas decisões pendentes, sem inventar entidade, público, poderes,
  município, profissional contratado, fornecedor ou assinatura.

## Verificações realizadas

| Verificação | Resultado |
| --- | --- |
| 22 arquivos Markdown com título | aprovado |
| Caracteres de substituição de encoding | nenhum encontrado |
| Links locais Markdown do pacote | nenhum caminho inexistente encontrado |
| Marcador acidental de edição no contrato | ausente na versão final |
| Teto automático igual ao valor pago | removido da minuta |
| Exclusão geral de perda de dados | removida da minuta |
| Referência antiga de hash circular `scope.json` no SOW | removida; procedimento de vínculo documentado |
| Estado de assinatura não ativada | explicitamente registrado |
| Estado de ponto 3 não concluído | explicitamente registrado |
| `git diff --check` nos caminhos documentais | sem erro; aviso de conversão LF/CRLF no README |
| Marcador literal `[PREENCHER]` no pacote | 79 ocorrências preservadas em modelos; não são dados preenchidos |

Os checks de texto são verificações mecânicas limitadas e não demonstram
validade jurídica, eficácia de cláusulas ou ausência de todas as inconsistências.
A contagem de placeholders não inclui todos os tipos de campos editáveis.

## Não executado

- aprovação de profissional jurídico/contábil;
- formalização empresarial, compra ou cadastro em fornecedor;
- prova real de representação, assinatura eletrônica ou pagamento;
- teste de emissão fiscal, atendimento ou recebimento/envio de e-mail;
- integração de webhook, verificação criptográfica de assinatura ou checkout;
- alteração no código do runner ou da landing, testes desses programas ou deploy;
- commit, push, envio de documentos a terceiros ou concessão de autorização.

## Condições para concluir

As informações da prestadora, público, profissionais, recebimento, atendimento
e escolha de assinatura precisam ser confirmadas pelo proprietário. Em seguida:
preencher/aprovar documentos, testar provedor e operação, conferir os gates e
só então habilitar o fluxo correspondente. A lista completa está em
[`18_FECHAMENTO_PONTO_3.md`](../../negocio/juridico/18_FECHAMENTO_PONTO_3.md).
