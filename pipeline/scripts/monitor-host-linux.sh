#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  printf 'usage: %s <baseline.env> <preflight-root> <output.json>\n' "$0" >&2
  exit 2
fi

baseline="$1"
preflight_root="$2"
output="$3"
value() { awk -F= -v key="$1" '$1 == key {sub(/^[^=]*=/, ""); print; found++} END {if (found != 1) exit 1}' "$baseline"; }
case_root="$(value CASE_ROOT)"
backup_root="$(value BACKUP_ROOT)"
min_free_gib="$(value MIN_FREE_GIB)"
max_backup_age="$(value MAX_BACKUP_AGE_HOURS)"
max_preflight_age="$(value MAX_PREFLIGHT_AGE_HOURS)"
[[ "$min_free_gib" =~ ^[0-9]+$ && "$max_backup_age" =~ ^[0-9]+$ && "$max_preflight_age" =~ ^[0-9]+$ ]] || { printf 'invalid numeric baseline\n' >&2; exit 1; }

free_kib="$(df -Pk "$case_root" | awk 'NR==2 {print $4}')"
free_gib="$((free_kib / 1024 / 1024))"
residual_containers="$(docker ps -aq --filter label=empresa-security.managed=true | wc -l | tr -d ' ')"
now_epoch="$(date +%s)"
newest_backup="$(find "$backup_root" -maxdepth 1 -type f -name '*.age.receipt.json' -printf '%T@\n' 2>/dev/null | sort -nr | head -n1 || true)"
newest_preflight="$(find "$preflight_root" -mindepth 2 -maxdepth 2 -type f -name SHA256SUMS -printf '%T@\n' 2>/dev/null | sort -nr | head -n1 || true)"
backup_age=-1
preflight_age=-1
[ -n "$newest_backup" ] && backup_age="$(( (now_epoch - ${newest_backup%.*}) / 3600 ))"
[ -n "$newest_preflight" ] && preflight_age="$(( (now_epoch - ${newest_preflight%.*}) / 3600 ))"
status=passed
[ "$free_gib" -ge "$min_free_gib" ] || status=failed
[ "$backup_age" -ge 0 ] && [ "$backup_age" -le "$max_backup_age" ] || status=failed
[ "$preflight_age" -ge 0 ] && [ "$preflight_age" -le "$max_preflight_age" ] || status=failed
[ "$residual_containers" -eq 0 ] || status=failed

umask 077
printf '{"schema_version":"1.0.0","status":"%s","free_gib":%d,"backup_age_hours":%d,"preflight_age_hours":%d,"residual_container_count":%d}\n' \
  "$status" "$free_gib" "$backup_age" "$preflight_age" "$residual_containers" > "$output"
[ "$status" = passed ] || exit 1
