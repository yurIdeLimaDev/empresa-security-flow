# Revisão de adoção das ferramentas — 2026-08-22

## Decisão

As 28 entradas do catálogo foram revisadas no commit exato registrado no
`tools.lock.json`. A decisão não é uma alegação de auditoria linha a linha nem
uma garantia de ausência de vulnerabilidades. É uma revisão de adoção com
escopo definido: procedência, release/commit, arquivos de dependências,
instalação e atualização, entrada/saída usada pelo adaptador, comportamento de
rede relevante, modos destrutivos, imagem de runtime e teste de liveness.

Vinte ferramentas ficam `approved` apenas para o invólucro e os argumentos
governados do projeto. Oito ficam `restricted` e o executor recusa sua execução.
Não há entrada `pending`.

| Ferramenta | Classe | Decisão | Motivo principal |
|---|---|---|---|
| subfinder | automatic | approved | código e `go.sum` pinados; binário reproduzível; somente enumeração passiva no P1 |
| amass | automatic | approved | imagem exata; P1 força `-passive`; modo ativo exige autorização P2 |
| httpx | automatic | approved | imagem exata; argumentos e budget restritos ao alvo |
| wappalyzer | disabled | restricted | CLI legado não é necessário; detecção fica no `httpx` e nas observações nativas |
| gitleaks | automatic | approved | imagem exata; somente filesystem; relatório redigido |
| trufflehog | automatic | approved | imagem exata; P1 exige `--no-verification` e fonte local |
| testssl | automatic | approved | imagem oficial 3.2 por digest, validada como 3.2.4; alvo único |
| nuclei | automatic | approved | imagem exata; atualização remota desligada; somente templates locais revisados |
| ZAP | automatic | approved | imagem exata; plano local obrigatório e rotas destrutivas proibidas |
| Schemathesis | automatic | approved | imagem exata; schema local pinado e escopo autorizado |
| jwt_tool | automatic | approved | imagem exata; playbook/bruteforce desligados por padrão |
| sqlmap | automatic | approved | fonte 1.10 empacotada deterministicamente; somente endpoint candidato e flags conservadoras |
| Dalfox | automatic | approved | imagem exata; stored XSS e OAST público proibidos |
| OSV-Scanner | automatic | approved | imagem oficial exata; leitura de repositório, sem autofix |
| Trivy | automatic | approved | imagem oficial exata; leitura de repositório, sem autofix |
| Semgrep | automatic | approved | imagem exata; regras locais e métricas desligadas |
| ffuf | disabled | restricted | sobreposição com feroxbuster; evita duplicar descoberta ativa e budget |
| feroxbuster | automatic | approved | imagem oficial por digest validada como 2.13.1; profundidade e taxa limitadas |
| HIBP (`curl`) | automatic | approved | contrato da API pinado; somente `breachedaccount`; e-mail não é persistido em claro |
| Hadrian | automatic | approved | código v1.0.0 e `go.sum` pinados; binário reproduzível; dry-run e identidades A/B obrigatórios |
| jsluice | automatic | approved | código e `go.sum` pinados; binário reproduzível; apenas bundles locais já baixados |
| dnsReaper | automatic | approved | imagem exata; recebe lista curada; saída continua candidata e não reivindica recurso |
| keyleak-detector | disabled | restricted | crawl/validação BaaS ativa e alegação de rollback ainda não demonstrada |
| Supabase-RLS-Checker | manual | restricted | GUI manual; anon key isolada não prova falha e confirmação A/B continua necessária |
| SupaShield | disabled | restricted | tag v0.3.0 contém package 0.2.1 e o adaptador anterior esperava uma saída JSON inexistente |
| firepwn | manual | restricted | GUI e operações graváveis; somente laboratório manual com limpeza explícita |
| rlsgate | disabled | restricted | reservado para revisão estática futura de repositório autorizado sob NDA |
| interactsh-server | infrastructure | restricted | infraestrutura separada, não scanner do runner; endpoint público é proibido |

## Evidência e critérios

- Os 28 commits foram resolvidos nos repositórios declarados. A verificação de
  assinatura do GitHub é registrada como evidência auxiliar, nunca como revisão
  de segurança suficiente.
- Todas as imagens usadas automaticamente são referenciadas por digest. Imagens
  de runtime com binário/arquivo local também verificam o SHA-256 do artefato
  antes do `docker run`.
- Hadrian, jsluice, subfinder e sqlmap possuem receitas reproduzíveis em
  `pipeline/scripts/rebuild-runtime-artifacts.sh`; builds Go usam `go.sum` e
  `-mod=readonly`. O sqlmap é empacotado com nomes, ordem e timestamps estáveis.
- O executor recusa `latest`, ferramenta restrita, revisão/commit divergente,
  referência de revisão alterada, digest inválido e runtime artifact com hash
  divergente.
- A superfície de execução aplica `no-new-privileges`, `cap-drop=ALL`, rootfs
  somente leitura, limites de CPU/memória/PIDs/arquivo, `/tmp` efêmero, timeout,
  remoção forçada após timeout e logs limitados.

## Limite residual

Uma atualização de qualquer commit, digest, artefato, contrato, adaptador ou
referência invalida o lock e exige nova revisão. Ferramenta aprovada pode ter
falhas próprias; o controle reduz a superfície e torna a versão executada
verificável, mas não transforma software terceiro em software confiável por
definição.

## Referências usadas

- Docker: `run`, `--read-only`, `--cap-drop`, `no-new-privileges` e limites de
  recursos: https://docs.docker.com/reference/cli/docker/container/run/ e
  https://docs.docker.com/engine/containers/resource_constraints/
- Docker `DOCKER-USER` e filtragem: https://docs.docker.com/engine/network/firewall-iptables/
- Sigstore/Cosign: https://docs.sigstore.dev/cosign/verifying/verify/
- SLSA Build Track: https://slsa.dev/spec/v1.2/build-track-basics
- OpenSSF Scorecard, dependências pinadas: https://github.com/ossf/scorecard/blob/main/docs/checks/internal/checks.yaml
- CycloneDX SBOM: https://www.cyclonedx.org/guides/OWASP_CycloneDX-Authoritative-Guide-to-SBOM-en.pdf
