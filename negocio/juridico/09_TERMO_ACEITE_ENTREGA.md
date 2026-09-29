# Termo de entrega e aceite

**Cliente:** `[PREENCHER]`  
**SOW/caso:** `[ID E VERSÃO]`  
**Data e fuso:** `[PREENCHER]`

**Revisão técnica final anterior à entrega:** `[RESPONSÁVEL, COMMIT, HASH DO
BUNDLE E REFERÊNCIA DA APROVAÇÃO]`.

Este é o aceite comercial do Cliente. Não substitui a revisão humana técnica
da Contratada nem permite dispensar gates de segurança. Não assinar antes de
conferir o material efetivamente recebido.

## 1. Materiais entregues

| Entregável | Versão/hash | Canal | Recebido |
| --- | --- | --- | --- |
| relatório executivo | `[SHA-256]` | `[CANAL]` | [ ] |
| relatório técnico saneado | `[SHA-256]` | `[CANAL]` | [ ] |
| patch/branch | `[COMMIT/SHA-256]` | `[CANAL]` | [ ] |
| resultado de reteste | `[SHA-256]` | `[CANAL]` | [ ] |
| limitações e pendências | `[REFERÊNCIA]` | `[CANAL]` | [ ] |

## 2. Verificações

- [ ] entregáveis correspondem ao SOW;
- [ ] achados indicam certeza e limitações;
- [ ] correções não incluem mudança funcional não autorizada;
- [ ] testes e gates acordados foram executados;
- [ ] o Cliente recebeu instruções de implantação/rollback;
- [ ] não há promessa de ausência total de vulnerabilidades;
- [ ] pendências e riscos aceitos estão listados abaixo.

## 3. Pendências/ressalvas

`[PREENCHER OU NENHUMA]`

## 4. Decisão

- [ ] aceito sem ressalvas;
- [ ] aceito com ressalvas acima;
- [ ] rejeitado por desconformidade objetiva: `[DESCREVER EVIDÊNCIA]`.

Aceite não transforma o serviço em certificação, não elimina garantia legal e
não aprova item fora do SOW. A janela de suporte corretivo é a do Contrato.

## 5. Próximos passos do Cliente

- implantar após backup e aprovação interna;
- monitorar o ambiente;
- tratar achados não corrigidos;
- atualizar dependências e controles;
- comunicar regressão reproduzível dentro do prazo de suporte.

**Cliente:** `[NOME, CARGO, ASSINATURA]`  
**Contratada:** `[NOME, CARGO, ASSINATURA]`
