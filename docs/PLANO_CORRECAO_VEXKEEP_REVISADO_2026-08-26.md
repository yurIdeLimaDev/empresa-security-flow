# Plano revisado de fechamento do MVP Vexkeep

Registro histórico. Desde 13/09/2026 o estado operacional é consolidado em
[ESTADO_ATUAL.md](ESTADO_ATUAL.md). Preservado para rastreabilidade, não para
reaplicar o Check pausado ou substituir decisões posteriores.

**Estado:** auditoria preservada; jornada de produto parcialmente substituída pelo Pipeline 0.5  
**Data:** 26 de agosto de 2026  
**Escopo:** operação comercial inicial e dependências para os primeiros clientes  
**Regra:** este documento substitui as escolhas conflitantes do arquivo externo `plano-correcao-vexkeep.md`. Não libera produção por si só.

> Decisão posterior: o Vexkeep Check foi pausado e o Pipeline 1 deixou de ser
> gratuito. A landing executa apenas o Pipeline 0.5 efêmero; Pipeline 1 ocorre
> depois da contratação. Para a jornada vigente, prevalecem
> [`PIPELINE_0_5_WEB.md`](PIPELINE_0_5_WEB.md) e
> [`PLANO_VEXKEEP_CHECK_V1.md`](PLANO_VEXKEEP_CHECK_V1.md).

## 1. Conclusão executiva

O plano externo contém uma boa lista de pendências, mas não deve ser aplicado literalmente. Os principais problemas encontrados foram:

- ele desconhecia o Worker, a máquina de estados e a migração D1 já existentes e propunha substituí-los por um modelo inferior;
- tratava o webhook do Asaas como assinatura HMAC, embora a integração oficial use um `authToken` configurado;
- assumia que Asaas eliminaria a validação municipal de NFS-e;
- escolhia Hetzner como se a localização fosse irrelevante, ignorando retenção, transferência internacional e operação de evidências;
- atribuía ao motor capacidade de correção ampla em JS/Supabase que ainda não foi validada;
- aplicava `custo x 3` como regra universal de preço;
- bloqueava execução quando o TLS estivesse inválido, embora observar problemas de TLS faça parte do produto;
- propunha HMAC com timestamp sem nonce como se isso, sozinho, impedisse replay.

Decisão arquitetural: o Pipeline 0.5 continua dentro do Worker do site. O Pipeline 1 integral continua em um executor Linux isolado e sem porta pública, mas só é acionado depois da contratação. Pipeline 2 e correção usam o mesmo tipo de ambiente controlado somente após os gates comerciais e de autorização. Não é seguro remover o backend do Pipeline 1 na implementação atual.

## 2. Pode o Pipeline 1 rodar somente dentro do site?

### Resposta direta

Não, não o Pipeline 1 integral atual.

| Local | O que cabe | Decisão |
| --- | --- | --- |
| Navegador do cliente | Interface e apresentação do resultado | Recusado para executar ferramentas: exporia lógica, políticas e alvos; sofre CORS; não oferece isolamento confiável; permitiria adulteração pelo usuário |
| Cloudflare Worker | Pipeline 0.5 com DNS e HTTP/S controlado | Mantido; publicado e sem persistência do alvo |
| Cloudflare Containers | Binários e tarefas longas, após redesenho | Não adotar agora; o motor atual depende de root, `iptables`, `DOCKER-USER` e redes Docker por engajamento |
| Host Linux isolado | Pipeline 1 integral, Pipeline 2 e correção autorizada | Mantido para o MVP; recebe jobs somente por conexão de saída |

Workers têm limites de memória e CPU e não executam o conjunto atual de binários Docker. Cloudflare Containers oferece VMs isoladas, mas o modo Docker-in-Docker oficial é rootless, não permite `iptables` e exige rede `host` para os containers internos. Isso entra em conflito direto com os controles já validados em `pipeline/internal/app/isolation.go` e `pipeline/internal/app/governed_tools.go`.

Uma versão futura pode empacotar o Pipeline 1 como runtime nativo de um único container e usar a lista de hosts de saída do próprio Cloudflare Containers. Isso é um novo perfil de execução, exige revisão de cada ferramenta e repetição de toda a matriz de isolamento. Migrar agora apenas para eliminar o host descartaria garantias já comprovadas.

