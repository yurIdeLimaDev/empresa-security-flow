#!/bin/sh
set -eu
NETWORK='eng_smoke_20260822_net'
CHAIN='ENG_SMOKE_20260_5470A8D5_EGR'
INPUT_CHAIN='ENG_SMOKE_20260_5470A8D5_INP'
BRIDGE_FILE='/work/case/isolation/bridge.name'
CONTAINERS=$(docker ps -q --filter network="$NETWORK")
if [ -n "$CONTAINERS" ]; then docker kill $CONTAINERS >/dev/null; fi
iptables -w -N "$CHAIN" 2>/dev/null || true
iptables -w -F "$CHAIN"
iptables -w -A "$CHAIN" -j DROP
iptables -w -N "$INPUT_CHAIN" 2>/dev/null || true
iptables -w -F "$INPUT_CHAIN"
iptables -w -A "$INPUT_CHAIN" -j DROP
if [ -f "$BRIDGE_FILE" ]; then
  BRIDGE=$(cat "$BRIDGE_FILE")
  while iptables -w -C DOCKER-USER -i "$BRIDGE" -j "$CHAIN" 2>/dev/null; do iptables -w -D DOCKER-USER -i "$BRIDGE" -j "$CHAIN"; done
  iptables -w -I DOCKER-USER 1 -i "$BRIDGE" -j "$CHAIN"
  while iptables -w -C INPUT -i "$BRIDGE" -j "$INPUT_CHAIN" 2>/dev/null; do iptables -w -D INPUT -i "$BRIDGE" -j "$INPUT_CHAIN"; done
  iptables -w -I INPUT 1 -i "$BRIDGE" -j "$INPUT_CHAIN"
fi
if [ -f '/work/case/runner.pid' ]; then kill -TERM "$(cat '/work/case/runner.pid')" 2>/dev/null || true; fi
