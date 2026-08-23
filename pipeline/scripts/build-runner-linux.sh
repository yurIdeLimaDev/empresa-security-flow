#!/usr/bin/env bash
set -euo pipefail

readonly GO_IMAGE='golang:1.26@sha256:45a5f7a810238aabcbad211d70b9ae082022d96f7c7259e94041ad1b933575ac'

pipeline_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cache="$(mktemp -d)"
trap 'rm -rf -- "$cache"' EXIT
mkdir -p "$pipeline_root/bin"

command -v docker >/dev/null || { printf 'Docker is required\n' >&2; exit 1; }

# Dependency retrieval is separated from both builds. go.sum and -mod=readonly
# prevent resolution from silently rewriting the dependency graph.
docker run --rm \
  --mount "type=bind,src=$pipeline_root,dst=/src,readonly" \
  --mount "type=bind,src=$cache,dst=/go/pkg/mod" \
  -w /src "$GO_IMAGE" sh -ceu \
  'GOMODCACHE=/go/pkg/mod GOFLAGS=-mod=readonly go mod download'

for output in pipeline-linux-amd64 pipeline-linux-amd64.rebuild; do
  docker run --rm --network none \
    --mount "type=bind,src=$pipeline_root,dst=/src,readonly" \
    --mount "type=bind,src=$cache,dst=/go/pkg/mod,readonly" \
    --mount "type=bind,src=$pipeline_root/bin,dst=/out" \
    -w /src "$GO_IMAGE" sh -ceu \
    "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=readonly GOMODCACHE=/go/pkg/mod go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o /out/$output ./cmd/pipeline"
done

cmp "$pipeline_root/bin/pipeline-linux-amd64" "$pipeline_root/bin/pipeline-linux-amd64.rebuild"
rm -f -- "$pipeline_root/bin/pipeline-linux-amd64.rebuild"
sha256sum "$pipeline_root/bin/pipeline-linux-amd64" > "$pipeline_root/bin/pipeline-linux-amd64.sha256"
printf 'runner rebuilt reproducibly: '
cat "$pipeline_root/bin/pipeline-linux-amd64.sha256"
