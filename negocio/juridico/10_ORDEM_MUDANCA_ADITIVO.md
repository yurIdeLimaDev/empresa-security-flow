# Ordem de mudança e aditivo de SOW

**Contrato:** `[ID]`  
**SOW original:** `[ID, VERSÃO E SHA-256]`  
**Ordem de mudança:** `[ID E VERSÃO]`

## 1. Solicitação

- solicitante e poderes: `[PREENCHER]`;
- motivo: `[PREENCHER]`;
- mudança requerida: `[PREENCHER]`;
- urgência: `[PREENCHER]`.

## 2. Impacto

- ativos adicionados/removidos: `[PREENCHER]`;
- técnicas/efeitos alterados: `[PREENCHER]`;
- dados pessoais/suboperadores: `[PREENCHER]`;
- risco e autorizações de terceiros: `[PREENCHER]`;
- preço: `[PREENCHER]`;
- prazo/janela: `[PREENCHER]`;
- entregáveis/aceite: `[PREENCHER]`;
- rollback/parada: `[PREENCHER]`.

## 3. Documentos substituídos

| Documento | Versão antiga | Nova versão/hash |
| --- | --- | --- |
| SOW | `[ID]` | `[ID/SHA-256]` |
| autorização | `[ID]` | `[ID/SHA-256]` |
| scope.json | `[SHA-256]` | `[SHA-256]` |
| engagement-policy.json | `[SHA-256]` | `[SHA-256]` |
| DPA | `[ID/N/A]` | `[ID/N/A]` |

Nenhuma mudança produz efeito técnico antes de todas as assinaturas e do novo
preflight. Conversa, ticket ou pagamento adicional isolado não altera escopo.

As versões/hashes finais novos são registrados no manifesto de fechamento,
após a assinatura. Dentro do pacote em assinatura, referenciar documentos por
ID, versão e seção; não inserir o futuro hash do próprio pacote nem criar
dependência circular com a configuração operacional. Preservar a versão antiga
e o motivo da substituição, sem sobrescrever documentos assinados.

**Cliente:** `[ASSINATURA, DATA E FUSO]`  
**Contratada:** `[ASSINATURA, DATA E FUSO]`
