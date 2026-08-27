# Pipeline de geração de leads

Estado vigente em 26 de agosto de 2026.

```text
Lead -> Landing page -> Pipeline 0.5 web gratuito
     -> resultado efêmero -> proposta
     -> contrato + pagamento -> Pipeline 1 integral
     -> evidência + perfil de stack JSON -> onboarding técnico
     -> Pipeline 2: módulos autorizados e roteados -> relatório final
```

O Pipeline 1 não pode ser esvaziado ou convertido diretamente no Pipeline 2.
Mesmo sem tecnologia/achado conclusivo, sua saída registra o que foi observado,
as limitações e o perfil `unknown`. A especialização ocorre depois.

## Entrada pela landing: Pipeline 0.5

`vexkeep.com` permite informar um domínio e receber na tela uma observação real
da superfície pública. O Worker consulta DNS, HTTP/HTTPS, oito cabeçalhos,
CSP, atributos de cookies sem valores, SPF/DMARC/MX/DNSSEC/CAA, security.txt,
uma amostra limitada de Certificate Transparency e até dois bundles JavaScript
do mesmo host para fingerprint textual. Corpos são limitados e o JavaScript do
alvo nunca é executado.

Essa resposta é o **Pipeline 0.5 web**, não o Pipeline 1. Ela não substitui o
runner integral com ferramentas em container, isolamento Linux, normalização e
evidência governada. Não há login, porta, formulário, credencial, exploração,
mutação ou persistência do domínio/resultado. O orçamento máximo é de 46
subrequisições externas e até oito observações. O site não inicia Pipeline 1
nem Pipeline 2.

O preço, prazo e limites comerciais aparecem somente depois da prévia. A
implementação e a documentação da interface pública são mantidas separadamente
no repositório privado da landing.

## Pipeline 1 — superfície pública e baixo impacto

**Objetivo:** produzir evidência defensável do que um visitante anônimo
consegue observar, sem autenticação, exploração, mutação ou contorno de
controle. Não é passivo estrito: há conexões HTTP/TLS/DNS de baixo impacto.

### Entrada e política

- domínio/base URL, budget, taxa e limite de download;
- `engagement-policy` com observação pública reconhecida, host/IP/porta,
  DNS/APIs externas e canaries;
- `tool-policy` com ferramentas P1;
- `test-matrix` pode estar vazia porque P1 não realiza teste ativo de dados;
- tudo é validado antes de criar uma execução.

### Descoberta

- Subfinder e Amass em modo passivo;
- Certificate Transparency somente se `crt.sh` estiver explicitamente em
  `external_apis`;
- DNS via resolver aprovado;
- dnsReaper automático somente sobre os subdomínios válidos produzidos por
  Subfinder/Amass. Paralelismo 10, sem domínio inteiro e sem claim. Saídas
  `CONFIRMED`/`POTENTIAL` ainda são candidatas; `UNLIKELY` fica separado
  para revisão adicional.

### Fingerprint e HTTP

- httpx e fingerprint nativo;
- GET em caminhos públicos previstos e HEAD da raiz;
- CORS só com origem controlada;
- host, porta e IP de conexão pinados; redirects, taxa, timeout e budget
  limitados;
- bundles JavaScript somente do mesmo host e dentro do limite de bytes.

### Análise local de bundles

- jsluice e Gitleaks offline;
- TruffleHog automático apenas em `bundles/`, com
  `--no-verification` obrigatório;
- keyleak-detector desligado porque o modo remoto faz crawl e pode validar BaaS;
- uma URL de projeto ou anon key é contexto, não prova de falha.

### HIBP automatizado

- o runner extrai no máximo 20 e-mails cujo domínio coincide com o alvo, apenas
  das páginas públicas configuradas (`/about`, `/team`, `/contact` etc.);
- consulta somente `/api/v3/breachedaccount/{email}`;
- `domain_search`, pastes e stealer logs são proibidos;
- chave via variável de ambiente e configuração curl por stdin;
- resultado guarda SHA-256 do e-mail, status e nomes de breaches, não o e-mail
  bruto;
- HTTP 404 significa “não observado pela consulta”, não prova universal de
  ausência;
- a automação técnica não resolve a base legal para tratar e-mail; operação
  real deve aguardar essa decisão.

### Saída

