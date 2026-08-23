# Empresa Security — verificação e correção

Repositório privado do fluxo técnico de verificação inicial, verificação
autorizada e correção de segurança. Landing page, projetos de clientes,
credenciais e resultados brutos não fazem parte deste repositório.

```text
lead -> Pipeline 1: superfície pública e baixo impacto
     -> perfil de stack + evidências
     -> contrato/pagamento/onboarding
     -> Pipeline 2: módulos autorizados
     -> modelo canônico + relatório
     -> correção serial sobre BEST
     -> gates globais -> revisão humana final -> entrega
```

## Organização

- `pipeline/`: runner Go, políticas, schemas, lock, SBOM e runtimes pinados;
- `correcao/`: contrato e documentação do control-plane de correção;
- `docs/`: arquitetura, decisões e modelo canônico;
- `validacao/`: apenas evidências atuais e sanitizadas do próprio motor.

## Verificação local

```powershell
cd pipeline
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go run ./cmd/pipeline supply-chain --lock tools.lock.json --sbom sbom.cdx.json --strict
```

Execução de scanners exige Linux root, Docker, política real aprovada e o
preflight do host:

```bash
cd pipeline
./scripts/build-runner-linux.sh
./scripts/deploy-preflight-linux.sh /caminho/engagement-policy.json /evidencia/nova
```

Os exemplos usam TEST-NET e placeholders deliberadamente inválidos. O runner
falha fechado. Não transforme um exemplo em execução real sem SOW, escopo,
canaries, adaptadores e hashes do engagement.

## Estado

O inventário contém 20 ferramentas automáticas com runtime por digest,
liveness, SBOM e revisão de adoção no wrapper exato. Oito entradas estão
restritas e não são executadas. A revisão é de adoção, não garantia de ausência
de vulnerabilidades no código upstream.

Consulte `docs/ARQUITETURA_IMPLEMENTADA.md` e
`correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md` antes de configurar um caso.
