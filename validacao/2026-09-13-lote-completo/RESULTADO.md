# Conclusão do lote prioritário local

Data: 13/09/2026. Estado: lote local implementado e validações concluídas.

Os oito pontos de preparação local foram atendidos dentro dos limites abaixo,
sem confundir execução local com homologação externa. Não houve publicação,
chamada a IA real, contratação, pagamento, assinatura ou teste de terceiros.

## Entregas

- Correção da entrega de patches: diff cumulativo derivado do Git no gate
  final, ligado à autorização por SHA-256, sem copiar patches externos.
- Correção da validação de endereços IPv4/IPv6 e tratamento do Turnstile.
- Dependências npm vulneráveis de desenvolvimento atualizadas com pins/lock.
- Matriz executável dos gates, ensaio completo com cliente fictício, POC de
  calibração/replay, métricas comparáveis e estimador por premissas explícitas.
- Coletor de aceitação do host, validação de baseline/evidência, checklist e
  procedimento de restore off-host; sem alegar teste no host definitivo.
  Falha no teardown/proxy do smoke agora retorna erro e impede fechar o
  manifesto de preflight como aprovado; scripts usam LF no checkout.
- CI reforçada, schemas/exemplos, SBOM npm e checagem documental ampliada.
- Estado atual central e limpeza do caminho obsoleto de cópia de patches.

## Evidências

| Verificação | Resultado observado |
| --- | --- |
| Coletor local | 21 verificações passaram, incluindo Go, landing, corpus documental e sintaxe dos oito scripts shell. |
| Motor Go | `go mod verify`, `go vet` e suíte com race detector passaram. |
| Landing React | Instalação limpa com lockfile e scripts de instalação desativados; lint, tipos, build e 17 testes passaram, sem testes ignorados. |
| Dependências npm | Auditoria completa, incluindo desenvolvimento: nove entradas de vulnerabilidade antes da atualização e zero conhecidas após os pins corretivos. Não é garantia contra vulnerabilidades desconhecidas. |
| Scripts Python | Sete testes passaram, incluindo aceite do host, comparação/custo e referências documentais. |
| Documentação | 56 documentos passaram na verificação de UTF-8, títulos e destinos locais; não valida juridicamente os textos. |
| Contratos de dados | 13 schemas compilados e dez exemplos JSON validados. |
| Gates | Seis grupos passaram: configuração de IA, autorização, allowlist, exposição de credenciais, revisão/entrega e não regressão/limites. |
| POC sintética | Dois positivos aceitos e cinco negativos rejeitados; zero aceitações inseguras no conjunto, BEST e revisão humana preservados. |
| Cliente fictício | Dez etapas passaram, até entrega criptografada e restore verificado; scanners, pagamentos, assinaturas e IA reais não foram utilizados. |
| Supply chain | Verificação estrita passou: 20 ferramentas aprovadas e oito restritas, sem promover automaticamente as restritas. SBOM de ferramentas e npm gerados. |
| CI | Três workflows conferidos localmente; sem executar GitHub Actions, push ou deploy. |

Referências canônicas desta entrega:

- [21 verificações locais](checagem-final/checks.json).
- [Auditoria npm](checagem-final/npm-audit.json) e [SBOM npm](checagem-final/landing-sbom.cdx.json).
- [Matriz de gates e testes](aceite-versionado/gate-matrix.json).
- [POC resumida](aceite-versionado/poc-summary.json) e [resultados por cenário](aceite-versionado/poc.json).
- [Ensaio do cliente fictício](aceite-versionado/customer-rehearsal.json).
- [Manifesto dos arquivos](MANIFESTO.json): snapshot local por SHA-256, não assinatura ou commit.

`aceite-versionado/` é a última execução do ensaio. Ela também retestou a
ampliação final do fingerprint do avaliador para fontes Go, módulos e runtime,
feita depois da suíte completa com race detector. `ensaio/` e `ensaio-final/`
registram iterações anteriores; não substituir a referência atual por elas.
Os testes guardam somente resultados saneados; fixtures, patches, identidades
e modelos fictícios ficam em pasta temporária. Novas execuções exigem destinos
novos, sem sobrescrever evidências.

