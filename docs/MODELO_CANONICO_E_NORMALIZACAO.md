# Modelo canônico e normalização

Contrato canônico preservado: schema `1.0.0`, implementado em agosto.
Consolidação documental em 13/09/2026; não houve migração do bundle. A nova
vinculação de `patch_sha256` pertence à autorização de entrega, não é mudança
de versão deste modelo. Ver [estado atual](ESTADO_FLUXO.md).

## Resposta direta sobre capacidades

Sim, o normalizador implementado consegue:

- deduplicar por fingerprint canônico;
- ranquear com decomposição do score;
- gerar bundle JSON e relatório Markdown;
- comparar engajamentos com verificação de escopo e cobertura;
- mostrar valor comercial qualitativo sem inventar ROI;
- ligar evidência, execução e insumo por SHA-256.

Ele **não prova automaticamente que todo alerta é verdadeiro**. O hash prova
integridade do artefato, não a correção do scanner. Resultados importados
nascem como `candidate` e exigem revisão, salvo a regra estreita do Hadrian.

## Entidades

- `Asset`: domínio, IP, URL, arquivo ou pacote.
- `Finding`: resultado deduplicável, estado, prioridade e valor qualitativo.
- `Evidence`: caminho/locator, SHA-256, sensibilidade e revisão.
- `ToolRun`: ferramenta, versão, digest do container, hash das políticas,
  parser, entrada e estado de supply chain.
- `RequestLog`: método, URL sanitizada, status, duração e hashes; sem corpo.
- `Engagement`: execução que agrupa os tool runs.
- `Scope`: limites e fingerprint usados na comparação.
- `Exception`: erro, lacuna ou decisão que impede conclusão.

`Report` não é entidade: é uma projeção determinística do bundle.

## Adaptadores

| Ferramenta | Entrada principal | Saída canônica |
|---|---|---|
| Subfinder / Amass | JSONL ou texto | Asset |
| httpx | JSONL | Asset, RequestLog, Evidence |
| Wappalyzer | JSON | Asset + tecnologias |
| Gitleaks / TruffleHog | JSON/JSONL | Finding + Evidence sensível |
| testssl | JSON | Finding + Evidence |
| Nuclei / ZAP | JSON/JSONL ou SARIF | Finding + Evidence |
| Schemathesis | JUnit XML ou SARIF | Finding + Evidence |
| jwt_tool | arquivo preservado | ToolRun + Exception |
| sqlmap | JSON simples ou CSV | Finding + Evidence sensível |
| Dalfox | JSON/JSONL ou SARIF | Finding + Evidence |
| OSV-Scanner / Trivy / Semgrep | JSON ou SARIF | Asset + Finding + Evidence |
| jsluice | JSONL | Asset URL + RequestLog candidato |
| Hadrian v1.0.0 | JSON | RequestLog + Finding + Evidence sensível |
| dnsReaper | JSON pós-processado | Finding candidato |
| HIBP | JSON pseudonimizado | Finding/Evidence sem e-mail bruto |

SARIF 2.1 genérico é detectado antes do adaptador específico. Query e fragmento
são removidos de URLs quando aplicável. Outputs brutos sensíveis permanecem
separados.

## Regra do Hadrian

O adaptador valida o contrato da release v1.0.0, preserva IDs, roles, endpoint,
status e hashes e ignora itens com `is_vulnerability: false`. Só marca
`validated` quando existem `setup_response`, `attack_response` e
`verify_response`. Comparação de identidades sem a terceira fase continua
`candidate`.

## Deduplicação e ranking

Segredos correlacionam por arquivo/linha; dependências por ecossistema,
pacote/versão e identificador; os demais itens por categoria, identificador,
ativo, localização e parâmetro. A mesclagem preserva evidências, ferramentas,
ocorrências e datas.

O ranking vai de 0 a 100 e expõe severidade, confiança, evidência, exposição,
explorabilidade e impacto técnico. Candidatos ficam limitados a 79;
`false_positive` e `fixed_verified` recebem zero. Não é CVSS nem estimativa
financeira.

## Comparação

`compare` produz `new`, `recurring`, `not_observed` e
`fixed_verified`. `not_observed` não significa corrigido. Se escopo ou
cobertura não forem compatíveis, o resultado é explicitamente não comparável.

Na correção, a regra é mais estrita: o scope fingerprint e o conjunto exato de
`tool + version + status + parser + parser_version + tool_digest +
policy_sha256 + supply_chain_status + command_sha256` precisam ser iguais entre
`BEST` e o candidato. Todo ToolRun candidato precisa registrar
`supply_chain_status=runner-gated`; exceção
bloqueante torna o bundle incomparável. O alvo deve desaparecer sob o reteste, o
score ativo precisa cair (ou permanecer igual somente no gate global/alteração
humana solicitada), e nenhum finding novo ou agravado pode aparecer. Essa é a
condição usada para promover um commit; `not_observed` isolado nunca basta.

## Uso

```powershell
cd pipeline
.\bin\pipeline.exe normalize --tool hadrian --input hadrian.json --output hadrian.bundle.json --engagement ENG-001 --scope SCOPE-001
.\bin\pipeline.exe consolidate --input hadrian.bundle.json --input zap.bundle.json --output consolidated.json --report consolidated.md
.\bin\pipeline.exe compare --previous anterior.json --current atual.json --output comparacao.json
```

Nas execuções, outputs conhecidos também são consolidados automaticamente em
`normalized-bundle.json` e `normalized-report.md`.
