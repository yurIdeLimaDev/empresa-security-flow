# Vexkeep: estado atual da implementação

## Correção publicada em 26/09/2026: Turnstile e título

Versão atual: `2cce29f1-11d5-4c15-87ac-8524518e6361`. A validação do
Turnstile usa `redirect: manual`, compatível com workerd, e rejeita redirects.
O formulário renova o desafio após cada tentativa enviada e bloqueia envio duplicado.
As frases animadas estão no título principal; foram retiradas do interior do formulário.
A flag `global_fetch_strictly_public` evita que a leitura pública do próprio domínio
contorne sua rota Worker. Não substitui egress com IP fixado.
Validação: 34 testes aprovados e consultas reais autorizadas a vexkeep.com, incluindo
repetição sem recarregar a página e HTTP 200 na versão final.
[Evidências e limites](../validacao/2026-09-26-turnstile-hero/RESULTADO.md).

## Publicação de 26/09/2026: oferta, acompanhamento e limites

Decisão nova do titular: correção avulsa e assinatura de **2 pedidos/mês**,
mensal ou anual. Substitui a hipótese de venda apenas pontual, não reativa
Check. [Oferta de uma página](../negocio/oferta/OFERTA_MVP_UMA_PAGINA.md),
[pesquisa de preço](../negocio/oferta/PESQUISA_PRECOS_2026-09-26.md) e
[backlog público de segurança](BACKLOG_SEGURANCA_APLICACAO_PUBLICA.md).
Headline/animação atualizados. Portal com sessão, isolamento por proprietário,
prazos/correções e recibo; publicação de metadados operada, não automática.
Fonte: [portal](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/PORTAL_CLIENTE.md). Migrações 0002–0004 aplicadas
no AUTH_DB após validação em staging. Worker publicado na versão
`aea1470c-efd6-4e25-8788-b1a641e81035`, preços confirmados por HTTP.
Preços r2: avulsa R$ 1.350,00, mensal R$ 449,10, anual R$ 5.389,20;
número USD em BRL menos 10%, sem câmbio. Cobrança continua desligada.
Cotas atômicas em D1, prazo global de 30s, leitura limitada e cancelamento
implementados. Franquia tem primitivas internas testadas, não integração financeira.
**Egress/rebinding ainda não resolvido:** validar DNS não fixa IP no fetch.
Validação: [resultado atualizado](../validacao/2026-09-26-seguranca-publicacao/RESULTADO.md),
32 testes, build, lint, tipos e audit de produção sem falhas. Sessões sintéticas
usam Better Auth real e D1 local workerd; consentimento Google real continua pendente.

## Histórico de produção e do motor

Interface em 16/09/2026: botão `Entrar` em destaque à direita do cabeçalho
fixo da landing, validado também no mobile; `/entrar` oferece somente Google. Apple foi retirado
da tela, mantendo backend/contas existentes. Cliente OAuth Google criado segundo
o titular; Client ID e secret configurados em produção, callback salvo no Google.
Login completo ainda não homologado. Publicação via plugin Cloudflare em 16/09/2026,
versão `659d1352-7104-4828-82ba-180c3ce12e2a`, confirmada por HTTP e navegador.
Guia: [ativação Google](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/GOOGLE_OAUTH.md).
O plugin resolveu o acesso à Cloudflare. O console Google foi acessado na conta
confirmada pelo titular e os links de branding foram salvos; app ainda em teste,
callback salvo após resolver a falha anterior. API informa google=true; teste
passou pela confirmação de identidade e chegou à tela de consentimento Google.
Consentimento, retorno, sessão e logout ainda precisam ser validados. O novo destino de
`contato@vexkeep.com` foi cadastrado, mas depende
da verificação do titular. A regra existente foi preservada até essa confirmação.
Evidências e limites no
[relatório local](../validacao/2026-09-16-login-google/RESULTADO.md).

Consolidação do motor em 13/09/2026, com estado da landing revisado em
22/09/2026. Este é o ponto único de consulta do estado atual. Datas em planos,
pesquisas e relatórios anteriores representam o que
foi decidido ou observado naquela ocasião, não uma nova validação de produção.

## Fluxo vigente

Prévia pública gratuita e efêmera no site → proposta → contrato/SOW,
autorização e pagamento → verificação integral governada → correção apenas
de segurança → verificadores independentes → revisão humana técnica final →
entrega criptografada. Um pagamento não concede autorização técnica.

O Check recorrente permanece pausado. Google tem credenciais configuradas e
aguarda consentimento e homologação de sessão; Apple permanece fora da interface.
Conta por senha,
pagamento e assinatura externos não foram ativados nesta etapa. A landing
React possui sete páginas; a coleta pública não persiste domínio/resultado,
enquanto autenticação usa banco e cookies próprios. Não confundir os dois.

