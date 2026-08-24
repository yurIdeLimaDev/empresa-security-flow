# Consulta Graph RAG do fluxo

Esta integração usa HippoRAG 2 como camada local de consulta sobre o fluxo de
verificação e correção. Ela não executa scanners, não altera o alvo e não é a
fonte de verdade: respostas precisam ser conferidas nas fontes recuperadas.

## Escopo indexado

O corpus é construído apenas de `README.md`, `docs/`, código e configuração
textuais de `pipeline/`, documentação/configuração de `correcao/` e as duas
validações finais de prontidão e isolamento. `correcao/casos/`, binários,
caches, artefatos operacionais e a landing page ficam fora por desenho. Isso
mantém o índice do fluxo separado do repositório React e não inclui o caso
CER-Fácil ou dados potencialmente sensíveis de cliente.

Antes de indexar, o construtor recusa arquivos fora de UTF-8, maiores que 512
KiB e padrões básicos de segredo. Isso é uma barreira complementar; Gitleaks
continua sendo o controle de segredo do repositório.

## Controle de fornecedor

HippoRAG é obtido exclusivamente do commit registrado em
`HIPPORAG_UPSTREAM.json`. O upstream não oferece uma assinatura Git verificável
para esse commit; por isso a adoção está limitada a uma camada interna de
consulta. O bootstrap instala somente dependências fixadas por versão e hash
em `requirements.lock` e instala o código HippoRAG sem resolver dependências
nem criar ambiente de build separado.

O índice, cache LLM e ambiente Python ficam em `.runtime/` e `data/`, ambos
ignorados pelo Git. Nunca registre `OPENAI_API_KEY` em arquivo, histórico ou
repositório.

## Primeira execução

Abra PowerShell neste diretório e execute:

```powershell
.\bootstrap.ps1
$env:OPENAI_API_KEY = 'sua-chave-da-api'
.\run.ps1 index
```

`bootstrap.ps1` usa Python 3.12 local, cria um ambiente isolado e executa uma
verificação sem rede do adaptador. `index` usa `gpt-5.6-terra` com
`reasoning_effort=high` para OpenIE e resposta, e
`text-embedding-3-large` para embeddings. O modelo e o esforço são gravados
na identidade do índice; uma configuração diferente não reutiliza o índice.

Uma assinatura ChatGPT/Codex não substitui uma chave da API OpenAI. A chave
permanece somente na sessão atual do terminal acima.

## Consultar na prática

```powershell
.\run.ps1 ask 'Quais gates impedem uma correção pior de ser promovida?'
.\run.ps1 ask 'Quais ferramentas do Pipeline 1 fazem conexões de rede e quais são seus limites?'
.\run.ps1 ask 'Onde está a evidência do teste integrado de firewall, Docker, proxy e canaries?'
```

Faça uma pergunta por vez, específica e verificável. Prefira pedir a relação
entre objetos do projeto: ferramenta → política → evidência; achado → ticket
→ gate → entrega; ou decisão → documento → implementação. A saída sempre
lista as fontes recuperadas. Abra-as e trate a resposta como uma síntese, não
como autorização operacional nem prova isolada.

Se algum arquivo permitido mudar, a consulta falha em vez de usar contexto
antigo. Reconstrua explicitamente:

```powershell
.\run.ps1 rebuild
```

Para conferir apenas o escopo e o fingerprint do corpus, sem chave e sem API:

```powershell
.\run.ps1 corpus
```

## Limites práticos

- A primeira indexação chama a API para extração de entidades/triplas e para
  embeddings; haverá custo e latência.
- O comando não envia `correcao/casos/`, binários ou a landing page.
- HippoRAG 2 não expõe nativamente `reasoning_effort`; o adaptador local o
  injeta explicitamente no endpoint Chat Completions compatível com Terra.
- A API não foi chamada nesta máquina porque não há `OPENAI_API_KEY` no
  ambiente. A preparação, o corpus, o adaptador Terra e uma indexação/retrieval
  HippoRAG com doubles offline foram verificados sem ela.
