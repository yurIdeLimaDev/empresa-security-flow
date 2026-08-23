#!/bin/sh
set -eu
NETWORK='eng_smoke_20260822_net'
DNS_IP='1.1.1.1'
IMAGE='curlimages/curl:8.15.0@sha256:4026b29997dc7c823b51c164b71e2b51e0fd95cce4601f78202c513d97da2922'
NEGATIVE_URL='http://negative.canary.test:8080/health'
POSITIVE_URL='http://target.canary.test:8080/health'
if ! docker run --rm --network host "$IMAGE" --silent --show-error --fail --connect-timeout 5 --max-time 10 --resolve 'negative.canary.test:8080:172.31.0.10' "$NEGATIVE_URL" >/dev/null; then
  echo 'canary negativo de controle não está acessível pelo host; não é possível provar o bloqueio' >&2
  exit 1
fi
if docker run --rm --network "$NETWORK" --dns "$DNS_IP" --sysctl net.ipv6.conf.all.disable_ipv6=1 "$IMAGE" --silent --show-error --connect-timeout 5 --max-time 10 --resolve 'negative.canary.test:8080:172.31.0.10' "$NEGATIVE_URL" >/dev/null 2>&1; then
  echo 'canary negativo alcançou destino fora da allowlist' >&2
  exit 1
fi
docker run --rm --network "$NETWORK" --dns "$DNS_IP" --sysctl net.ipv6.conf.all.disable_ipv6=1 "$IMAGE" --silent --show-error --fail --connect-timeout 5 --max-time 15 --resolve 'target.canary.test:8080:172.30.0.10' "$POSITIVE_URL" >/dev/null
echo 'canary de egress aprovado'
