#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  printf 'usage: %s <baseline.env> <new-evidence-directory>\n' "$0" >&2
  exit 2
fi

baseline="$1"
evidence="$2"
[ "$(uname -s)" = Linux ] || { printf 'Linux is required\n' >&2; exit 1; }
[ "$(id -u)" -eq 0 ] || { printf 'root is required\n' >&2; exit 1; }
[ -r "$baseline" ] || { printf 'baseline is not readable\n' >&2; exit 1; }
if [ -e "$evidence" ] && [ -n "$(find "$evidence" -mindepth 1 -maxdepth 1 -print -quit 2>/dev/null)" ]; then
  printf 'evidence directory must be new or empty\n' >&2
  exit 1
fi
mkdir -p "$evidence"

value() {
  local key="$1" result
  result="$(awk -F= -v key="$key" '$1 == key {sub(/^[^=]*=/, ""); print; found++} END {if (found != 1) exit 1}' "$baseline")" || {
    printf 'missing or duplicate baseline key: %s\n' "$key" >&2
    exit 1
  }
  case "$result" in *$'\n'*|*$'\r'*) printf 'invalid baseline value\n' >&2; exit 1;; esac
  printf '%s' "$result"
}

operator="$(value OPERATOR_USER)"
case_root="$(value CASE_ROOT)"
backup_root="$(value BACKUP_ROOT)"
updates_service="$(value SECURITY_UPDATES_SERVICE)"
reference_image="$(value PYTHON_REFERENCE_IMAGE)"
[[ "$operator" =~ ^[a-z_][a-z0-9_-]*$ ]] || { printf 'invalid operator user\n' >&2; exit 1; }
for path in "$case_root" "$backup_root"; do
  [[ "$path" = /* && "$path" != / && "$path" != /home && "$path" != /srv ]] || { printf 'unsafe dedicated root\n' >&2; exit 1; }
done
[[ "$reference_image" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] || { printf 'reference image is not pinned\n' >&2; exit 1; }

for command in docker iptables realpath sha256sum stat systemctl sshd; do
  command -v "$command" >/dev/null || { printf 'missing executable: %s\n' "$command" >&2; exit 1; }
done
id "$operator" >/dev/null
docker version >/dev/null
iptables -w -S DOCKER-USER >/dev/null
systemctl is-active --quiet docker
systemctl is-enabled --quiet "$updates_service"
docker image inspect "$reference_image" --format '{{.Id}}' >/dev/null

for path in "$case_root" "$backup_root"; do
  [ -d "$path" ] || { printf 'dedicated root missing\n' >&2; exit 1; }
  [ "$(stat -c '%U:%G' "$path")" = "$operator:$operator" ] || { printf 'dedicated root ownership drift\n' >&2; exit 1; }
  [ "$(stat -c '%a' "$path")" = 700 ] || { printf 'dedicated root mode drift\n' >&2; exit 1; }
done

sshd_effective="$(sshd -T)"
grep -qx 'passwordauthentication no' <<<"$sshd_effective" || { printf 'SSH password authentication must be disabled\n' >&2; exit 1; }
grep -Eq '^permitrootlogin (no|prohibit-password)$' <<<"$sshd_effective" || { printf 'SSH root policy drift\n' >&2; exit 1; }
grep -qx 'pubkeyauthentication yes' <<<"$sshd_effective" || { printf 'SSH public key authentication must be enabled\n' >&2; exit 1; }

{
  printf 'schema_version=1.0.0\n'
  printf 'checked_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf 'status=passed\n'
  printf 'docker_user_chain=present\n'
  printf 'docker_service=active\n'
  printf 'security_updates=enabled\n'
  printf 'ssh_key_only=passed\n'
  printf 'dedicated_roots=passed\n'
  printf 'reference_image=present_by_digest\n'
} > "$evidence/host-drift.txt"
sha256sum "$baseline" "$evidence/host-drift.txt" > "$evidence/SHA256SUMS"
printf 'host drift verification passed\n'
