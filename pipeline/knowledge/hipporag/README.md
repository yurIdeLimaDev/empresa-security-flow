# Conhecimento local do projeto para o Codex

Esta é a camada de recuperação que o Codex consulta antes de ler arquivos do
projeto. Ela usa HippoRAG 2 com embeddings por termos e extração de relações
determinística, ambos locais. Não usa `OPENAI_API_KEY`, não usa a assinatura
do Codex como chave e não envia conteúdo, consultas ou índices pela rede.

O Codex continua sendo responsável por raciocinar e responder. O MCP devolve
somente até seis trechos relevantes e suas fontes, poupando o contexto de uma
varredura desnecessária do repositório.

## Atualização do índice

É automática. Em cada consulta MCP, o servidor calcula o fingerprint do
corpus permitido. Se qualquer arquivo incluído mudou, o índice anterior é
descartado e reconstruído antes de responder. Não há ação manual necessária
no uso normal, desde que o MCP esteja ligado à mesma cópia/corpus consultado.
`run.ps1 rebuild` existe apenas para diagnóstico.

Em 13/09, o MCP conectado retornou versões de agosto, enquanto esta worktree
já continha a revisão local atual. Não assumir que o rebuild de um corpus
atualiza outro checkout. Confirme caminhos/fingerprint do servidor antes de
alterar sua configuração. O corpus local não inclui a landing; confirme-a nos
arquivos relevantes. Nenhuma API de modelo precisa ser ativada para isso.

## Escopo indexado

O corpus inclui `README.md`, `AGENTS.md`, `docs/`, a documentação desta
integração, código e configuração textuais de `pipeline/`,
documentação/configuração de `correcao/` e as validações finais de prontidão e
isolamento. Ele exclui `correcao/casos/`, binários, caches, artefatos
operacionais e a landing page React. Essa separação evita incluir o caso
CER-Fácil ou dados potencialmente sensíveis do cliente.

O construtor recusa arquivos fora de UTF-8, maiores que 512 KiB e padrões
básicos de chaves privadas. Gitleaks permanece o controle de segredo do
repositório.

## Componentes

- `scripts/project_rag.py`: índice local e recuperação HippoRAG.
- `scripts/mcp_server.py`: servidor MCP stdio exposto ao Codex como
  `project_knowledge` com as ferramentas `search` e `status`.
- `data/`: índice e marcador; ignorados pelo Git.
- `HIPPORAG_UPSTREAM.json`: commit fixado do upstream.

## Operação manual de diagnóstico

```powershell
cd C:\Users\yuris\OneDrive\Documents\ChatGPT\Empresa-secutiry\pipeline\knowledge\hipporag

# Apenas na primeira máquina/clonagem: instala HippoRAG no ambiente isolado.
.\bootstrap.ps1

# Cria ou confirma o índice local.
.\run.ps1 index

# Consulta sem o Codex, útil para diagnóstico.
.\run.ps1 search 'Quais gates impedem uma correção pior de ser promovida?'

# Mostra estado/fingerprint; não chama rede.
.\run.ps1 corpus
```

No uso normal, não execute `search`: pergunte normalmente ao Codex. A
configuração MCP em `C:\Users\yuris\.codex\config.toml` inicia o servidor e
o `AGENTS.md` do projeto instrui o agente a pesquisar primeiro o índice.

## Limites honestos

- Este não é o perfil HippoRAG com OpenIE gerado por um LLM remoto. Sem uma
  chave de API, a extração do grafo é lexical e determinística; é adequada para
  localizar documentação, políticas, decisões, ferramentas e relações de
  projeto, mas pode não captar sinônimos ou conceitos implícitos tão bem quanto
  uma indexação com modelo remoto.
- O servidor nunca substitui fontes: se os trechos forem insuficientes, o
  Codex deve abrir somente o arquivo-fonte indicado para confirmar.
- O runtime isolado do HippoRAG continua grande por depender de Torch e
  Transformers. Ele e o índice não entram no Git.
