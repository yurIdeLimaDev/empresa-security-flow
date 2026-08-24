#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  printf 'usage: %s <baseline.env> <engagement-policy.json> <new-evidence-directory>\n' "$0" >&2
  exit 2
fi

baseline_input="$1"
policy_input="$2"
evidence_input="$3"
pipeline_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
runner="$pipeline_root/bin/pipeline-linux-amd64"

[ "$(uname -s)" = Linux ] || { printf 'Linux is required\n' >&2; exit 1; }
[ "$(id -u)" -eq 0 ] || { printf 'root is required\n' >&2; exit 1; }
[ -x "$runner" ] || { printf 'build the runner with build-runner-linux.sh first\n' >&2; exit 1; }
[ -r "$baseline_input" ] || { printf 'host baseline is not readable\n' >&2; exit 1; }
[ -r "$policy_input" ] || { printf 'engagement policy is not readable\n' >&2; exit 1; }

for executable in docker iptables realpath sha256sum sysctl; do
  command -v "$executable" >/dev/null || { printf 'missing executable: %s\n' "$executable" >&2; exit 1; }
done
docker version >/dev/null
iptables -w -S DOCKER-USER >/dev/null

if [ -e "$evidence_input" ] && [ -n "$(find "$evidence_input" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]; then
  printf 'evidence directory must be new or empty\n' >&2
  exit 1
fi
mkdir -p "$evidence_input/isolation"
"$pipeline_root/scripts/verify-host-drift-linux.sh" "$baseline_input" "$evidence_input/host-drift"
readonly policy="$(realpath "$policy_input")"
readonly evidence="$(realpath "$evidence_input")"

{
  printf 'checked_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf 'kernel=%s\n' "$(uname -srvmo)"
  printf 'docker_server=%s\n' "$(docker version --format '{{.Server.Version}}')"
  printf 'docker_security_options=%s\n' "$(docker info --format '{{json .SecurityOptions}}')"
  printf 'ip_forward=%s\n' "$(sysctl -n net.ipv4.ip_forward)"
  printf 'runner_sha256=%s\n' "$(sha256sum "$runner" | awk '{print $1}')"
  printf 'baseline_sha256=%s\n' "$(sha256sum "$baseline_input" | awk '{print $1}')"
  printf 'policy_sha256=%s\n' "$(sha256sum "$policy" | awk '{print $1}')"
} > "$evidence/host-facts.txt"

(
  cd "$pipeline_root"
  "$runner" supply-chain --lock tools.lock.json --sbom sbom.cdx.json --strict \
    --output "$evidence/supply-chain-report.json"
  "$runner" runtime-check --lock tools.lock.json --sbom sbom.cdx.json --pull --strict \
    --output "$evidence/runtime-readiness.json"
)

"$runner" isolation-smoke --engagement-policy "$policy" \
  --case-dir "$evidence/isolation" --max-requests 1 --requests-per-second 1

(
  cd "$evidence"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
)

printf 'deploy preflight passed; evidence: %s\n' "$evidence"
