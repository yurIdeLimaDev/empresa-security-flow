# Ensaio integrado Linux/root

Este diretório define exclusivamente um ensaio descartável e controlado. Os
dois serviços HTTP são BusyBox locais em redes privadas `172.30.0.0/24` e
`172.31.0.0/24`; não há alvo, credencial ou dado de terceiro.

O procedimento executa `pipeline isolation-smoke`, que aplica o firewall
`DOCKER-USER` e `INPUT`, prova o bloqueio do canary negativo, a liberação do
positivo e uma requisição HTTP pelo proxy de auditoria. O diretório `case/`
precisa ser novo e vazio; o runner recusa reutilizá-lo para não misturar a
evidência. O `trap` do script remove os dois containers e as redes internas ao
sair.

Imagens fixadas:

- `docker:29.6-dind@sha256:8d335fc12e365d1fbfab43a20140162cff4a5ff544bb951104e079137300e593`;
- `busybox:1.37.0@sha256:7a3ebe5bfd1a4a19797d20b0c0bb39d44393e9a03fd852c0865b0f540d868df0`;
- `curlimages/curl:8.15.0@sha256:4026b29997dc7c823b51c164b71e2b51e0fd95cce4601f78202c513d97da2922`.
