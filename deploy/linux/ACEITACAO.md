# Aceitação reproduzível do host

Preparado em 13/09/2026. Este pacote não instala serviços, contrata VPS nem
declara o host aprovado. Os testes locais do coletor usam evidências fictícias.

## Pré-condições

- Host Linux dedicado com root, Docker, iptables/DOCKER-USER, SSH, atualizações,
  usuário operacional e imagem de referência exatamente como no baseline.
- Baseline real fora do Git: arquivo regular, root, 0600; raízes de casos e
  backups distintas, sem symlinks, 0700, propriedade do operador.
- Runner compilado/revisado e seu SHA-256 registrados independentemente.
- Política para canaries e alvo exclusivamente controlados pelo operador;
  não usar a URL de um prospect. Ferramentas permanecem pinadas e governadas.
- Destinos novos, sob diretório privado do operador, para cada tentativa.
  Falha preserva evidência parcial e nunca se torna resultado aprovado.

## Procedimento

1. Conferir baseline e política reais. Executar como root:
   `bash pipeline/scripts/deploy-preflight-linux.sh BASELINE POLICY NOVO-PREFLIGHT`.
   O script verifica drift, supply chain, liveness dos runtimes por digest e
   o teste integrado de isolamento (canary proibido bloqueado, permitido
   acessível, requisição pelo proxy registrada). As imagens vêm do lock;
   provisionamento/pull é explícito, nunca `latest`.
2. Em uma pasta de caso sintético dentro da raiz autorizada, colocar um pequeno
   arquivo não sensível. Gerar `backup-create --case-root RAIZ --case-dir CASO
   --output DESTINO-EXTERNO.age --age-recipient CHAVE-PUBLICA` pelo runner.
   Conferir também `DESTINO-EXTERNO.age.receipt.json`.
3. Copiar o backup criptografado para o armazenamento off-host escolhido.
   Baixar **essa cópia** em máquina separada que guarda a identidade privada.
   Não copiar a identidade para o VPS. Executar `backup-restore --backup
   COPIA.age --identity IDENTIDADE --output-dir NOVO-RESTORE`.
4. Conferir `RESTORE-VERIFIED.json`: hash do backup, contagem e `verified=true`.
   Transportar somente esse recibo de volta para a conferência. Hash do backup
   baixado deve coincidir com o recibo de criação, e a fonte off-host deve ser
   registrada no runbook privado. O coletor não atesta sozinho onde a cópia viveu.
5. Executar em ambiente privado com Python:

```text
python pipeline/scripts/host_acceptance.py --preflight NOVO-PREFLIGHT --backup-receipt CRIACAO.age.receipt.json --restore-receipt RESTORE-VERIFIED.json --runner-sha256 HASH-ESPERADO --output NOVO-ACEITE.json
```

6. O coletor exige manifesto íntegro; relatórios de todas as ferramentas do
   lock, com imagens exatas; SSH/roots/Docker aprovados; canaries/proxy aprovados;
   runner exato; restore vinculado por hash/contagem. Evidências devem ter até
   24 horas, tolerância máxima de cinco minutos no futuro. Ausência, duplicata,
   travessia, symlink, hash divergente ou estado inconclusivo bloqueia.
7. O JSON exporta somente decisões e hashes dos artefatos esperados. Não enviar
   `host-facts.txt`, logs de runtime, baseline, política ou pasta de evidência
   bruta a prospect/CI. O JSON não é atestado remoto nem autorização de cliente.
8. Antes de cada cliente, repetir o preflight; executar o monitor e a revisão
   final conforme [operação](../../docs/runbooks/OPERACAO_HOST.md) e
   [emergência](../../docs/runbooks/EMERGENCIA.md).

## Aceite objetivo

Código de saída zero e `status=passed`, todos os checks verdadeiros, restore
da cópia off-host efetivamente observado e registro privado de quem executou.
Se qualquer condição faltar, não executar cliente. Homologar ainda latência,
limites de recursos e teardown no host real; a aprovação é daquela revisão,
política e data, não permanente. Não alterar configurações silenciosamente
para obter um resultado verde.
