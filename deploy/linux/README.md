# Baseline Linux

[Pacote de aceitação e restore](ACEITACAO.md), preparado em 13/09/2026.
Scripts agora recusam baseline sem root/0600 e evidência reutilizada ou por
symlink. Isso não atesta a configuração de um host que ainda não foi testado.

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
