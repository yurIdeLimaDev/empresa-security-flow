# Decisões incorporadas

Consolidação: 13/09/2026. As seções datadas preservam decisões históricas;
quando houver evolução, vale o estado abaixo e o [estado atual](ESTADO_FLUXO.md).

## Decisões vigentes consolidadas

- Geração independente de fornecedor implementada, desativada até escolha e
  homologação; testes usam simulador/replay, sem outro modelo de IA.
- Finalização gera patch cumulativo de BEST e o vincula à autorização por hash;
  o empacotamento não recebe mais patches externos.
- Oito prioridades locais validadas, incluindo gates, POC, CI, ensaio fictício
  e preparação de aceitação do host. Não significam release ou aceite externo.
- Google/Apple possuem backend implementado; credenciais e homologação faltam.
  Compra, assinatura e conta por senha não estão ativadas.
- Minutas jurídicas/procedimentos já existem; revisão profissional e dados
  reais continuam pendentes. O adiamento de agosto não descreve mais a
  preparação documental atual.
- A foca vigente na landing é vermelha, transparente e sem pedra; a decisão
  visual anterior foi substituída pela escolha posterior do proprietário.
- Check recorrente continua pausado. Nomes de pipelines são internos.
- O MCP consultado pode apontar outra cópia: trechos antigos exigem confirmação
  direta nos arquivos, não mudança de modelo.


## Fonte de verdade

As decisões e anexos enviados nesta iteração prevalecem sobre versões antigas
do repositório. A implementação interrompida foi comparada com elas antes de
ser retomada.

## Consulta Graph RAG

1. HippoRAG 2 foi adotado somente como camada interna de consulta sobre o
   fluxo de verificação/correção, nunca como fonte de verdade ou autorizador.
2. A origem fica pinada no commit
   `2f52a86dd04e4633703bd2fb3bb6a37683ac3cfb` do repositório canônico
   `OSU-NLP-Group/HippoRAG`; o commit não possui assinatura Git verificável.
3. O runtime Windows usa Python 3.12 e dependências com versão e SHA-256
   fixados. A consulta não requer chave OpenAI: o Codex chama um servidor MCP
   local, que executa HippoRAG com extração de relações e embeddings
   determinísticos, sem rede.
4. O corpus exclui deliberadamente caso CER-Fácil, `correcao/casos/`, binários,
   caches, artefatos e a landing page React. O MCP confere o fingerprint em
   cada consulta e reconstrói o índice automaticamente quando o corpus muda.
5. O perfil anterior que dependia de Terra/API foi removido do caminho de
   operação. O MCP oferece somente recuperação; o modelo do Codex formula a
   resposta com os trechos retornados.

## Fluxo e ferramentas

1. O Pipeline 1 permanece obrigatório e com o mesmo papel comercial/técnico.
2. O nome vigente é **superfície pública e baixo impacto**, pois GET/HEAD,
   httpx e testssl fazem conexões.
3. Hadrian v1.0.0 substitui diff genérico em BOLA/BFLA.
4. Scenario runner próprio fica somente em crédito/billing.
5. AuthProbe é opcional e mutuamente exclusivo com Hadrian, mas continua
   desligado porque não há origem canônica verificada para a ferramenta
   descrita.
6. Supabase-RLS-Checker fica manual/GUI no Pipeline 2 e usa o repositório
   canônico `sahilahluwalia`; anon key isolada não prova falha de RLS.
7. keyleak-detector fica desligado no Pipeline 1 e no Pipeline 2 até validação
   do rollback. rlsgate fica futuro/NDA.
8. **dnsReaper passou a automático** no Pipeline 1, somente sobre a saída curada
   de Subfinder/Amass, sem reivindicar recurso e sem promover candidato a
   takeover confirmado.
9. **HIBP passou a automático** no Pipeline 1, somente por e-mail corporativo
   publicado e pelo endpoint breachedaccount. Busca por domínio é
   estruturalmente proibida no runner.
10. TruffleHog passou a automático no P1 apenas sobre bundles locais e sempre
    com `--no-verification`.

## Decisões de segurança front-loaded

- Não há aprovação humana por execução. Lead/scope e três políticas concentram
  as decisões do onboarding.
- Pipeline 1 usa reconhecimento público formalizado; Pipeline 2 usa autorização
  explícita.
- A matriz pode ser vazia quando nenhuma ferramenta ativa de dados está
  habilitada. Ferramenta futura que compare identidades ou altere estado só
  poderá ser automatizada com matriz não-vazia e nova revisão de adoção.
- NS autoritativo guarda host e IP, porque o firewall trabalha por IP e a
  evidência precisa preservar o nome.
- Todos os campos de flags são obrigatórios no JSON; não se depende de default
  implícito.
- O canary negativo aponta para endpoint controlado pelo operador e é testado
  pelo host antes de se exigir que a rede o bloqueie.
- O proxy é auditoria. O firewall continua a barreira principal.

## Correções feitas durante a implementação

- O SARIF observado no branch principal do Hadrian não existe na release
  v1.0.0; foi mantido o JSON da release pinada.
