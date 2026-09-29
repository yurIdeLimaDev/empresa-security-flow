# Gate jurídico antes de execução

**Caso:** `[ID]`  
**Responsável:** `[NOME]`  
**Data/hora/fuso:** `[PREENCHER]`

Qualquer resposta negativa bloqueia a execução.

**Natureza:** checklist operacional revisado em 11/09/2026; não é integração
executável de assinatura nem parecer jurídico. O runner confere campos e hash
do SOW, mas esse hash não prova assinatura ou poderes. Não marcar itens como
concluídos sem referência à evidência. Item inaplicável exige justificativa;
autorização e identidade nunca podem ser dispensadas por conveniência.

## Partes e pagamento

- [ ] identidade empresarial e signatários conferidos;
- [ ] poderes do representante do Cliente comprovados;
- [ ] proposta e resumo pré-contratual entregues;
- [ ] contrato-mestre assinado;
- [ ] pagamento conciliado conforme a proposta;
- [ ] documento fiscal preparado;
- [ ] eventual arrependimento/cancelamento tratado.

- [ ] público e incidência das normas de consumo avaliados para o caso;
- [ ] suporte, restituição e responsabilidade aprovados sem campos pendentes;
- [ ] provedor/método de assinatura e todos os signatários conferidos;
- [ ] pacote final assinado, comprovante e cópia conservável recuperados;
- [ ] documentos apresentados e assinados vinculados sem hash circular;
- [ ] assinatura não foi inferida de checkbox, retorno de URL ou pagamento;
- [ ] atendimento/cancelamento sem login foi testado conforme `17`.

## Ativos e autorização

- [ ] titularidade ou cadeia de autorização comprovada;
- [ ] SOW assinado e vigente;
- [ ] autorização expressa assinada e vigente;
- [ ] terceiros/provedores autorizaram quando necessário;
- [ ] ativos, portas, identidades, técnicas e efeitos estão fechados;
- [ ] contato de emergência foi testado;
- [ ] janela e fuso foram confirmados no mesmo dia.

## Dados e segurança

- [ ] DPA assinado ou justificativa documentada de inaplicabilidade;
- [ ] papéis controlador/operador definidos por operação;
- [ ] suboperadores e transferências aprovados;
- [ ] dados sintéticos e contas de teste foram priorizados;
- [ ] retenção e exclusão foram comunicadas;
- [ ] backup e rollback do Cliente foram confirmados;
- [ ] credenciais estão no cofre e nunca nos documentos;
- [ ] canal de incidente está disponível.
- [ ] se houver geração externa de patches, fornecedor e envio estão
  autorizados no caso; fontes/metadados foram revisados e a checagem local de
  credenciais foi respeitada. Ausência de alerta não comprova anonimização.

## Integridade técnica

- [ ] SHA-256 do contrato, SOW, autorização, scope e policy conferidos;
- [ ] documento humano e JSON coincidem;
- [ ] ferramentas e versões correspondem à política aprovada;
- [ ] preflight Linux, isolamento, egress e canaries passaram;
- [ ] orçamento de requisições e parada estão configurados;
- [ ] branch/worktree e estado base foram registrados;
- [ ] não há ação irreversível ou deploy em produção não autorizado.

- [ ] `sow_path` aponta para o arquivo final assinado que contém o SOW;
- [ ] o hash pós-assinatura foi calculado sobre o arquivo exato, não sobre a minuta;
- [ ] configuração gerada depois coincide com os limites assinados;
- [ ] nenhuma assinatura/contrato foi usado como substituto da revisão técnica final.

## Decisão

- [ ] **LIBERADO** — todos os gates passaram;
- [ ] **BLOQUEADO** — pendências: `[LISTAR]`.

**Responsável jurídico-operacional:** `[ASSINATURA]`  
**Responsável técnico:** `[ASSINATURA]`

**Referências das evidências e aprovações:** `[MANIFESTO PRIVADO E REFERÊNCIAS]`.
Este registro não vai para Git contendo dados de clientes ou contratos reais.
