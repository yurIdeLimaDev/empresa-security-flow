# Política contratual de retenção e destruição

**Versão:** `[PREENCHER]`  
**Responsável:** `[PREENCHER]`

Revisão de 11/09/2026: a matriz abaixo continua sendo proposta para aprovação,
não prova de configuração aplicada. Não usar seus prazos para exclusão
automática antes das verificações desta seção.

Prazos contam do evento indicado e representam máximos operacionais, salvo
preservação legal, fiscal, investigação de incidente ou defesa de direito. A
preservação excepcional deve ter fundamento, escopo, responsável e revisão.

| Categoria | Prazo padrão | Início | Destino final |
| --- | --- | --- | --- |
| prévia pública | sem persistência de alvo/resultado | resposta | memória descartada; logs do provedor conforme contrato |
| proposta não aceita | 90 dias | expiração | excluir, salvo pedido ou obrigação |
| contrato, SOW e autorização | 5 anos | encerramento | excluir ou anonimizar, após revisão jurídica |
| dados cadastrais e cobrança | prazo fiscal/contábil confirmado | emissão/pagamento | descarte seguro após obrigação |
| credenciais de teste | até 24h após término técnico | fim do reteste | revogar e apagar imediatamente |
| evidência bruta | 7 dias | coleta | apagar após saneamento e confirmação |
| workspace/cópia de trabalho | 7 dias após aceite | aceite | apagar com verificação |
| relatório saneado | 30 dias após aceite | aceite | entregar e apagar cópia operacional |
| backup cifrado do caso | até 90 dias | criação | expiração automática e teste amostral |
| logs técnicos mínimos | 90 dias | evento | expirar, salvo incidente |
| registros de acesso à aplicação sujeitos ao art. 15 do Marco Civil | 6 meses quando aplicável, salvo preservação válida | evento | descarte após obrigação; não confundir com logs de scan |
| registro de incidente LGPD | mínimo 5 anos | registro | revisão antes do descarte |
| comprovante de exclusão | 5 anos | emissão | descarte após prazo jurídico aprovado |
| CRM | relação + 90 dias | último contato | excluir ou manter com base documentada |

### Decisões obrigatórias antes de aplicar a matriz

1. Distinguir registros de acesso ao site, logs de execução, evidências e
   registros de assinatura. O prazo genérico de 90 dias não substitui a guarda
   exigida pelo art. 15 do Marco Civil quando a operação estiver sujeita a ele.
   A regra legal não exige guardar corpos, credenciais ou resultados de scans.
   [Marco Civil](https://www.planalto.gov.br/ccivil_03/_ato2011-2014/2014/lei/l12965.htm).
2. Resolver a compatibilidade de sete dias de evidência bruta com entrega de
   7–10 dias úteis e reteste/suporte. Definir qual evidência mínima saneada
   preserva a prova e o prazo aprovado; não apagar prova indispensável nem
   manter o bruto indefinidamente.
3. Preservar pacote apresentado, final assinado, anexos, comprovante e vínculo
   de hashes pelo prazo probatório efetivamente aprovado. Cinco anos neste
   modelo não é prazo universal de toda pretensão civil, fiscal ou regulatória.
4. Confirmar a retenção e a exportação dos dados do provedor de assinatura;
   encerrar assinatura comercial do fornecedor não pode eliminar a única prova.
5. Classificar retenções excepcionais com fundamento, acesso restrito e revisão.
   Não usar "defesa de direitos" para conservar indiscriminadamente todo dado.

## Regras

1. Coletar somente o necessário e preferir dados sintéticos.
2. Evidência bruta não entra em Git, e-mail comum, mensageria ou relatório.
3. Backup não prolonga retenção sem limite; deve ter expiração definida.
4. Exclusão lógica sem expiração de backup não é encerramento completo.
5. Credencial é revogada além de apagada.
6. Litígio, incidente ou obrigação pode suspender descarte somente do material
   necessário, com acesso segregado e revisão trimestral.
7. O Cliente recebe comprovante com categorias e datas, não detalhes que
   recriem o dado eliminado.

## Processo de exclusão

1. inventariar cópias ativas, temporárias e backups;
2. interromper jobs e revogar acessos;
3. eliminar workspace, evidência e credenciais;
4. confirmar expiração do backup e eventuais retenções excepcionais;
5. registrar executor, horário, categorias, resultado e exceções;
6. emitir o documento `11_COMPROVANTE_ENCERRAMENTO_EXCLUSAO.md`.

## Aprovação necessária

Os prazos contratuais, fiscais e probatórios devem ser confirmados pelo
advogado e contador antes do primeiro caso. Esta política não autoriza guardar
dados pessoais sem finalidade e base legal atuais.