As 12 chamadas registradas pela POC são do transporte simulado. O custo de
0,024 na moeda fictícia `TEST` confere apenas a aritmética do estimador;
não é preço, medição de tokens nem custo previsto de um fornecedor real.
Falhas de preparação encontradas durante o desenvolvimento (destino de backup
aninhado e seleção de schemas) foram corrigidas e retestadas; não há falhas
pendentes nos checks locais citados.

## Mudança operacional importante

A entrega agora exige `approved-security.patch`, gerado na finalização e
vinculado por SHA-256 à autorização. `--patch-root` não é mais aceito.
Autorizações antigas precisam de nova finalização com a aprovação aplicável
em um destino novo; não inserir hashes manualmente para contornar o gate.
O renderer não precisa de Git, pois recebe o patch canônico já finalizado.

## Limpeza e preservação

Removida a função `copyApprovedPatches` e sua chamada no empacotamento, bem como
a obrigação de `--patch-root` na CLI/renderer. O parâmetro antigo é reconhecido
apenas para recusar uso com mensagem de migração; não existe bypass silencioso.
O adaptador manual ativo não foi removido.

Mapeamento: chamada em `delivery.go`, CLI em `cmd/pipeline/main.go`, renderer
em `scripts/render-delivery-linux.sh`, testes de entrega e documentação do
motor/correção. Todos foram atualizados juntos. Planos de negócio/Check antigos
foram classificados como históricos; revisões pinadas e resultados de agosto
não foram reescritos ou apagados. Pastas `legal-flow-sync` e `legal-landing-sync`
contêm metadados Git, não lixo comum; preservadas sem manipular seu histórico.
Exclusões/alterações que já existiam na worktree não pertencem a este lote.

## Referências da auditoria

Regras conservadoras de rede foram confrontadas com o
[registro IPv6 da IANA](https://www.iana.org/assignments/iana-ipv6-special-registry/)
e a [orientação de SSRF da OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html).
Validação DNS antes do fetch reduz riscos, mas não constitui pin de conexão
no Worker: mantenha-o sem bindings de rede privada e com proteção da plataforma.
Advisories npm foram consultados pelo registro e resolvidos sem `audit fix --force`.
Pins diretos: `@cloudflare/vite-plugin` 1.54.8 e `wrangler` 4.131.1.
Overrides corretivos: baseline-browser-mapping 2.11.0, browserslist 4.28.7,
fast-uri 3.1.6, fflate 0.7.5 e js-yaml 4.3.2. Não são dependências novas de produto.

## Limites

Revisão focada nas fronteiras de autorização, contexto/saída, rede, integridade
de arquivos, entrega, dependências e CI. Não é auditoria independente integral
de cada ferramenta de terceiros, garantia jurídica ou certificação de segurança.
POC representa um controle sintético, não capacidade geral de corrigir software.
Provedor IA, host real, OAuth, assinatura, cobrança, emissão fiscal e release
remota permanecem sujeitos às decisões e homologações explicitadas no
[estado atual](../../docs/ESTADO_ATUAL.md).

O pacote de aceitação do host foi implementado e testado com fixtures, não
executado num host Linux root definitivo. Sintaxe shell validada no Windows
não prova firewall, Docker, proxy ou canaries integrados em Linux. O coletor
valida evidências fornecidas pelo operador; não é atestação remota e o restore
real exige baixar e recuperar o backup off-host.

O GraphRAG foi consultado primeiro, mas retornou estado antigo e não cobre a
landing React. Por isso a implementação foi confirmada diretamente nos
arquivos pertinentes. O corpus local passou no build/testes; isso não confirma
reindexação do MCP conectado. Nenhum outro modelo, subagente ou provedor real
foi usado para gerar patches nesta validação.
