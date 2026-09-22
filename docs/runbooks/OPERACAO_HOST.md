# Runbook — operação do host

1. Prepare `deploy/linux/baseline.env` com dono root e modo 0600; mantenha uma
   raiz exclusiva de casos e outra de evidências.
2. Rode `verify-host-drift-linux.sh` como root. Corrija qualquer divergência.
3. Rode `deploy-preflight-linux.sh <baseline> <engagement-policy> <evidence-dir>`
   antes de cada caso; o diretório deve ser novo.
4. Monitore disco, idade do último backup, idade do preflight e containers
   gerenciados residuais com `monitor-host-linux.sh`.
5. Gere backup `.age` fora da raiz dos casos. A chave privada não fica no host.
6. Restaure em diretório novo e confirme `RESTORE-VERIFIED.json` e hashes antes
   de declarar o backup concluído.
7. Para parada, siga `EMERGENCIA.md`. Não remova manualmente redes ou chains
   antes da preservação.

O ensaio em Docker Linux não substitui o preflight do host dedicado escolhido.

## Aceitação e recuperação vigentes

Baseline deve ser arquivo regular root/0600, sem symlinks no caminho. Raízes
de casos e backup precisam ser separadas, não aninhadas. Evidências exigem
novo destino e diretório pai existente; não reutilizar saída anterior.

Siga o [pacote de aceitação](../../deploy/linux/ACEITACAO.md): copiar backup
cifrado off-host, baixá-lo e restaurar onde a identidade privada é custodiada.
O coletor verifica integridade, idade, pins e recibos; não provisiona o host nem
aprova cliente. Falha de teardown/proxy invalida o fechamento do preflight.
O lote de 13/09 testou esses controles localmente, não num host definitivo.