## Oito prioridades locais

| Item | Implementação/critério local | Limite externo |
| --- | --- | --- |
| 1. Auditoria | Revisão das fronteiras de rede, autorização, geração, arquivos e entrega; reinstalação limpa, lint, tipos, build, testes e auditoria de dependências incluindo desenvolvimento. | Não é certificação nem teste de produção. |
| 2. Documentação | Este estado central, atualização dos guias operacionais e classificação dos documentos antigos; verificação dos links locais dos diretórios públicos. | Modelos jurídicos ainda precisam de dados e revisão profissional. |
| 3. Gates | Matriz executável de configuração, autorização, allowlist, credenciais, limites, BEST e revisão/entrega. | Não substitui custódia/controle de acesso do operador. |
| 4. POC | Sete cenários de calibração, replay offline, métricas e comparação pelo hash do avaliador; estimativa de custo somente com premissas explícitas. | Não mede capacidade de uma IA real ou cobertura de todas as falhas. |
| 5. CI | Go, race, vet, schemas/exemplos, SBOM de ferramentas, SBOM npm, docs, scripts e ensaio; actions pinadas. | Workflows locais, sem push nem execução remota nesta entrega. |
| 6. Host | Preflight mais estrito, coletor de aceite saneado e roteiro de backup/restore com hashes. | Host definitivo não provisionado nem aprovado aqui. |
| 7. Cliente fictício | Modelos contratuais simulados vinculados por hash → autorização negada/concedida → plano → bundle sintético → geração → gates → revisão simulada → entrega → restore. | Scanners, assinatura, cobrança e IA são explicitamente simulados/ausentes. |
| 8. Limpeza | Removido caminho inseguro de copiar patches externos na entrega; referências atualizadas. Históricos e casos preservados. | Exclusões preexistentes e repositórios de sincronização não foram atribuídos a este lote. |

Resultados efetivamente observados, inclusive falhas e retestes, ficam no
[relatório deste lote](../validacao/2026-09-13-lote-completo/RESULTADO.md).
Comandos: [validação local](VALIDACAO_LOCAL_MVP.md),
[kit de avaliação](../correcao/avaliacao/README.md),
[aceitação do host](../deploy/linux/ACEITACAO.md).

## Entrega de patches: mudança de contrato

`remediation-finalize` deriva `approved-security.patch` do diff cumulativo
baseline → BEST, sem executar filtros Git ou código do cliente. Seu hash entra
na autorização final. `delivery-package` confere esse hash e inclui
`patches/security.patch`; não aceita mais `--patch-root`. O adaptador manual
continua recebendo patches para propor correções, mas não é fonte dos arquivos
entregues. Isso evita substituir silenciosamente um patch depois da revisão e
preserva o resultado cumulativo quando tickets são reabertos.

Autorizações antigas sem `patch_sha256` não são compatíveis com o novo
empacotador: executar a finalização novamente, com aprovação aplicável e destino
novo. Não preencher o hash manualmente. O renderer continua offline e não
precisa de Git; recebe a autorização e o patch canônico já gerados pelo motor.

## Pendências antes de clientes reais

- Selecionar provedor/modelo de IA e condições de tratamento; validar qualidade
  em projetos representativos. Nenhum provedor foi escolhido automaticamente.
- Validar host Linux root, firewall, Docker, proxy, canaries, backup off-host,
  restore e operação do perfil concreto do cliente.
- Concluir identidade empresarial, revisão jurídica/contábil, assinatura,
  pagamento, emissão fiscal e credenciais de autenticação.
- Publicar uma revisão controlada e executar CI/deploy na infraestrutura real.
- Realizar revisão humana final de cada entrega; o ensaio não aprova clientes.

## Autoridade documental

O roadmap de 11/09 mantém o backlog de tecnologia/negócios; os itens locais
concluídos são atualizados por este estado e pelo relatório de 13/09. Pesquisas,
revisões de adoção pinadas por hash e evidências antigas ficam preservadas.
Preços, contratos ou configuração remota antigos não são fatos atuais sem
reconfirmação. O GraphRAG conectado ainda retornou versões antigas nesta
sessão; construir o corpus local não significa atualizar esse MCP remoto/local.

## Documentação e diagramas

Consolidação documental de 13/09: [índice por finalidade](INDICE_DOCUMENTACAO.md)
e [sete diagramas](DIAGRAMAS_MERMAID.md). Os documentos atuais distinguem
implementação, simulação, histórico e dependência externa. A atualização
não altera produção, políticas assinadas, evidências ou modelos de IA.
