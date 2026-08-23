# Ferramentas pesquisadas para os pipelines

Consulta atualizada em 22 de agosto de 2026 usando repositórios e documentação
oficiais. Versões, commits, digests e estado de revisão estão em
`pipeline/tools.lock.json`; componentes estão em `pipeline/sbom.cdx.json`.

## Pipeline 1

- [Subfinder](https://github.com/projectdiscovery/subfinder),
  [OWASP Amass](https://github.com/owasp-amass/amass),
  [httpx](https://github.com/projectdiscovery/httpx),
  [WappalyzerGo](https://github.com/projectdiscovery/wappalyzergo),
  [testssl.sh](https://github.com/testssl/testssl.sh),
  [jsluice](https://github.com/BishopFox/jsluice) e
  [Gitleaks](https://github.com/gitleaks/gitleaks) compõem descoberta,
  fingerprint e análise local.
- [TruffleHog](https://github.com/trufflesecurity/trufflehog) é automático
  apenas em filesystem local, sem verificação.
- [dnsReaper](https://github.com/punk-security/dnsReaper) 2.0.3 passou a
  automático, mas todo resultado continua candidato e nenhum recurso é
  reivindicado.
- [HIBP API v3](https://haveibeenpwned.com/API/V3) é acessada pelo curl oficial
  pinado e somente pelo endpoint breachedaccount.
- [keyleak-detector](https://github.com/Amal-David/keyleak-detector) permanece
  desligado; crawl e validação BaaS não são estáticos.

## Pipeline 2

- [feroxbuster](https://github.com/epi052/feroxbuster) é o content discovery
  padrão; [ffuf](https://github.com/ffuf/ffuf) fica alternativa desativada.
- Amass ativo, [Schemathesis](https://github.com/schemathesis/schemathesis),
  [Nuclei](https://github.com/projectdiscovery/nuclei),
  [OSV-Scanner](https://github.com/google/osv-scanner),
  [Semgrep](https://github.com/semgrep/semgrep) e
  [Trivy](https://github.com/aquasecurity/trivy) entram quando rota e insumos
  existem.
- [Hadrian](https://github.com/praetorian-inc/hadrian) v1.0.0 é a integração
  de BOLA/BFLA. Usa o JSON da release; SARIF posterior à release não foi
  incorporado.
- [ZAP](https://github.com/zaproxy/zaproxy) usa Automation Framework;
  [jwt_tool](https://github.com/ticarpi/jwt_tool),
  [Dalfox](https://github.com/hahwul/dalfox) e
  [sqlmap](https://github.com/sqlmapproject/sqlmap) possuem wrappers limitados.
- [Interactsh](https://github.com/projectdiscovery/interactsh) só é aceito como
  servidor próprio.
- [Supabase-RLS-Checker](https://github.com/sahilahluwalia/Supabase-RLS-Checker)
  é o repositório canônico escolhido e permanece manual/GUI.
  `hand-dot/supabase-rls-checker` é uma extensão diferente e foi rejeitada
  para esta função.
- [SupaShield](https://github.com/Rodrigotari1/supashield) fica desligado: a
  tag v0.3.0 contém pacote 0.2.1 e não oferece o contrato JSON esperado;
  [firepwn](https://github.com/0xbigshaq/firepwn-tool) permanece manual/GUI;
  [rlsgate](https://github.com/GerardoRdz96/rlsgate) fica futuro/NDA.

## Decisões de integração

- repositório/versão descoberta não equivale a aprovação;
- executor externo é Docker-only e exige digest;
- npm usa lockfile e instalação congelada; sem `npx` solto;
- PyPI/npm/Go nunca são instalados sem pin;
- imagens e artefatos precisam de revisão e SBOM;
- keyleak só poderá mudar de estado após teste de rollback com evidência;
- AuthProbe não foi integrado: o projeto homônimo pesquisado é OAuth para MCP,
  não a ferramenta BOLA descrita.

## Estado

O lock contém 28 entradas. As 20 automáticas possuem container por digest,
revisão de adoção `approved`, SBOM e liveness aprovado. Oito ficaram
`restricted` e não são executáveis pelo runner. Não existe entrada `pending`
nem ferramenta automática sem runtime.

A aprovação é estreita ao commit, wrapper e argumentos documentados; não é
garantia geral sobre o projeto terceiro. A revisão completa e os limites estão
em `REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md`.
