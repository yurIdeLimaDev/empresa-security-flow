# Diagramas Mermaid do fluxo vigente

Os diagramas abaixo representam apenas caminhos executáveis e gates que mudam
o resultado. Eles não representam ferramentas restritas como se fossem etapas
automáticas nem criam bypass para autorização, supply chain ou revisão final.

## 1. Jornada completa: geração, verificação e entrega

```mermaid
flowchart LR
    lead[Lead] --> landing[Landing page React]
    landing --> p1[Pipeline 1\nSuperfície pública e baixo impacto]
    p1 --> evidence[Evidência saneada\n+ stack profile]
    evidence --> outreach[Outreach baseado em achado real]
    outreach --> commercial{Contrato e\n pagamento confirmados?}
    commercial -- não --> stopCommercial[Encerrar sem Pipeline 2]
    commercial -- sim --> onboarding[Onboarding técnico\npor schema]
    onboarding --> policy{Escopo, autorização,\npolíticas e isolamento válidos?}
    policy -- não --> blocked[Fail-closed\nsem execução ativa]
    policy -- sim --> p2[Pipeline 2\nMódulos autorizados e roteados]
    p2 --> canonical[Normalização, deduplicação\ne ranking explicável]
    canonical --> report[Bundle canônico\n+ relatório validado]
    report --> remediation[Correção serial\nsobre BEST]
    remediation --> global[Gate global]
    global --> finalReview{Revisão humana final}
    finalReview -- aprovado --> delivery[Entrega vinculada\na commit e hashes]
    finalReview -- alterações --> remediation
    finalReview -- rejeitado --> closed[Encerrar sem entrega]
```

Regras visíveis no diagrama: Pipeline 1 sempre ocorre antes do onboarding; um
perfil de stack vazio não pula o fluxo. Pipeline 2 só começa após contrato,
pagamento e entradas técnicas válidas.

## 2. Execução governada e isolamento por engajamento

```mermaid
flowchart TB
    input[Escopo + engagement policy\n+ tool policy + test matrix] --> schemas{Schemas estritos\ne hashes válidos?}
    schemas -- não --> deny[Recusar execução]
    schemas -- sim --> supply{Lock + SBOM + revisão\n+ digest/runtime aprovados?}
    supply -- não --> deny
    supply -- sim --> host{Linux root +\npreflight do host?}
    host -- não --> deny
    host -- sim --> isolate[Isolamento por engagement\nbridge sem IPv6]
    isolate --> firewall[DOCKER-USER: allowlist\nDNS e IP:porta + DROP]
    firewall --> hostBlock[INPUT bloqueia container→host\nexceto proxy de auditoria]
    hostBlock --> canaries{Canary negativo bloqueado\ne positivo permitido?}
    canaries -- não --> preserve[Preservar evidência\ne abortar]
    canaries -- sim --> executor[Executor Docker-only\nrootfs RO, cap-drop, no-new-privileges\nlimites e timeout]
    executor --> transport[Transporte com host/IP/porta\npinados + proxy para RequestLog]
    transport --> toolRun[ToolRun + hashes de insumos\n+ artefatos 0600]
    toolRun --> teardown[Stop/teardown idempotente\ncom preservação prévia]
```

O proxy produz auditoria; firewall e políticas são as barreiras de segurança.
Falha de supply chain, canary ou preflight não possui caminho de continuação.

## 3. Correção monotônica sobre BEST

```mermaid
flowchart TB
    validated[Bundle validado] --> preflight{Adaptador e gates\npinados e íntegros?}
    preflight -- não --> blocked[Plano bloqueado]
    preflight -- sim --> plan[Plano imutável\nTicket por finding]
    plan --> ticket{Há ticket elegível\ndentro dos limites?}
    ticket -- não --> global[Executar gates globais\nsobre BEST]
    ticket -- sim --> worktree[Worktree isolada\npartindo de BEST atual]
    worktree --> patch[Adaptador manual\npatch no caminho exato]
    patch --> candidate[Commit candidato]
    candidate --> gates[Quality + retest +\nsecurity-full]
    gates --> compare{Escopo, cobertura e\npostura comparáveis a BEST?}
    compare -- não --> reject[Rejeitar tentativa\nregistrar evidência]
    compare -- sim --> regression{Novo achado, severidade\nou rank maior?}
    regression -- sim --> reject
    regression -- não --> promote[Promover candidato\npara BEST]
    reject --> retry{Ainda há tentativa\npermitida?}
    retry -- sim --> worktree
    retry -- não --> ticket
    promote --> ticket
    global --> globalPass{Gates globais\naprovados?}
    globalPass -- não --> retain[Manter BEST\nsem entrega]
    globalPass -- sim --> human{Revisão humana final\ncommit + bundle hash}
    human -- aprovado --> authorization[Autorização de entrega]
    human -- alterações --> ticket
    human -- rejeitado --> retain
```

Somente `promote` altera BEST. Rejeições preservam a versão anterior, e limites
por ticket/caso impedem loop infinito.

## 4. Entrega e recuperação verificáveis

```mermaid
flowchart LR
    authorization[Autorização final\ncommit + hashes] --> package{Plano, estado, aprovação,\nbundle e patches conferem?}
    package -- não --> refuse[Recusar pacote]
    package -- sim --> artifacts[HTML + PDF + bundle permitido\n+ patches + SHA256SUMS]
    artifacts --> manifest[Manifesto determinístico\n+ ZIP determinístico]
    manifest --> encrypt[Criptografia age\nchave pública do destinatário]
    encrypt --> delivery[Entrega criptografada]
    caseData[Raiz exclusiva do caso] --> backup[Backup tar + manifesto\ncriptografado com age]
    backup --> offHost[Identidade privada\nfora do host operacional]
    offHost --> restore[Restauração em diretório novo]
    restore --> verify{Manifesto interno\ne hashes conferem?}
    verify -- sim --> recover[RESTORE-VERIFIED.json]
    verify -- não --> backupFailure[Backup não concluído]
```

Determinismo aplica-se até o ZIP; o ciphertext age varia por desenho
criptográfico. A restauração verificada é requisito para considerar um backup
concluído.
