#!/bin/sh
set -eu
NETWORK='eng_smoke_20260822_net'
CHAIN='ENG_SMOKE_20260_5470A8D5_EGR'
INPUT_CHAIN='ENG_SMOKE_20260_5470A8D5_INP'
DNS_IP='1.1.1.1'
if ! docker network inspect "$NETWORK" >/dev/null 2>&1; then
  docker network create --driver bridge --ipv6=false --opt com.docker.network.enable_ipv6=false "$NETWORK" >/dev/null
fi
if [ "$(docker network inspect "$NETWORK" --format '{{.EnableIPv6}}')" != "false" ]; then echo 'rede existente possui IPv6 habilitado' >&2; exit 1; fi
BRIDGE=$(docker network inspect "$NETWORK" --format '{{index .Options "com.docker.network.bridge.name"}}')
if [ -z "$BRIDGE" ] || [ "$BRIDGE" = "<no value>" ]; then
  NET_ID=$(docker network inspect "$NETWORK" --format '{{.Id}}')
  BRIDGE=br-$(printf '%s' "$NET_ID" | cut -c1-12)
fi
iptables -w -N "$CHAIN" 2>/dev/null || true
iptables -w -F "$CHAIN"
while iptables -w -C DOCKER-USER -i "$BRIDGE" -j "$CHAIN" 2>/dev/null; do iptables -w -D DOCKER-USER -i "$BRIDGE" -j "$CHAIN"; done
iptables -w -N "$INPUT_CHAIN" 2>/dev/null || true
iptables -w -F "$INPUT_CHAIN"
while iptables -w -C INPUT -i "$BRIDGE" -j "$INPUT_CHAIN" 2>/dev/null; do iptables -w -D INPUT -i "$BRIDGE" -j "$INPUT_CHAIN"; done
iptables -w -A "$INPUT_CHAIN" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -w -A "$INPUT_CHAIN" -j DROP
iptables -w -I INPUT 1 -i "$BRIDGE" -j "$INPUT_CHAIN"
iptables -w -A "$CHAIN" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -w -A "$CHAIN" -d "$DNS_IP" -p udp --dport 53 -j ACCEPT
iptables -w -A "$CHAIN" -d "$DNS_IP" -p tcp --dport 53 -j ACCEPT
iptables -w -A "$CHAIN" -d '172.30.0.10' -p tcp --dport 8080 -j ACCEPT
iptables -w -A "$CHAIN" -j DROP
iptables -w -I DOCKER-USER 1 -i "$BRIDGE" -j "$CHAIN"
printf '%s\n' "$BRIDGE" > '/work/case/isolation/bridge.name'
