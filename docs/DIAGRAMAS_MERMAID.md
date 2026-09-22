# Diagramas Mermaid do fluxo vigente

Atualização documental: 13/09/2026. Fontes: [estado atual](ESTADO_FLUXO.md),
[arquitetura](ARQUITETURA_IMPLEMENTADA.md) e
[geração de patches](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).

Os nomes de pipelines são internos, não textos da interface pública. Os
diagramas descrevem regras implementadas e a jornada pretendida; conexões
comerciais não significam integração automática já publicada. IA real, host
definitivo, assinatura e cobrança permanecem pendentes de homologação.

## 1. Jornada e fronteiras de autorização

```mermaid
flowchart TB
    lead["Lead"] --> web["Landing React: prévia pública gratuita"]
    web --> preview["Resultado verdadeiro, parcial e efêmero"]
    preview --> offer["Proposta: preço, prazo e escopo"]
    offer --> commercial{"Contrato, SOW, autorização e pagamento conferidos?"}
    commercial -- não --> stop["Não iniciar serviço contratado"]
    commercial -- sim --> policy1{"Política P1, observação pública,<br/>supply chain e isolamento válidos?"}
    policy1 -- não --> blocked["Bloquear execução"]
    policy1 -- sim --> p1["Pipeline 1 integral obrigatório"]
    p1 --> profile["Evidência + perfil de stack<br/>unknown não pula etapas"]
    profile --> onboarding["Confirmar escopo, contas e módulos"]
    onboarding --> policy2{"Política P2 e autorização<br/>compatíveis com os módulos?"}
    policy2 -- não --> blocked
    policy2 -- sim --> p2["Pipeline 2: apenas módulos autorizados"]
    p2 --> bundle["Normalizar, deduplicar, ranquear<br/>e validar bundle/evidências"]
    bundle --> repair["Correção elegível sobre BEST<br/>com limites e verificadores separados"]
    repair --> global{"Gates globais aprovados?"}
    global -- não --> retain["Manter BEST sem entrega"]
    global -- sim --> review{"Revisão humana técnica final"}
    review -- aprovado --> finalize["Finalizar e vincular patch cumulativo por hash"]
    review -- alterações --> reopen["Reabrir tickets explícitos<br/>somente dentro dos limites"]
    reopen --> repair
    review -- rejeitado --> retain
    finalize --> delivery["Conferir pacote e entregar criptografado"]
```

Pagamento/login não autorizam testes. O motor P1 exige reconhecimento formal
da observação pública; o gate comercial é também procedimento operacional,
não validação criptográfica de contrato pelo runner. Sem achados elegíveis,
não inventar tickets: o plano de correção pode ficar bloqueado. A prévia
pública não inicia o runner nem persiste domínio/resultado como caso contratado.

## 2. Prévia pública e autenticação são caminhos separados

```mermaid
flowchart TB
    form["Domínio informado"] --> human{"Turnstile válido para esta tentativa?"}
    human -- não --> message["Confirme que você é um humano"]
    human -- sim --> waiting["Verificando... + indicador de loading"]
    waiting --> network{"Alvo e limites de rede permitidos?"}
    network -- não --> error["Erro controlado, sem varredura ampliada"]
    network -- sim --> scan["Leitura pública limitada<br/>sem executar JavaScript do alvo"]
    scan --> result["Até oito observações e limitações"]
    scan -- falha --> error
    result --> reset["Limpar token e exigir nova confirmação"]
    error --> reset
    reset --> form
    result --> offer["Oferta posterior, sem cobrança ativa"]
    login["Entrar"] --> enabled{"Credenciais Google ou Apple configuradas?"}
    enabled -- não --> unavailable["Provedor indisponível"]
    enabled -- sim --> oauth["OAuth e sessão Better Auth em D1"]
    oauth --> account["Conta autenticada não autoriza testes"]
```

O caminho OAuth tem implementação local, mas não foi homologado com
provedores reais neste lote. D1/cookies de autenticação não significam
persistência do resultado da prévia. Check e scheduler continuam pausados.

## 3. Execução governada e isolamento

```mermaid
flowchart TB
    input["Escopo, políticas, matriz e hashes"] --> valid{"Schemas e semântica válidos?"}
    valid -- não --> deny["Recusar execução"]
    valid -- sim --> supply{"Lock, SBOM, revisão, digest<br/>e runtime aprovados?"}
    supply -- não --> deny
    supply -- sim --> host{"Linux root e preflight do host?"}
    host -- não --> deny
    host -- sim --> isolate["Bridge exclusiva sem IPv6"]
    isolate --> firewall["DOCKER-USER: DNS e IP:porta permitidos<br/>DROP para o restante"]
    firewall --> inputblock["INPUT bloqueia container para host<br/>exceto proxy de auditoria"]
    inputblock --> canary{"Canary negativo bloqueado<br/>e positivo permitido?"}
    canary -- não --> abort["Preservar evidências e abortar"]
    canary -- sim --> execute["Docker por digest, pull never, rootfs RO<br/>capabilities zeradas, limites e timeout"]
    execute --> logs["Transporte pinado, RequestLog<br/>ToolRun e hashes de artefatos"]
    logs --> cleanup{"Teardown e fechamento do proxy passaram?"}
    cleanup -- não --> fail["Retornar erro<br/>não aprovar manifesto de preflight"]
    cleanup -- sim --> endok["Encerrar execução e preservar evidência"]
```

O proxy audita; firewall e políticas limitam a rede. HTTPS sem MITM registra
o túnel CONNECT, não cada requisição. Ferramentas restritas não têm caminho
automático. Teste histórico em Linux descartável não aprova o host definitivo.

## 4. Geração e correção sobre BEST

