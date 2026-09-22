# Geração de patches e execução automática

Atualização documental: 13/09/2026. Implementação inicial de geração em 11/09. Camada independente de fornecedor, com checagem local
de possível credencial no pedido completo antes de qualquer transmissão.
Contratação jurídica continua parcialmente preparada, sem provedor ativado.

## Estado e limites

Implementada a camada independente de fornecedor: protocolo JSON versionado,
cliente HTTPS para um gateway futuro, validação/aplicação de propostas,
registro de hashes e execução sequencial até a revisão humana final.

**Não há provedor, modelo, esforço ou credencial selecionados. A geração real
continua desativada.** Os testes usam respostas programadas, arquivos fictícios
e verificadores de laboratório. Eles verificam o motor, não a capacidade de
uma IA corrigir aplicações de clientes. O perfil Python anterior continua
manual por padrão; não foi substituído por um simulador de produção.

O gateway será uma pequena integração que traduz este contrato para a API do
fornecedor escolhido. **Não basta apontar `endpoint` para uma API de modelo
qualquer:** entrada e saída precisam respeitar os schemas deste projeto.
Nenhum SDK, fornecedor ou serviço adicional foi contratado/instalado.

## Funcionamento

1. `remediation-preflight` confere o hash do próprio runner, os gates e a
   configuração da geração. Geração desativada, credencial ausente, envio não
   autorizado ou parâmetros incompletos bloqueiam antes de criar worktrees.
2. `remediation-run` cria ou retoma um plano imutável e reserva o caso por lock.
3. Para cada ticket elegível, prepara worktree sobre BEST e seleciona somente
   os arquivos explicitamente listados em `context_files`.
4. Envia uma única requisição por tentativa. O gateway devolve substituições
   de conteúdo, não comandos. O runner confere identidade da requisição,
   SHA-256 do corpo, commit, fornecedor/modelo/esforço e hash dos arquivos.
5. Aplica a proposta no worktree descartável, gera `generated.patch`, verifica
   o diff antes de criar o commit e chama os verificadores independentes.
6. Só a avaliação canônica pode promover BEST. Rejeições consomem tentativas;
   tentativas seguintes partem de BEST, nunca do candidato rejeitado.
   O próximo pedido inclui um resumo limitado dos motivos/status anteriores,
   produzido pelo motor; logs brutos dos verificadores não são enviados.
7. Todos os tickets precisam terminar nos estados admitidos. O runner reserva
   uma tentativa para a avaliação global. Falta de cobertura, falha dos gates,
   regressão ou limites esgotados impedem concluir a correção.
8. Sucesso termina em `awaiting_final_review`. O comando não cria aprovação,
   não chama `remediation-finalize`, não publica e não entrega ao cliente.

O adaptador manual existente continua disponível pelos comandos individuais.
A automação nova não o usa como fallback silencioso.

## Configuração futura

O arquivo [agent-generation.disabled.example.json](../config/agent-generation.disabled.example.json)
é apenas o objeto `agent`, para incorporar a uma política concreta já
revisada. Não é uma configuração completa de caso e não é executável.

| Campo | Regra |
|---|---|
| `command` | Somente `["builtin:patch-proposal"]`. A implementação é interna ao runner. |
| `executable_sha256` | Hash real do runner compilado. Recalcular após atualização, sem trocar hashes de políticas antigas. |
| `enabled` | `false` até a integração do fornecedor estar pronta e aprovada. |
| `endpoint` | HTTPS de gateway controlado/aprovado, sem usuário/senha, query ou fragmento. |
| `provider`, `model`, `reasoning_effort` | Identificadores explícitos. Sem modelo padrão, descoberta automática ou fallback. |
| `api_key_env` | Nome da variável da credencial. Valor nunca no JSON. Proibida na allowlist de ambiente dos gates. |
| `source_transfer_approved` | Autorização prévia do envio dos arquivos e informações do achado ao gateway/fornecedor. Não é autorização presumida pelo pagamento. |
| `context_files` | Lista exata de 1–32 arquivos Git regulares, UTF-8; sem glob. Seleção definida no onboarding, não pela IA. |
| `max_context_bytes` | Soma dos conteúdos: 1 KiB–1 MiB. Não é promessa de orçamento em tokens/reais. |
| `max_response_bytes` | Corpo de resposta: 1 KiB–2 MiB. |
| `timeout_seconds` | Prazo de cada tentativa, herdado do agente. |