Fontes: [limites de Workers](https://developers.cloudflare.com/workers/platform/limits/), [arquitetura de Containers](https://developers.cloudflare.com/containers/platform-details/architecture/), [controle de tráfego de saída](https://developers.cloudflare.com/containers/platform-details/outbound-traffic/), [Docker-in-Docker rootless](https://developers.cloudflare.com/sandbox/guides/docker-in-docker/).

## 3. Auditoria ponto a ponto do plano recebido

### 2. Contrato-mestre, SOW, proposta e autorização

**Veredito:** concordância parcial.

Manter quatro artefatos separados. O SOW deve espelhar os campos operacionais e os hashes precisam ser rastreáveis. Porém:

- checkbox e Turnstile registram uma declaração do solicitante; não provam controle do domínio;
- o Pipeline 1 só executa depois de contratação e autorização; a prova técnica de domínio pode ser exigida no onboarding conforme o caso;
- Pipeline 2 exige autorização expressa assinada, SOW, pagamento e onboarding;
- os hashes do SOW e das políticas pertencem ao fluxo contratado e autorizado; o Pipeline 0.5 não cria SOW nem autorização para etapas profundas;
- os modelos criados no repositório são rascunhos operacionais, não documentos juridicamente aprovados.

**Executado:** modelos organizados em `negocio/modelos/` e matriz para revisão jurídica em `negocio/operacao/`.

**Depende do proprietário:** contratar advogado com experiência em contratos de segurança e privacidade; decidir foro, razão social, limites financeiros e política de reembolso; assinar um piloto ponta a ponta.

### 3. Pagamento, reembolso, conciliação e NFS-e

**Veredito:** manter Asaas como candidato principal, corrigindo a implementação.

- O webhook usa o `authToken` configurado e recebido no header `asaas-access-token`; não há HMAC descrito no protocolo oficial.
- Entrega é pelo menos uma vez; o handler precisa ser idempotente.
- A API usa `access_token`, e não `Authorization: Bearer`.
- O evento de liberação deve ser conferido no modelo do produto. Para o fluxo inicial, `PAYMENT_RECEIVED` é o gate de recebimento; retorno do checkout nunca ativa o caso.
- Consultar o pagamento pela API antes da ativação é defesa adicional útil.
- NFS-e via Asaas existe, mas depende de município, cadastro fiscal, código de serviço e suporte da prefeitura. Não substitui validação do contador.

**Não executado:** integração sem conta PJ, credenciais de sandbox, política de preço e decisão fiscal seria um endpoint incompleto e perigoso.

**Depende do proprietário:** CNPJ/conta PJ, cadastro Asaas, credenciais de sandbox e decisão do contador sobre município, CNAE, código de serviço e emissão.

Fontes: [criação de webhook](https://docs.asaas.com/reference/create-new-webhook), [autenticação de webhook](https://docs.asaas.com/docs/send-types), [autenticação da API](https://docs.asaas.com/docs/authentication), [eventos de pagamento](https://docs.asaas.com/docs/payment-events), [NFS-e](https://docs.asaas.com/docs/emitindo-notas-fiscais-de-servico).

### 4. Preço de uma eventual recorrência futura

**Veredito:** manter pausado. Se a recorrência voltar, substituir `custo direto x 3`.

Usar uma meta explícita de margem de contribuição e capacidade:

```text
custo variável por execução
+ horas humanas por execução x custo-hora sustentável
+ parcela de infraestrutura e suporte
= custo unitário completo

preço de piso = custo unitário completo / (1 - margem de contribuição alvo)
```

Impostos e taxa do provedor devem ser modelados como percentuais ou valores reais, não como multiplicador escondido. Antes de publicar preço do Check, executar 20 casos controlados. Os primeiros dois pilotos podem receber proposta com preço calculado; não é necessário esperar dois clientes pagarem para definir o preço do primeiro.

**Executado:** planilha-base em `negocio/operacao/CUSTO_CHECK.example.csv`.

**Depende do proprietário:** preencher custo-hora, impostos validados, margem alvo e dados das 20 execuções.

### 5. Nicho inicial e prospecção

**Veredito:** rejeitar a restrição a JS/Supabase.

ICP inicial:

```text
SaaS B2B brasileiro com aplicação web pública,
decisor técnico acessível,
um domínio principal,
necessidade concreta de confiança, venda ou lançamento,
e capacidade de provar domínio e assinar autorização quando houver Pipeline 2.
```

O Check é stack-agnostic dentro da superfície pública. A correção automática não é: o perfil validado atual é Python 3.13 com biblioteca padrão. JS/Supabase pode ser qualificado para observação e revisão manual, mas não deve receber promessa de correção automática antes de um adapter e gates próprios serem validados.

**Depende do proprietário:** montar 50 contas manualmente, fazer contatos 1:1 e selecionar dois pilotos pagos sem varredura de terceiros.

### 6. CRM

**Veredito:** adiar HubSpot.

Para dois pilotos, usar o CSV privado versionado fora de evidências técnicas. Isso evita dependência, campos automáticos ainda não definidos e transferência desnecessária de contatos. Migrar para HubSpot, Pipedrive ou equivalente quando houver volume que justifique automação.

**Executado:** `negocio/operacao/CRM_PILOTOS.example.csv` com os campos mínimos e estágios vigentes. A cópia com dados reais deve ficar fora do Git.

**Depende do proprietário:** preencher dados reais e fazer revisão semanal. O arquivo não deve ir a repositório público.

### 7. E-mail profissional

**Veredito:** trocar a decisão automática por custo e necessidade.

Para uma caixa e aliases, a escolha padrão é Exchange Online Plan 1. Microsoft 365 Business Basic só faz sentido se OneDrive/Teams forem necessários; Google Workspace faz sentido se Gmail/Docs forem preferência operacional. Login Google futuro não exige hospedar o e-mail no Google.

Configuração obrigatória em qualquer opção:

- um único SPF coerente com os emissores reais;
- DKIM 2048 quando suportado;
- DMARC começando em `p=none`, com relatórios monitorados, e endurecimento gradual;
- subdomínio separado para envio transacional futuro;
- MFA e conta de recuperação sob controle do fundador.

**Não executado:** o provedor precisa ser contratado para fornecer MX e seletores DKIM exatos. Publicar registros genéricos poderia derrubar recebimento ou autenticação.

Fontes: [Exchange Online](https://www.microsoft.com/pt-br/microsoft-365/exchange/exchange-online-business-plans-and-pricing), [Google Workspace](https://workspace.google.com/intl/pt-BR/business/), [DKIM no Google](https://support.google.com/a/answer/174124), [DMARC no Google](https://support.google.com/a/answer/10032473).

### 8. Jurídico, privacidade e retenção

**Veredito:** manter revisão jurídica, rejeitar a troca imediata de prazos.

Até parecer escrito, manter como proposta operacional:

- prévia gratuita: sem persistência;
- caso ativo: metadados mínimos enquanto ativo;
- relatório saneado: ativo e 30 dias após cancelamento;
- evidência bruta: até 7 dias;
- backup cifrado: máximo de 90 dias;
- financeiro/fiscal: prazo definido por contador e advogado.

Os prazos de 72 horas e 90 dias do plano externo não têm validação operacional. Reduzir evidência bruta para 72 horas pode impedir reteste e suporte; ampliar todo saneado para 90 dias pode aumentar tratamento sem necessidade.

Agentes de pequeno porte podem ter obrigações simplificadas e dispensar encarregado formal, mas precisam de canal do titular e continuam responsáveis por segurança e direitos. Transferências internacionais precisam ser avaliadas nos termos da Resolução ANPD 19/2024.

**Executado:** matriz de decisão jurídica criada; a política pública não foi alterada para afirmar prazos ainda não aprovados.

Fontes: [Resolução ANPD 2/2022](https://www.gov.br/anpd/pt-br/assuntos/noticias-periodo-eleitoral/agentes-de-tratamento-de-pequeno-porte-podem-se-adequar-a-lgpd-por-meio-de-procedimento-simplificado), [Resolução ANPD 19/2024](https://www.gov.br/anpd/pt-br/assuntos/noticias/anpd-aprova-o-regulamento-de-transferencia-internacional-de-dados).

### 9. D1, migração e segredos

**Veredito:** manter a implementação existente; rejeitar o schema substituto.

A migração existente separa casos, eventos, jobs, relatórios, exclusão e recibos e impõe estados por `CHECK`. O schema proposto no plano apagaria gates e colocaria token de prova em texto claro.

**Executado:** 

- criado D1 `vexkeep-check-staging`;
- migração real `landing-page/migrations/0001_vexkeep_check.sql` aplicada;
- leitura e escrita sintéticas de caso/job validadas e removidas;
- binding registrado em `landing-page/wrangler.vexkeep-check.staging.jsonc`;
- adicionados cooldown atômico por origem, heartbeat e contador de falhas consecutivas.

O banco de staging foi criado na região ENAM e deve conter apenas dados sintéticos. Não é uma decisão de residência para produção.

**Não executado:** D1 de produção, Worker de staging, segredos e domínio de rota. A política dinâmica ainda não possui hash final e o runner não existe; configurar placeholder em produção violaria fail-closed.

### 10. Host Linux

**Veredito:** Ubuntu 24.04 LTS e controles são adequados; Hetzner não é a escolha padrão.

Localização não é irrelevante. O host processa alvos, logs e evidência. A escolha padrão do MVP é um host dedicado em região São Paulo de provedor com trilha empresarial, usando AWS Lightsail/EC2 ou outro fornecedor brasileiro aprovado após diligência. Hetzner pode ser reconsiderado somente depois da avaliação de transferência internacional, DPA, suboperadores, custo de saída e resposta a incidente.

O host deve passar, como root:

- drift zero;
- `deploy-preflight-linux.sh` em diretório novo;
- Docker pinado e `DOCKER-USER` presente;
- firewall de entrada fechado, salvo administração restrita;
- canaries, proxy e egress validados;
- backup cifrado fora do host e restauração por hash;
- monitoramento sem dados de cliente.

**Depende do proprietário:** contratar o host e informar o canal de alerta. Não foi feita compra em nome do usuário.

Fontes: [regiões do Lightsail](https://docs.aws.amazon.com/lightsail/latest/userguide/understanding-regions-and-availability-zones-in-amazon-lightsail.html), [preços do Lightsail](https://aws.amazon.com/lightsail/pricing/).

### 11. Runner-agent pull

**Veredito:** manter pull egress-only; não trocar automaticamente por HMAC com timestamp.

O Worker já possui pull, report e failure protegidos por bearer de serviço. HMAC com timestamp e segredo compartilhado não impede replay sem nonce persistido e não melhora o impacto de segredo roubado. Para um agente no MVP:

- TLS obrigatório;
- segredo aleatório com pelo menos 256 bits, armazenado fora do Git;
- rotação e revogação documentadas;
- heartbeat periódico;
- jobs com claim atômico;
- no máximo três falhas consecutivas automáticas;
- cancelamento observado no heartbeat;
- nenhuma porta pública no runner.

Cloudflare Queues com pull consumer é alternativa futura adequada, pois já oferece lease, ack e retry. Migrar agora criaria outra dependência antes do runner estar validado.

**Executado no Worker:** heartbeat, detecção de job stale após três minutos, cancelamento de job ao detectar caso cancelado, retry após uma hora e trava após três falhas consecutivas, além de códigos de falha saneados.

**Pendente:** binário/agente Linux, systemd, execução do preflight, montagem segura da política e envio dos hashes finais.

Fonte: [pull consumers do Cloudflare Queues](https://developers.cloudflare.com/queues/configuration/pull-consumers/).

### 12. Política dinâmica por domínio

**Veredito:** concordância parcial.

- DNS TXT é preferencial e `/.well-known/` continua como fallback; ambos já estão implementados.
- A política deve ser criada para cada execução, com DNS resolvido e IPs públicos pinados antes do preflight.
- O runner precisa revalidar DNS imediatamente antes da execução e comparar o hash recebido.
- Porta 80 pode ser necessária para observar redirect HTTP para HTTPS. Ela não deve ser proibida genericamente.
- TLS inválido não deve bloquear toda execução: é uma condição que `testssl` e a inspeção TLS precisam registrar. A conexão deve continuar limitada ao host e à porta autorizados.
- O canary positivo precisa ser um endpoint controlado alcançável na política; usar o alvo como canary positivo mistura disponibilidade do cliente com funcionamento do isolamento.
- O canary negativo deve ficar fora da allowlist e ser controlado pela Vexkeep.

**Executado:** cooldown atômico de 24 horas depois da validação de Turnstile e do alvo; registros expirados são removidos pelo cron. Sem prova de domínio, o Check completo continua bloqueado.

**Pendente crítico:** construtor determinístico de política e validação no agente. O hash estático de template não pode liberar produção.

### 13. Casos controlados

**Veredito:** ampliar de oito para doze.

1. prova DNS válida e inválida;
2. fallback `/.well-known/` válido e inválido;
3. SSRF por IP privado, resposta DNS mista e mudança DNS;
4. cooldown concorrente do mesmo domínio;
5. claim concorrente por dois agentes;
6. heartbeat, stale e limite de três falhas;
7. Pipeline 1 com preflight novo e hash conferido;
8. rejeição de política com hash divergente;
9. relatório saneado e rejeição de URL/e-mail/saída bruta;
10. cancelamento antes e durante execução, sem resíduos;
11. exclusão, recibo e expiração de backup;
12. reboot/queda do Worker sem rede, container ou chain órfã.

**Pendente:** depende do runner e do host real. Nenhuma evidência deve ser criada artificialmente para marcar esses gates.

### 14. Pagamento Asaas

**Veredito:** corrigir e adiar até sandbox.

Fluxo aprovado:

```text
prova de domínio
-> checkout criado no servidor
-> Asaas envia webhook com authToken
-> validar token por comparação constante
-> deduplicar evento
-> consultar pagamento na API por ID
-> confirmar valor, moeda, vínculo e estado
-> PAYMENT_RECEIVED ativa
-> reembolso/chargeback suspende novas execuções
```

Não chamar isso de webhook HMAC. Guardar hash do payload e identificadores opacos, sem dados de cartão.

**Pendente:** conta, credenciais, preço, sandbox e decisão fiscal.

### 15. Cron, stale e exclusão

**Veredito:** concordância parcial.

**Executado:** cron de cinco minutos no config de staging; sweeper de heartbeat; fila de execuções vencidas; fila de exclusão; limpeza de cooldown expirado.

**Pendente:** agente de exclusão no host, manifesto cifrado/saneado, restore testado e conciliação financeira. A retenção permanece 7/30/90 provisória, não 72h/90d genéricos.

### 16. Deploy

**Veredito:** manter gate, sem deploy público do Check nesta iteração.

Staging D1 não equivale a produto pronto. Produção continua bloqueada por:

- política dinâmica e hash;
- runner Linux e host;
- doze casos controlados;
- pagamento e reembolso em sandbox;
- decisão jurídica e política de privacidade compatível;
- e-mail profissional e canal de suporte;
- preço calculado;
- rollback do Worker e modo park do agente.

A landing e a prévia pública existentes permanecem independentes desses recursos. Não houve regressão no site publicado.

### 17. Fora do MVP

**Veredito:** manter.

Login social, portal completo, e-mail transacional, HIBP no Check, OAST público e correção automática multistack continuam fora. O produto não deve crescer antes de dois pilotos e 20 execuções controladas produzirem dados reais.

## 4. Ordem revisada de execução

1. Proprietário formaliza empresa, conta PJ, contador e advogado.
2. Proprietário contrata e-mail; publicar SPF/DKIM/DMARC com valores reais.
3. Preencher custo e CRM local; selecionar dois pilotos.
4. Contratar host em São Paulo e executar drift/preflight/backup/restore.
5. Implementar runner-agent e política dinâmica por execução.
6. Executar os doze casos controlados e repetir até todos passarem.
7. Criar Asaas sandbox, implementar e testar pagamento/reembolso.
8. Aprovar contratos, retenção e privacidade.
9. Publicar Worker de staging, testar rollback e observar 48 horas.
10. Criar D1 de produção e liberar Check somente com o gate completo.

## 5. Fontes internas usadas

- `docs/PLANO_VEXKEEP_CHECK_V1.md`
- `docs/LANDING_COMERCIAL_E_OPERACAO_PUBLICA.md`
- `docs/ARQUITETURA_IMPLEMENTADA.md`
- `docs/PIPELINE_DE_GERACAO_DE_LEADS.md`
- `docs/ESPECIFICACAO_IMPLEMENTACAO_PRIMEIRO_CASO.md`
- `pipeline/internal/app/isolation.go`
- `pipeline/internal/app/governed_tools.go`
- `landing-page/worker/runtime-template.mjs`
- `landing-page/migrations/0001_vexkeep_check.sql`

## 6. Estado honesto após esta iteração

O D1 de staging e controles do Worker avançaram, mas o Check integral ainda não está pronto para deploy nem para cliente. A lacuna dominante é o executor Linux com política dinâmica validada. Não foi tentado esconder essa dependência dentro do site, porque isso reduziria a segurança já comprovada pelo projeto.
