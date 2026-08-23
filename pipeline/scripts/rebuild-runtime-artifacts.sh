#!/usr/bin/env bash
set -euo pipefail

# Rebuilds only the four source-derived runtime artifacts. Registry images are
# pulled separately by digest; this script never installs project dependencies
# on the host.

readonly GO124='golang:1.24@sha256:d2d2bc1c84f7e60d7d2438a3836ae7d0c847f4888464e7ec9ba3a1339a1ee804'
readonly GO125='golang:1.25@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80'
readonly DISTROLESS='gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab'
readonly DEBIAN='debian:trixie-slim@sha256:3a39a0592364683e6bab97937b72cad5a8fa6dcbbee90edb3bb48c7f8e94f258'
readonly PYTHON='python:3.13-slim-bookworm@sha256:00faa2debb87529f9f0764e9491d8ba400a3678976616c3bd7cb193745ac20d1'

readonly SUBFINDER_COMMIT='74a4de720d615df0ca9a6635d873bb8bf711df4d'
readonly HADRIAN_COMMIT='9513d9ee8df9148d6db58060ff287e05bed882fb'
readonly JSLUICE_COMMIT='0ddfab153e060a9eeaded4d8669233f7c071e7e4'
readonly SQLMAP_COMMIT='ea8c6bdb63a3b2da1584f328836eb0d28116f7c4'

readonly SUBFINDER_SHA='f874b5463d1d1e338d3f0312b93464d7ada609dbfd3bc180daaeca1fc24a3ceb'
readonly HADRIAN_SHA='20345648122283b4ca984181f64a89f525b37b3038148fdc20851bffab47c8da'
readonly JSLUICE_SHA='49640235e94e96066cbbf13de3f5738331c4048083b134a9d7e68e2931eb8b21'
readonly SQLMAP_SHA='66054f4518c59fefccbdf59ebabc91cafc6d52a62fb6b8542a9cf6e99790e78a'

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
output="$repo_root/pipeline/tool-artifacts/linux-amd64"
work="$(mktemp -d)"
trap 'rm -rf -- "$work"' EXIT

for executable in docker git sha256sum python3; do
  command -v "$executable" >/dev/null || { printf 'missing executable: %s\n' "$executable" >&2; exit 1; }
done

clone_exact() {
  local repository="$1" commit="$2" destination="$3"
  git init -q "$destination"
  git -C "$destination" remote add origin "https://github.com/$repository.git"
  git -C "$destination" fetch -q --depth 1 origin "$commit"
  test "$(git -C "$destination" rev-parse FETCH_HEAD)" = "$commit"
  git -C "$destination" checkout -q --detach FETCH_HEAD
}

build_go() {
  local image="$1" source="$2" package="$3" name="$4" cgo="$5"
  local cache="$work/cache-$name"
  mkdir -p "$cache" "$output"
  docker run --rm \
    --mount "type=bind,src=$source,dst=/src,readonly" \
    --mount "type=bind,src=$cache,dst=/go/pkg/mod" \
    --mount "type=bind,src=$output,dst=/out" \
    -w /src "$image" sh -ceu \
    "GOMODCACHE=/go/pkg/mod go mod download; CGO_ENABLED=$cgo GOOS=linux GOARCH=amd64 GOFLAGS=-mod=readonly GOMODCACHE=/go/pkg/mod go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o /out/$name $package"
}

clone_exact projectdiscovery/subfinder "$SUBFINDER_COMMIT" "$work/subfinder"
clone_exact praetorian-inc/hadrian "$HADRIAN_COMMIT" "$work/hadrian"
clone_exact BishopFox/jsluice "$JSLUICE_COMMIT" "$work/jsluice"
clone_exact sqlmapproject/sqlmap "$SQLMAP_COMMIT" "$work/sqlmap"

build_go "$GO125" "$work/subfinder" ./cmd/subfinder subfinder 0
build_go "$GO124" "$work/hadrian" ./cmd/hadrian hadrian 0
build_go "$GO124" "$work/jsluice" ./cmd/jsluice jsluice 1
python3 "$repo_root/pipeline/scripts/build_sqlmap_artifact.py" \
  --source "$work/sqlmap" --output "$output/sqlmap.pyz"

printf '%s  %s\n' "$SUBFINDER_SHA" "$output/subfinder" | sha256sum -c -
printf '%s  %s\n' "$HADRIAN_SHA" "$output/hadrian" | sha256sum -c -
printf '%s  %s\n' "$JSLUICE_SHA" "$output/jsluice" | sha256sum -c -
printf '%s  %s\n' "$SQLMAP_SHA" "$output/sqlmap.pyz" | sha256sum -c -

docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges=true \
  --read-only --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m,mode=1777 --env HOME=/tmp \
  --mount "type=bind,src=$output/subfinder,dst=/opt/tool/subfinder,readonly" \
  --entrypoint /opt/tool/subfinder "$DISTROLESS" -version >/dev/null
docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges=true \
  --read-only --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m,mode=1777 --env HOME=/tmp \
  --mount "type=bind,src=$output/hadrian,dst=/opt/tool/hadrian,readonly" \
  --entrypoint /opt/tool/hadrian "$DISTROLESS" version >/dev/null
docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges=true \
  --read-only --tmpfs /tmp:rw,noexec,nosuid,nodev,size=256m,mode=1777 --env HOME=/tmp \
  --mount "type=bind,src=$output/jsluice,dst=/opt/tool/jsluice,readonly" \
  --entrypoint /opt/tool/jsluice "$DEBIAN" --help >/dev/null
docker run --rm --network none --cap-drop ALL --security-opt no-new-privileges=true \
  --read-only --tmpfs /tmp:rw,noexec,nosuid,nodev,size=512m,mode=1777 --env HOME=/tmp \
  --mount "type=bind,src=$output/sqlmap.pyz,dst=/opt/tool/sqlmap.pyz,readonly" \
  --entrypoint python3 "$PYTHON" /opt/tool/sqlmap.pyz --version >/dev/null

printf 'runtime artifacts rebuilt and verified\n'
