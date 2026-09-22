#!/usr/bin/env bash
# Shared input checks; never source the operator-provided baseline as code.
host_baseline_check() {
  local input="$1"
  [ -f "$input" ] && [ ! -L "$input" ] || { printf 'baseline must be an ordinary file\n' >&2; return 1; }
  [ "$(realpath -m -- "$input")" = "$(realpath -ms -- "$input")" ] || { printf 'baseline path must not contain symlinks\n' >&2; return 1; }
  [ "$(stat -c '%u:%a' "$input")" = '0:600' ] || { printf 'baseline requires root ownership and mode 0600\n' >&2; return 1; }
}

host_new_evidence() {
  local input="$1"
  [ ! -e "$input" ] && [ ! -L "$input" ] || { printf 'evidence directory must not exist\n' >&2; return 1; }
  [ "$(realpath -m -- "$input")" = "$(realpath -ms -- "$input")" ] || { printf 'evidence path must not contain symlinks\n' >&2; return 1; }
  umask 077
  mkdir -m 700 -- "$input"
}