Exemplo de uso após preencher/aprovar a integração e gerar política nova:

```text
pipeline remediation-preflight --config <politica.json> --output <readiness.json> --strict
pipeline remediation-run --config <politica.json>
pipeline remediation-run --config <politica.json> --plan <caso/remediation/plan.json>
```

Não executar o exemplo contra fornecedores/alvos para descobrir credenciais,
configurações ou modelos. Configuração de gateway é entrada operacional
confiável, não URL fornecida por visitante da landing.

## Contrato para o futuro gateway

- [Schema da requisição](../../pipeline/config/schemas/patch-generation-request.schema.json).
- [Schema da resposta](../../pipeline/config/schemas/patch-generation-response.schema.json).
- POST com `Content-Type: application/json`, `Authorization: Bearer ...` e
  `X-Patch-Request-SHA256` (hash dos bytes exatos recebidos).
- Responder HTTP 200 / JSON. O hash não deve ser recalculado sobre JSON
  reformatado. Outros códigos, redirecionamentos e conteúdo extra falham.
- Cada resposta deve repetir `request_id`, `request_sha256`, `base_commit`,
  `provider`, `model` e `reasoning_effort` recebidos, sem substituições.
- `status=patch`: `replacements` contém `{path, before_sha256, content}`.
  `content` é o novo conteúdo integral. O runner produz o diff; não executa
  scripts nem aplica comandos enviados pelo modelo.
- `status=blocked` ou `no_change`: não há correção aceita. A tentativa é
  contabilizada e o fluxo permanece sujeito aos limites.
- Arquivos e descrição do achado são dados não confiáveis, separados das
  instruções fixas. O gateway deve preservar essa separação e não dar ao modelo
  ferramentas, shell, acesso ao host, aprovação ou credenciais de execução.
- A repetição de modelo/effort é uma declaração verificável no protocolo, não
  atestado criptográfico de qual modelo o fornecedor realmente executou. A
  futura integração deve conferir metadados e proibir substituições no servidor.

## Proteções e escolhas desta versão

- Não há processo de IA local com acesso ao worktree: o cliente HTTPS recebe
  somente dados. O provedor não recebe caminhos absolutos, ambiente completo,
  bundle/evidência bruta, testes protegidos, políticas ou autorização de entrega.
- O motor não usa proxy de ambiente, não segue redirects, não faz retry de
  transporte, não armazena cookies e mantém a validação TLS padrão.
- Antes do transporte, inspeciona os valores e chaves do JSON já montado,
  incluindo arquivos, critérios, descrição do achado e feedback. Recusa a
  credencial do próprio gateway e padrões de chaves privadas, tokens,
  atribuições literais de senha/chave, URLs com credenciais e autenticação.
  O erro não inclui o valor encontrado; nenhum conteúdo é mascarado ou
  reescrito, preservando os hashes. O bloqueio segue os limites de tentativas.
- JSON com campos desconhecidos/duplicados, aliases por capitalização,
  conteúdo adicional, truncamento ou excesso de tamanho é recusado.
- Alterações limitadas a arquivos existentes que foram enviados e estão no
  escopo do ticket. Não cria/remove/renomeia arquivos nem muda permissões.
- Testes, caminhos protegidos, arquivos ocultos, instruções `AGENTS.md`/`SKILL.md`,
  dependências, symlinks, submódulos e binários não são contexto editável nesta
  versão. Atualizações de dependências permanecem no fluxo específico anterior.
- Contexto insuficiente deve produzir bloqueio, não ampliação automática do
  acesso. Esses limites são deliberados para a primeira integração.
- Hash e estado do worktree são reconferidos após a resposta. Diff fora do
  escopo/limites é rejeitado antes de commit e gates; BEST não é alterado.
- Sem nova dependência: somente biblioteca padrão Go e funções existentes.
- Gates continuam sendo executores previamente revisados/pinados, não
  verificadores escritos pelo gerador. Devem usar isolamento e testes de
  comportamento adequados à stack; a função de aplicar patches não substitui
  esse trabalho.

**Limites importantes:** lista de arquivos e padrões locais não detectam todos
os dados pessoais ou credenciais, especialmente valores ofuscados, separados,
codificados ou em formatos desconhecidos. Um JWT público ou exemplo literal
também pode bloquear. Isso não confirma vulnerabilidade nem autoriza envio;
selecionar/sanear fontes antes de aprovar o envio continua obrigatório.
O transporte não impõe sozinho firewall, política
de retenção, custo monetário ou isolamento do gateway. Testes e diff não provam
semanticamente que toda alteração é só de segurança ou que todo defeito foi
detectado. A revisão humana final e a validação do perfil continuam necessárias.

