# Baseline Linux

Material de deploy reproduzível para um host dedicado. Copie
`baseline.env.example` para fora do Git, preencha os valores reais, restrinja a
0600 e execute os scripts como root.

- `verify-host-drift-linux.sh`: SSH por chave, usuário operacional, updates,
  Docker, `DOCKER-USER`, raízes e imagem de referência;
- `deploy-preflight-linux.sh`: drift + política + isolamento, sempre antes do
  caso e com evidência em diretório novo;
- `monitor-host-linux.sh`: apenas contagens/idades saneadas;
- `stop-case-linux.sh`: preservação antes de teardown por labels exatas;
- `render-delivery-linux.sh`: empacotamento offline em imagem por digest.

O baseline não contém credencial. A identidade age de restauração deve ficar
fora do host operacional.
