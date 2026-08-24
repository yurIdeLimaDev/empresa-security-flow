# Pipeline de geração de leads para revisão de segurança

Estado: implementação e documentação atualizadas em 24 de agosto de 2026.

```mermaid
flowchart LR
    lead[Lead] --> landing[Landing React] --> p1[Pipeline 1]
    p1 --> evidence[Evidência + perfil] --> outreach[Outreach]
    outreach --> commercial{Contrato + pagamento?}
    commercial -- sim --> onboarding[Onboarding + políticas] --> p2[Pipeline 2]
    commercial -- não --> end[Encerrar]
    p2 --> bundle[Bundle validado] --> remediation[Correção sobre BEST]
    remediation --> review[Revisão humana final] --> delivery[Entrega]
```

O fluxo original foi mantido. O Pipeline 1 não é ignorado nem encurtado quando
o perfil fica vazio: ele continua sendo a etapa que coleta evidência pública e
produz o perfil usado na conversão. O Pipeline 2 só começa depois do gate
comercial e do onboarding.

## O que existe

- runner Go em modo `plan` por padrão;
- schemas estritos para lead, escopo, política de engajamento, política de
  ferramenta, matriz de teste e log de execução;
- rede Docker exclusiva por engajamento, egress `DROP` no `DOCKER-USER`,
  bloqueio do acesso ao host no `INPUT`, canaries positivo/negativo, parada de
  emergência e teardown;
- executor Docker por digest, timeout externo, hashes de políticas/insumos,
  SBOM e revisão de código fail-closed;
- proxy HTTP/SOCKS5 para auditoria, sem tratá-lo como barreira de segurança;
- HIBP por e-mail público e dnsReaper automatizados no Pipeline 1;
- normalização, deduplicação, ranking explicável, relatório e comparação de
  engajamentos no modelo canônico;
- correção por ticket em worktree isolado, gates finitos, comparação de postura
  e retenção monotônica da melhor versão;
- uma única revisão humana, no final e vinculada aos hashes entregues.
- primeiro caso real local de correção executado sobre o ZIP CER-Fácil; o
  melhor candidato passou nos gates técnicos e aguarda revisão humana final.

## Organização do repositório

- `pipeline/`: motor compartilhado de verificação, normalização e correção;
- `correcao/`: contratos, configuração de exemplo e documentação da correção;
- `landing-page/`: site React, isolado do motor operacional;
- `validacao/`: somente ensaios, evidências e resultados de validação.
- `pipeline/knowledge/hipporag/`: consulta Graph RAG local exposta ao Codex
  por MCP; o índice e os caches ficam fora do Git.

## Estado real de prontidão

As pendências de implementação do motor foram fechadas:

- as 28 entradas receberam revisão de adoção no commit exato: 20 foram
  aprovadas somente para seus wrappers governados e oito foram restringidas;
- as 20 ferramentas automáticas possuem runtime por digest, contrato de
  liveness e SBOM; nenhuma ferramenta executável ficou sem container;
- as 20 imagens passaram no `runtime-check --pull --strict`; as oito entradas
  manuais/desligadas/infraestrutura são recusadas pelo executor;
- firewall, Docker, proxy e canaries passaram no ensaio integrado Linux root;
- instalação limpa/reprodutível, adulteração de artefato, limites de saída,
  deduplicação com 5.000 bundles, corrida de dados e ciclo de correção foram
  testados;
- o fluxo de correção agora possui `remediation-preflight`: plano algum é
  criado antes de resolver e conferir o SHA-256 do agente e de todos os gates.

Isso deixa o produto **tecnicamente preparado para onboarding, mas não libera
um caso real por default**. Cada deploy ainda precisa substituir os exemplos
TEST-NET/placeholders por política, SOW, escopo e adaptadores daquela stack, e
executar `deploy-preflight-linux.sh` no host Linux escolhido. São entradas do
engagement, não pendências que possam ser preenchidas genericamente sem
inventar autorização ou comandos do cliente.

O caso CER-Fácil permanece candidato à revisão humana final. Seus ToolRuns são
anteriores ao fechamento desta cadeia; precisam ser reexecutados pelo runner
governado atual antes de qualquer autorização de entrega.

As verificações e correções do CER-Fácil ocorreram somente na cópia local do
ZIP fornecido; nenhum alvo remoto do projeto foi testado. A landing React do
serviço foi preservada e não foi publicada.

Documentos principais:

- [especificação vigente](docs/PIPELINE_DE_GERACAO_DE_LEADS.md)
- [arquitetura implementada](docs/ARQUITETURA_IMPLEMENTADA.md)
- [cadeia de ferramentas](docs/CADEIA_DE_FERRAMENTAS.md)
- [modelo canônico](docs/MODELO_CANONICO_E_NORMALIZACAO.md)
- [consulta Graph RAG](pipeline/knowledge/hipporag/README.md)
- [fluxo de correção segura](correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md)
- [fechamento da prontidão operacional](validacao/2026-08-22-fechamento-prontidao-operacional/RESULTADO.md)
- [caso corrigido do CER-Fácil](correcao/casos/2026-08-22-cer-facil/README.md)
- [reteste da correção do CER-Fácil](validacao/2026-08-22-cer-facil-correcao/RESULTADO.md)
- [validação de isolamento/políticas](validacao/2026-08-22-isolamento-politicas/RESULTADO.md)
- [diagramas Mermaid detalhados](docs/DIAGRAMAS_MERMAID.md)

## Estado do primeiro caso

Há um perfil executável ensaiado ponta a ponta em laboratório privado. Os
runbooks estão em `docs/runbooks/`, o perfil em
`correcao/adapters/python-3.13-stdlib/` e o resultado saneado em
`validacao/2026-08-24-primeiro-caso-controlado/`.

Antes de caso real, o Linux dedicado precisa passar drift/preflight como root.
HIBP, OAST público, terceiros, múltiplas stacks e landing page ficaram fora.