- O repositório `authprobe/authprobe` encontrado é OAuth para MCP, não a
  pré-checagem de BOLA descrita, e foi rejeitado.
- A validação DNS no transporte passou a conectar somente ao IP já aprovado.
- O acesso do container ao host Docker foi fechado com chain `INPUT`; somente
  as portas do proxy de auditoria são liberadas.
- O executor antigo de binário local, não utilizado, foi removido. Execução
  externa agora é Docker-only.
- Dalfox v3 usa o subcomando `scan`; a forma antiga não foi mantida.
- Supabase-RLS-Checker possui `pnpm-lock.yaml`, então o comando correto é
  `pnpm install --frozen-lockfile`, não npm.
- Taxas fracionárias nunca são truncadas silenciosamente.

## Escopo adiado na iteração de agosto

A modelagem completa de LGPD/base legal e métricas comerciais foi adiada
naquela iteração. Desde 11/09 existem minutas e procedimentos, ainda sem
aprovação profissional e operação real. Isso não torna a automação HIBP juridicamente neutra:
o código está pronto, mas uso real de e-mail deve aguardar a decisão
jurídica/operacional correspondente.

## Registro histórico de 26 de agosto de 2026: landing e conversão

1. A entrada da landing passou a ser o domínio do próprio solicitante.
2. O site executa somente uma prévia verdadeira e limitada do Pipeline 1. O
   runner integral continua obrigatório e ainda não foi ligado à requisição web.
3. A oferta aparece somente depois do resultado e descreve correção, limites e
   prazo conforme os sinais observados.
4. O preço de lançamento ficou em R$ 4.750 por projeto, 5% abaixo da referência
   brasileira pública de R$ 5.000. O piso sustentável calculado antes da
   proposta continua soberano.
5. Foi escolhido projeto único, não assinatura, até existir histórico de
   recorrência, margem e capacidade de entrega.
6. Na decisão original, conta Google, Apple, e-mail e pagamento eram cascas.
   Posteriormente Google/Apple receberam backend; credenciais/homologação,
   pagamento e conta por senha continuam pendentes. Não simular ativação real.
7. O caminho sem conta será compra como convidado. A comunicação não promete
   anonimato financeiro.
8. A primeira escolha visual sem vermelho foi substituída pela foca vermelha
   escolhida posteriormente, sem pedra e com fundo realmente transparente.

## Decisões da correção

1. O diagrama de correção foi aceito parcialmente.
2. A revisão humana intermediária por ticket foi removida. O único gate humano
   é final, depois da validação global.
3. O MVP executa tickets serialmente sobre uma única `BestRef`. Paralelismo foi
   adiado porque merges concorrentes podem violar a garantia de não regressão.
4. Um candidato só substitui `BEST` se todos os gates passarem, o diff respeitar
   o escopo e o bundle canônico tiver cobertura idêntica sem novo achado nem
   piora de severidade/rank.
5. Versão rejeitada permanece como evidência, mas nunca substitui a melhor.
6. Limites por ticket/caso impedem loop infinito.
7. `changes_requested` na revisão final reabre apenas tickets explícitos e não
   remove os limites nem a regra de não regressão.
8. Alterações de produto, interface, conteúdo ou arquitetura sem necessidade
   direta para a correção de segurança continuam proibidas.
9. O provedor de IA é pluggable, mas o adaptador é executado pela CLI com
   SHA-256 pinado, timeout, ambiente reduzido, worktree e commit automático.
10. Um preflight obrigatório resolve e confere o SHA-256 do agente e de todos
    os gates antes do plano; hash placeholder ou executável ausente bloqueia.

## Fechamento das pendências operacionais

1. As 28 entradas receberam revisão de adoção no commit exato: 20 aprovadas
   somente para execução governada e oito restritas.
2. Toda entrada automática possui imagem por digest e liveness; ferramentas
   restritas não recebem container artificial apenas para completar contagem.
3. Subfinder, Hadrian, jsluice, sqlmap e o runner possuem receitas de build
   reprodutível; artefatos locais são conferidos por SHA-256.
4. O runtime aplica rootfs read-only, capabilities zeradas,
   `no-new-privileges`, limites de recursos, timeout, logs limitados e contrato
   de artefatos.
5. O ciclo Docker/firewall/proxy/canaries passou em Linux root descartável. A
   repetição no host escolhido virou gate de deploy, porque não é possível
   certificar antecipadamente um host ainda não definido.

## 24 de agosto de 2026 — primeiro caso

- um único perfil Python 3.13 stdlib; múltiplas stacks continuam fora;
- adaptador manual por patch externo, sem LLM remoto;
- onboarding genérico usa hashes sentinela para bloquear preflight acidental;
- age 1.3.1 pinado em `go.mod`, `go.sum` e SBOM;
- ZIP determinístico; ciphertext não determinístico por segurança;
- aprovação humana somente no final, vinculada a commit e bundle;
- laboratório privado, sem dados de terceiros nem rede pública;
- todo host dedicado continua obrigado a seu próprio drift/preflight.
