#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 4 ]; then
  printf 'usage: %s <case-root> <case-directory> <engagement-id> <new-stop-evidence-directory>\n' "$0" >&2
  exit 2
fi

case_root="$(realpath "$1")"
case_dir="$(realpath "$2")"
engagement_id="$3"
evidence="$4"
[[ "$engagement_id" =~ ^eng_[a-z0-9_]+$ ]] || { printf 'invalid engagement id\n' >&2; exit 1; }
case "$case_dir/" in "$case_root/"*) ;; *) printf 'case directory outside exclusive root\n' >&2; exit 1;; esac
[ "$case_dir" != "$case_root" ] || { printf 'refusing case root as target\n' >&2; exit 1; }
if [ -e "$evidence" ] && [ -n "$(find "$evidence" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]; then
  printf 'stop evidence directory must be new or empty\n' >&2
  exit 1
fi
mkdir -p "$evidence/preserved"

find "$case_dir" -type f \( -name '*.json' -o -name '*.ndjson' -o -name '*.log' -o -name 'SHA256SUMS' \) -print0 |
  sort -z | while IFS= read -r -d '' file; do
    relative="${file#"$case_dir"/}"
    mkdir -p "$evidence/preserved/$(dirname "$relative")"
    cp -- "$file" "$evidence/preserved/$relative"
  done
(cd "$evidence/preserved" && find . -type f -print0 | sort -z | xargs -0 -r sha256sum) > "$evidence/SHA256SUMS.before-stop"

[ -x "$case_dir/isolation/eng-stop.sh" ] && "$case_dir/isolation/eng-stop.sh"
[ -x "$case_dir/isolation/network-teardown.sh" ] && "$case_dir/isolation/network-teardown.sh"
docker ps -aq --filter label=empresa-security.managed=true --filter "label=empresa-security.engagement=$engagement_id" | xargs -r docker rm -f >/dev/null
printf 'stopped_at=%s\nstatus=preserved_then_torn_down\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$evidence/stop-result.txt"
sha256sum "$evidence/SHA256SUMS.before-stop" "$evidence/stop-result.txt" > "$evidence/SHA256SUMS"
