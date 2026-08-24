#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  printf 'usage: %s <pipeline-linux-amd64> <delivery-job-root>\n' "$0" >&2
  exit 2
fi

readonly IMAGE='gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab'
runner="$(realpath "$1")"
job="$(realpath "$2")"
pipeline_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[ -x "$runner" ] || { printf 'runner is not executable\n' >&2; exit 1; }
[ -r "$runner.sha256" ] || { printf 'runner checksum file is required\n' >&2; exit 1; }
(cd "$(dirname "$runner")" && sha256sum -c "$(basename "$runner").sha256")
for file in delivery-authorization.json plan.json state.json approval.json bundle.json recipient.txt; do
  [ -r "$job/$file" ] || { printf 'missing delivery input: %s\n' "$file" >&2; exit 1; }
done
[ -d "$job/patches" ] || { printf 'missing patches directory\n' >&2; exit 1; }
mkdir -p "$job/output"
[ -z "$(find "$job/output" -mindepth 1 -maxdepth 1 -print -quit)" ] || { printf 'output must be empty\n' >&2; exit 1; }
recipient="$(tr -d '\r\n' < "$job/recipient.txt")"
[[ "$recipient" =~ ^age1[023456789acdefghjklmnpqrstuvwxyz]+$ ]] || { printf 'invalid public age recipient\n' >&2; exit 1; }

docker image inspect "$IMAGE" --format '{{.Id}}' >/dev/null || { printf 'pinned render image must be provisioned before the case\n' >&2; exit 1; }
docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges=true --read-only \
  --pids-limit 64 --memory 256m --cpus 1 --user "$(id -u):$(id -g)" \
  --tmpfs /tmp:rw,noexec,nosuid,nodev,size=32m \
  --mount "type=bind,src=$runner,dst=/pipeline,readonly" \
  --mount "type=bind,src=$pipeline_root,dst=/pipeline-src,readonly" \
  --mount "type=bind,src=$job,dst=/job" -w /pipeline-src "$IMAGE" \
  /pipeline delivery-package \
    --authorization /job/delivery-authorization.json --plan /job/plan.json \
    --state /job/state.json --approval /job/approval.json --bundle /job/bundle.json \
    --patch-root /job/patches --output-dir /job/output --age-recipient "$recipient"