- `stack-profile.json` com observações e proveniência;
- relatório de prospecção limitado pela evidência;
- manifesto, outputs brutos protegidos, bundle normalizado e checksums;
- candidatos exigem revisão humana antes de contato comercial.

## Transição comercial

Pipeline 0.5 -> proposta -> contrato + pagamento -> Pipeline 1 integral. O
Pipeline 2 exige ainda SOW, contato de emergência, políticas reais,
contas/matriz quando aplicável e autorização. Uma observação do Pipeline 0.5 ou
um achado no Pipeline 1 nunca concede essa autorização.

## Pipeline 2 — engajamento ativo autorizado

### Gate e isolamento

Antes de ferramenta externa:

1. schemas e semântica das políticas passam;
2. scope e engagement policy concordam em host/porta/taxa/budget;
3. pagamento, SOW e autorização estão válidos;
4. rede exclusiva e chains `DOCKER-USER`/`INPUT` são aplicadas;
5. negativo é bloqueado e positivo alcança o alvo;
6. supply chain e duração da ferramenta são aprovadas.

Decisões são front-loaded; não há aprovação por execução.

### Recon e API

- feroxbuster é o padrão automático, com taxa do alvo, concorrência/recursão
  conservadoras e timeout externo; ffuf fica desativado como alternativa;
- Amass ativo exige autorização, domínio do SOW, datasources explícitos,
  `enum -list`, limite de queries e NS autoritativos aprovados;
- OpenAPI é baixado pelo transporte pinado antes de Schemathesis/Hadrian;
- sem contrato confiável, o runner não inventa um schema.

### Autorização e lógica

- Hadrian v1.0.0 compara identidades e substitui diff de hash;
- dry-run precede mutações; contas, roles e auth são insumos explícitos;
- apenas o relatório em três fases pode gerar `validated`;
- AuthProbe continua desligado e nunca rodaria junto com Hadrian;
- scenario runner aceita só `credit`/`billing` nos módulos
  `business-logic`/`race-conditions`, com número de estado antes/depois.

### BaaS/storage

- Supabase-RLS-Checker é manual/GUI, commit pinado, execução local com
  `pnpm install --frozen-lockfile`, guiada pela matriz;
- anon key sozinha não prova falha de RLS; confirmação exige identidades A/B;
- SupaShield fica desligado: a tag declarada e a versão do pacote divergem e a
  saída JSON esperada pelo adaptador anterior não existe nessa release;
- firepwn é manual/GUI e exige projeto/contas de teste e limpeza;
- keyleak ativo fica desligado até validação empírica do rollback;
- rlsgate fica futuro, estático, sob NDA.

### DAST, injeção e código

- Nuclei somente com templates locais; Interactsh apenas self-hosted;
- ZAP usa imagem por digest e YAML renderizado, autenticação verificada, rotas
  destrutivas excluídas e tempos finitos;
- jwt_tool limita o padrão aos testes baseline, alg:none e RS256→HS256 quando
  aplicável; playbook/bruteforce permanecem desligados;
- Dalfox usa o comando v3 `scan`, somente lista de reflexões e payloads locais;
- sqlmap exige request capturado, parâmetro, safe URL e target customizado;
  time-based/stacked dependem de flags explícitas;
- repositório autorizado pode receber Gitleaks, OSV-Scanner, Semgrep local e
  Trivy.

### Normalização e entrega

Outputs conhecidos viram o modelo canônico; a consolidação deduplica, ranqueia,
liga evidências por SHA-256 e compara somente coberturas compatíveis. Relatório
final não transforma candidato em achado confirmado sem evidência.

## Estado de deploy

A arquitetura e os gates internos estão prontos para onboarding. Das 28
entradas, 20 automáticas foram aprovadas no wrapper exato, possuem runtime por
digest/SBOM e passaram no liveness; oito foram explicitamente restringidas. O
ciclo firewall/proxy/canary passou em Linux root controlado.

Um engagement real ainda nasce bloqueado até receber política, SOW, alvo,
canaries e adaptadores reais. O host escolhido precisa passar por
`deploy-preflight-linux.sh`; templates continuam deliberadamente não
executáveis para impedir que dados de exemplo sejam confundidos com
autorização.

LGPD/base legal ampla e métricas comerciais continuam fora desta iteração por
decisão do projeto; isso não elimina a necessidade de resolvê-las antes da
operação que as exigir.
