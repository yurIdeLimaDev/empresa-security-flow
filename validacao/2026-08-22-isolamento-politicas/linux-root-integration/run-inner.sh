#!/bin/sh
# Executed only inside the disposable Docker-in-Docker Linux/root validation host.
set -eu
cd /work

ENGAGEMENT_NETWORK=eng_smoke_20260822_net
NEGATIVE_NETWORK=eng_smoke_20260822_negative
TARGET_NAME=empresa-security-controlled-target
NEGATIVE_NAME=empresa-security-controlled-negative
BUSYBOX_IMAGE=busybox:1.37.0@sha256:7a3ebe5bfd1a4a19797d20b0c0bb39d44393e9a03fd852c0865b0f540d868df0

cleanup() {
  docker rm -f "$TARGET_NAME" "$NEGATIVE_NAME" >/dev/null 2>&1 || true
  docker network rm "$ENGAGEMENT_NETWORK" "$NEGATIVE_NETWORK" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker network create --driver bridge --ipv6=false --subnet 172.30.0.0/24 "$ENGAGEMENT_NETWORK" >/dev/null
docker network create --driver bridge --ipv6=false --subnet 172.31.0.0/24 "$NEGATIVE_NETWORK" >/dev/null

docker run --detach --name "$TARGET_NAME" --network "$ENGAGEMENT_NETWORK" --ip 172.30.0.10 "$BUSYBOX_IMAGE" sh -c 'mkdir -p /www && printf "ok\n" >/www/health && exec httpd -f -p 8080 -h /www' >/dev/null
docker run --detach --name "$NEGATIVE_NAME" --network "$NEGATIVE_NETWORK" --ip 172.31.0.10 "$BUSYBOX_IMAGE" sh -c 'mkdir -p /www && printf "control\n" >/www/health && exec httpd -f -p 8080 -h /www' >/dev/null

sleep 1
/work/pipeline-linux isolation-smoke --engagement-policy /work/engagement-policy.json --case-dir /work/case --max-requests 1 --requests-per-second 1