## Evidência e retomada

Artefatos operacionais ficam somente na raiz externa do caso:

- `agents/<ticket>/<tentativa>/generated.patch`: patch produzido; contém código
  do cliente, portanto é confidencial e não deve ir para o Git do motor;
- `generation-receipt.json`: identificador, hashes de requisição/resposta/patch,
  base e parâmetros declarados, inclusive quando uma resposta foi rejeitada;
- `agent-run.json`: status, hash do runner, artefatos e log com erros locais;
- `gates/...`: verificadores e seus artefatos; uma rejeição por bundle ausente
  fica em `automation-rejection.json`;
- `automation-summary.json`: último resumo operacional, BEST, tentativas e
  estado; `delivery_authorized` permanece falso neste comando.

`applied=true` no recibo significa apenas aplicação no worktree descartável;
não significa aprovação nos gates, promoção de BEST ou autorização de entrega.

Não persistimos o corpo integral da requisição nem respostas brutas do modelo.
Hash permite correlação/integridade, não prova que a proposta é verdadeira ou
segura. A retenção dos artefatos do caso segue a política aprovada de operação.

`automation.lock` impede outro ciclo e comandos individuais concorrentes.
Saída normal remove o lock. Interrupção abrupta pode deixá-lo: confirmar que
nenhum processo do caso está ativo, preservar os registros e investigar antes
de removê-lo. Não há remoção automática de locks antigos.

O estado `generating` é persistido antes de consultar o gateway. Se houver queda
nessa fase, a retomada bloqueia em vez de repetir uma chamada de resultado/custo
desconhecidos. Isso é recuperação operacional excepcional, não aprovação humana
rotineira de cada ticket. Não editar manualmente o estado para fabricar sucesso.

Falha de transporte/timeout ou resposta interrompida também bloqueiam em
`blocked_provider_uncertain` após contabilizar a tentativa, sem uma segunda
consulta automática. Conferir o registro do gateway antes de decidir a recuperação.

## O que falta quando o fornecedor for escolhido

1. Implementar a tradução do contrato no gateway, fixar versão/dependências,
   revisar o código, conferir modelo/esforço efetivos e configurar credencial.
2. Aprovar envio, saneamento, retenção, acesso e fornecedor; manter chaves fora
   do repositório e dos ambientes de testes do cliente.
3. Definir e impor orçamento monetário/tokens no gateway, além dos limites
   locais de bytes, tempo e tentativas. Tratar rate limit e interrupções sem
   duplicar cobrança silenciosamente.
4. Validar endpoint, TLS, firewall, gateway e runner no Linux dedicado. O ensaio
   desta entrega não repete o ensaio Linux/root nem ativa serviço remoto.
5. Executar correções geradas por IA real em laboratório representativo com
   gates independentes, medir resultado e só então liberar o perfil para clientes.

## Fontes técnicas consultadas

Complemento de 13/09/2026: [kit de avaliação offline](../avaliacao/README.md)
com métricas, comparação e matriz de gates; [estado atual](../../docs/ESTADO_FLUXO.md)
com a mudança do contrato de entrega. Isso não seleciona provedor nem aprova
qualidade de uma IA real.

Resultado e limites do ensaio local:
[validação de 11/09/2026](../../validacao/2026-09-11-geracao-patches/RESULTADO.md).
O reforço posterior, seus testes e as verificações locais do MVP estão em
[fechamento local](../../validacao/2026-09-11-fechamento-local/RESULTADO.md).

A separação entre instruções/dados, validação de saída e privilégios mínimos
segue a [orientação da OWASP sobre prompt injection](https://cheatsheetseries.owasp.org/cheatsheets/LLM_Prompt_Injection_Prevention_Cheat_Sheet.html).
Controles de transporte foram conferidos na [documentação de net/http](https://pkg.go.dev/net/http#Client).
O adaptador manual e suas regras de aplicação continuam seguindo
[git apply](https://git-scm.com/docs/git-apply).

Diagramas de geração, gates e finalização:
[fluxos vigentes](../../docs/DIAGRAMAS_MERMAID.md). O hash da suíte sintética
inclui fontes Go, módulos, runtime, sistema e arquitetura; não compara
avaliadores diferentes como se fossem a mesma referência.
