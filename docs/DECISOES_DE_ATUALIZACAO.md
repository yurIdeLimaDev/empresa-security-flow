# Decisões incorporadas

Data: 23 de agosto de 2026.

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
3. O runtime Windows usa Python 3.12, dependências com versão e SHA-256
   fixados e uma chave OpenAI somente na sessão do terminal.
4. O corpus exclui deliberadamente caso CER-Fácil, `correcao/casos/`, binários,
   caches, artefatos e a landing page React. Mudança no corpus invalida a
   consulta até uma reconstrução explícita.
5. A chamada real à API não foi executada nesta máquina, pois não há
   `OPENAI_API_KEY` configurada. Corpus, adaptador Terra e indexação/retrieval
   offline foram validados sem rede.

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

## Fora desta iteração

A modelagem completa de LGPD/base legal e métricas comerciais continua adiada
por decisão do projeto. Isso não torna a automação HIBP juridicamente neutra:
o código está pronto, mas uso real de e-mail deve aguardar a decisão
jurídica/operacional correspondente.

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
