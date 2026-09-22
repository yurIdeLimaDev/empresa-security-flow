# Runbook de revisão final e entrega

Estado: 13/09/2026. Não ativa fornecedor, não instala patches no cliente e não
cria uma aprovação humana por conta própria.

## 1. Revisar

A automação para em `awaiting_final_review`. Conferir plano, BEST, bundle,
escopo, cobertura, gates globais e limitações. A revisão técnica humana examina
o diff completo e a evidência, não apenas confirma build verde.

Registrar decisão pelo schema existente, vinculada ao commit e hash do bundle.
`changes_requested` reabre apenas tickets explícitos dentro dos limites;
`rejected` encerra sem entrega. Não editar estado para contornar gates.

## 2. Finalizar

`remediation-finalize` recebe configuração, plano, aprovação e caminho de
saída da autorização. Use diretório exclusivo dessa finalização. Na aprovação,
o motor deriva o diff cumulativo baseline → BEST, gera `approved-security.patch`
ao lado da autorização e preenche `patch_sha256`. Não copia patches de proposta.

Se o patch já existir, não sobrescrevê-lo: investigar o estado e usar destino
novo com a revisão aplicável. Autorizações antigas sem hash de patch precisam
de nova finalização, não preenchimento manual. Não dispensar a aprovação.

## 3. Preparar o renderer

`render-delivery-linux.sh` recebe o runner Linux e uma raiz exclusiva de job.
Entradas nessa raiz:

- `delivery-authorization.json` e `approved-security.patch`, gerados juntos;
- `plan.json`, `state.json`, `approval.json` e `bundle.json` correspondentes;
- `recipient.txt` com recipient público age X25519, nunca identidade privada.

O runner precisa do seu `.sha256`; a imagem por digest deve estar provisionada.
O renderer roda sem rede e não precisa de Git. `output` deve estar vazio.
Não montar dados de outros casos. Divergência entre commit, revisão, estado,
bundle e patch bloqueia o pacote; `--patch-root` é recusado.

O pacote inclui `patches/security.patch` quando não vazio, relatórios e
manifesto. O ZIP é determinístico; a criptografia age não é.

## 4. Entregar e encerrar

Conferir hashes e destinatário, transferir o pacote cifrado pelo canal
contratado e registrar recebimento. Hash local não comprova recebimento/aceite.
O aceite comercial é separado da revisão técnica. Seu modelo de aceite é
mantido fora deste repositório técnico.

Aplicar retenção e encerramento contratados. Backup exige restore verificado:
[aceitação e recuperação](../../deploy/linux/ACEITACAO.md).
Não tratar ensaio fictício como aprovação de entrega real.