```mermaid
flowchart TB
    bundle["Bundle validado com achados elegíveis"] --> preflight{"Configuração e gates íntegros?"}
    preflight -- não --> blocked["Bloquear sem entrega"]
    preflight -- sim --> plan["Plano imutável e lock de estado"]
    plan --> ticket{"Ticket elegível e orçamento disponível?"}
    ticket -- não --> done{"Tickets em estados admitidos<br/>e orçamento global disponível?"}
    done -- não --> blocked
    done -- sim --> global["Gates globais sobre BEST"]
    ticket -- sim --> worktree["Worktree isolada a partir de BEST"]
    worktree --> mode{"Modo explícito do agente"}
    mode -- geração --> context{"Provedor, modelo, effort, autorização,<br/>allowlist e conteúdo sem credencial?"}
    context -- não --> blocked
    context -- sim --> gateway["Gateway futuro: proposta JSON limitada<br/>sem shell ou filesystem para o modelo"]
    mode -- manual --> manual["Adaptador pinado recebe patch externo"]
    gateway --> proposal{"Proposta, fontes, metadados<br/>e diff permitidos?"}
    manual --> proposal
    proposal -- não --> reject["Rejeitar e preservar BEST"]
    proposal -- sim --> commit["CLI cria commit candidato"]
    commit --> gates["Quality, retest e security-full independentes"]
    gates --> compare{"Gates aprovados, alvo resolvido,<br/>cobertura comparável e sem regressão?"}
    compare -- não --> reject
    compare -- sim --> promote["Promover candidato a BEST"]
    reject --> ticket
    promote --> ticket
    global --> approved{"Gates globais aprovados?"}
    approved -- não --> blocked
    approved -- sim --> pending["awaiting_final_review<br/>sem autorizar entrega"]
```

As tentativas consomem limites por ticket/caso, inclusive a reserva global.
Falhas incertas de geração não são repetidas indiscriminadamente. O modo
automático não recorre silenciosamente ao manual. Apenas a avaliação pode
promover BEST. A não regressão é relativa à cobertura e aos testes, não prova
universal de ausência de falhas. O gateway real continua desativado.

## 5. Revisão final e patch canônico

```mermaid
flowchart TB
    best["BEST com gate global aprovado"] --> human{"Decisão humana vinculada<br/>ao commit e bundle SHA-256"}
    human -- rejeitado --> stop["Bloquear entrega"]
    human -- alterações --> reopen["Reabrir somente tickets explícitos<br/>mesmos limites e verificadores"]
    reopen --> rerun["Reexecutar correção e gate global"]
    rerun --> best
    human -- aprovado --> diff["Finalização deriva diff Git<br/>baseline até BEST"]
    diff --> canonical["approved-security.patch em destino novo<br/>hash na autorização"]
    canonical --> verify{"Plano, estado, revisão, bundle,<br/>commit e patch conferem?"}
    verify -- não --> refuse["Recusar empacotamento"]
    verify -- sim --> zip["HTML, PDF, bundle permitido e security.patch<br/>manifesto e ZIP determinísticos"]
    zip --> encrypt["Criptografia age com recipient público"]
    encrypt --> deliver["Pacote cifrado para canal autorizado"]
```

Não existe entrada livre de patches na entrega: `--patch-root` é recusado.
O renderer recebe o patch já finalizado e não precisa de Git. Autorizações
antigas sem `patch_sha256` exigem nova finalização com aprovação aplicável e
destino novo. Hashes provam integridade relativa aos registros do operador,
não identidade humana, assinatura contratual ou recebimento pelo cliente.
O ciphertext age não é determinístico.

## 6. Backup e aceitação do host

```mermaid
flowchart TB
    baseline["Baseline root 0600<br/>raízes separadas e sem symlinks"] --> preflight["Preflight em diretório novo<br/>drift, supply chain, runtimes e isolamento"]
    data["Raiz exclusiva do caso"] --> backup["Backup com manifesto<br/>cifrado para recipient público"]
    backup --> copy["Copiar para armazenamento off-host"]
    copy --> download["Baixar backup para ambiente de recuperação"]
    download --> restore["Restaurar em destino novo<br/>identidade privada fora do runner"]
    restore --> receipts["Recibos de backup e restore<br/>hashes e contagem de entradas"]
    preflight --> collect["Coletor de aceitação<br/>manifestos, hashes, validade e canaries"]
    receipts --> collect
    collect --> pass{"Todas as evidências exigidas conferem?"}
    pass -- não --> blocked["blocked: não aprovar host"]
    pass -- sim --> ready["Relatório saneado passed<br/>ainda não autoriza um cliente"]
```

Copiar e baixar off-host são passos operacionais a executar no ambiente real.
O coletor não provisiona o host e não fornece atestação remota; usa evidências
do operador. Ver [aceitação do host](../deploy/linux/ACEITACAO.md).

## 7. Evidência local versus liberação real

```mermaid
flowchart LR
    source["Código, locks e documentação"] --> checks["Build, testes, race, vet<br/>schemas, SBOM e links"]
    source --> poc["POC: dois positivos e cinco negativos<br/>replay offline, sem IA real"]
    source --> rehearsal["Cliente fictício<br/>contratos e revisão simulados"]
    checks --> local["Lote local validado"]
    poc --> local
    rehearsal --> local
    local --> external["Ainda faltam host real, perfil de cliente,<br/>IA, contratação e release homologados"]
    external --> approval["Aceite operacional do caso<br/>e revisão técnica final da entrega"]
```

O [relatório de 13/09](../validacao/2026-09-13-lote-completo/RESULTADO.md)
registra o que foi executado. Comparações da POC exigem mesmo fingerprint
do avaliador/runtime; custo só é estimado com premissas declaradas.
