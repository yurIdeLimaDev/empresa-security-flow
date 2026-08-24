# Runbook — emergência

1. Pare novas execuções e preserve o diretório de evidência em destino novo.
2. Execute `pipeline/scripts/stop-case-linux.sh <case-root> <case-dir>
   <engagement-id> <preservation-dir>` como root no host dedicado.
3. O script só atua em containers com labels gerenciadas e do engagement exato.
4. Registre checksums, logs e estado antes de remover containers e recursos.
5. Revogue credenciais temporárias, rotacione segredos potencialmente expostos
   e confirme que a chave de backup permaneceu fora do host.
6. Não retome até novo preflight e revisão da causa. Se a preservação falhar,
   não execute teardown automático.

Alertas e registros operacionais não devem conter segredo, e-mail, URL privada,
evidência bruta ou identificador do caso.
