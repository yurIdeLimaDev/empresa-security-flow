# Registro simplificado de operações de tratamento

Revisar trimestralmente e sempre que houver novo provedor, finalidade ou fluxo.

Atualização de 11/09/2026: as operações abaixo são modelos a completar, não
registros de ativação. Acrescentar contratação por provedor de assinatura,
categorias mínimas dos signatários, fundamento, país e retenção após escolher
o serviço. Biometria e documentos integrais não são coleta padrão. Registros
de acesso sujeitos ao Marco Civil precisam de categoria própria, conforme `07`.

| ID | Operação | Papel Vexkeep | Titulares/dados | Finalidade | Base | Compartilhamento/país | Retenção | Segurança | Risco e decisão |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ROPA-01 | contato comercial por e-mail | controladora | nome, e-mail, empresa, mensagem | responder e preparar proposta | contrato/preliminares | provedor de e-mail `[PAÍS]` | último contato + 90 dias | MFA, acesso mínimo | `[VALIDAR]` |
| ROPA-02 | conta social | controladora | nome, e-mail, ID do provedor, sessão, IP/UA | autenticação e segurança | contrato/preliminares | Cloudflare, Google/Apple | conta ativa + prazo técnico | tokens cifrados, cookie seguro | ativar só com transparência |
| ROPA-03 | prévia pública | controladora do serviço | domínio e dados técnicos de conexão | entregar consulta e prevenir abuso | solicitação/interesse legítimo | Cloudflare | alvo/resultado não persistidos | SSRF guard, Turnstile, logs mínimos | confirmar logs do provedor |
| ROPA-04 | faturamento | controladora | cadastro fiscal, cobrança, pagamento | contratar, cobrar, emitir nota | contrato/obrigação legal | pagamento/contabilidade | prazo legal | sem cartão na Vexkeep | definir fornecedor |
| ROPA-05 | caso contratado | operadora ou controladora conforme SOW | contas de teste, logs e evidência mínima | avaliar, corrigir e retestar | instrução do Cliente | suboperadores aprovados | política `07` | isolamento, criptografia, hashes | DPA obrigatório |
| ROPA-06 | incidente | controladora/operadora conforme fato | eventos, contatos e dados afetados | conter, comunicar e comprovar | obrigação/defesa de direitos | Cliente, jurídico, autoridade | mínimo 5 anos quando aplicável | acesso restrito e integridade | não guardar além do necessário |

## Campos de revisão

- responsável: `[PREENCHER]`;
- versão/data: `[PREENCHER]`;
- canal do titular: `[PREENCHER]`;
- encarregado indicado ou justificativa de dispensa: `[PREENCHER]`;
- relatório de impacto necessário? `[SIM/NÃO/FUNDAMENTO]`.
