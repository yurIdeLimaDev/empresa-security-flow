# Controle da cadeia de ferramentas

Estado: fechado e revalidado em 24 de agosto de 2026.

## Regra de execução

O runner não instala ferramentas em tempo de execução e só executa uma entrada
`automatic` quando todos estes vínculos coincidem:

- repositório, tag/versão e commit exatos, nunca `latest`;
- decisão de adoção vinculada ao commit e ao SHA-256 da referência de revisão;
- imagem Docker por `@sha256:<digest>`;
- runtime local, quando necessário, pinado por SHA-256 e montado read-only;
- dependências congeladas no build e componente no SBOM CycloneDX;
- pin da política igual ao lock, contrato de liveness aprovado e executor
  Docker-only.

O executor acrescenta `--pull=never`, rootfs read-only, `cap-drop=ALL`,
`no-new-privileges`, limites de CPU/memória/PIDs/descritores, tmpfs controlado,
timeout externo, remoção forçada após timeout e logs limitados. Artefatos de
saída precisam estar dentro do diretório do caso e recebem SHA-256 no log.

## Inventário vigente

O lock contém 28 entradas:

- 20 `automatic`, com revisão `approved`, container por digest e liveness
  aprovado;
- oito `restricted`: Wappalyzer e ffuf desativados; keyleak-detector,
  SupaShield e rlsgate desativados; Supabase-RLS-Checker e firepwn manuais;
  Interactsh somente como infraestrutura separada;
- zero `pending` e zero ferramenta automática sem runtime por digest.

`restricted` é uma decisão explícita, não uma aprovação para execução. O
executor recusa essas oito entradas. A revisão foi de adoção no wrapper exato,
não auditoria linha a linha nem garantia de ausência de vulnerabilidades. O
registro completo está em
`REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md`.

## Artefatos próprios reproduzíveis

Subfinder, Hadrian, jsluice e sqlmap não foram associados a imagens aleatórias.
O projeto contém artefatos linux/amd64 e uma receita que recompila os commits
pinados com builders por digest, `go.sum`/`-mod=readonly` ou empacotamento Python
determinístico. Dois builds independentes devem produzir os hashes registrados
no lock. O runner linux/amd64 também é construído duas vezes e o script falha
se os hashes divergirem.

## Verificações concluídas

- os 28 commits declarados foram resolvidos; 16 têm verificação de assinatura
  válida no GitHub e 12 são commits não assinados, condição registrada sem ser
  convertida em confiança artificial;
- os 20 runtimes automáticos foram baixados pelo digest exato e passaram no
  comando de liveness sob `network=none` e o perfil endurecido;
- o relatório estrito da cadeia e o relatório de runtime estão aprovados em
  `validacao/2026-08-22-fechamento-prontidao-operacional/`;
- alteração no arquivo de revisão, commit, digest, SBOM, comando, artefato ou
  hash faz a validação falhar fechada.

## Gates operacionais

```powershell
cd pipeline
.\bin\pipeline.exe supply-chain --lock tools.lock.json --sbom sbom.cdx.json --strict --output supply-chain-report.json
.\bin\pipeline.exe runtime-check --lock tools.lock.json --sbom sbom.cdx.json --pull --strict --output runtime-readiness.json
```

No host Linux selecionado para operação, o gate obrigatório é:

```bash
./scripts/build-runner-linux.sh
./scripts/deploy-preflight-linux.sh /caminho/baseline.env /caminho/engagement-policy.json /evidencia/nova
```

Esse último passo não pode ser pré-aprovado genericamente: ele vincula kernel,
Docker, firewall, política, runner e canaries do host real. O ensaio equivalente
já passou em Linux root descartável e controlado.

## Limites da garantia

Digest prova identidade, não autoria; assinatura prova a identidade definida
pela política, não segurança do código; SBOM descreve componentes, não elimina
vulnerabilidades. Por isso a confiança vem da combinação de pins, revisão de
adoção, runtime restrito, evidência e revalidação. Qualquer atualização exige
novo ciclo.

## Dependências do primeiro caso

`filippo.io/age` 1.3.1 é biblioteca do runner, não download operacional. Está
pinada em `go.mod`/`go.sum` e no SBOM CycloneDX. O perfil Python reutiliza uma
imagem já aprovada no lock; o runner confere lock, SBOM, revisão e hash do
artefato antes de cada gate.
