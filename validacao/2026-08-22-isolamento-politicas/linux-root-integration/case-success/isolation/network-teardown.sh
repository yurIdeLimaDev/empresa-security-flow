#!/bin/sh
set -eu
NETWORK='eng_smoke_20260822_net'
CHAIN='ENG_SMOKE_20260_5470A8D5_EGR'
INPUT_CHAIN='ENG_SMOKE_20260_5470A8D5_INP'
BRIDGE_FILE='/work/case/isolation/bridge.name'
if [ -f "$BRIDGE_FILE" ]; then
  BRIDGE=$(cat "$BRIDGE_FILE")
  while iptables -w -C DOCKER-USER -i "$BRIDGE" -j "$CHAIN" 2>/dev/null; do iptables -w -D DOCKER-USER -i "$BRIDGE" -j "$CHAIN"; done
  while iptables -w -C INPUT -i "$BRIDGE" -j "$INPUT_CHAIN" 2>/dev/null; do iptables -w -D INPUT -i "$BRIDGE" -j "$INPUT_CHAIN"; done
fi
iptables -w -F "$CHAIN" 2>/dev/null || true
iptables -w -X "$CHAIN" 2>/dev/null || true
iptables -w -F "$INPUT_CHAIN" 2>/dev/null || true
iptables -w -X "$INPUT_CHAIN" 2>/dev/null || true
docker network rm "$NETWORK" >/dev/null 2>&1 || true
