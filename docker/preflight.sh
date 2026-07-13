#!/bin/sh
set -u

policy=strict
cpu="${HHE_CPU:-}"
minimum_memory_gib="${HHE_MIN_MEMORY_GIB:-32}"
failures=0

usage() {
  cat <<'EOF'
Usage: docker/preflight.sh [--warn] [--cpu N] [--min-memory-gib N]

Read-only Linux host checks for reproducible HHE benchmarks. Strict mode exits
nonzero on unmet requirements; --warn reports the same findings and exits zero.
EOF
}

pass() { printf 'PASS: %s\n' "$1"; }
warn() { printf 'WARN: %s\n' "$1" >&2; failures=$((failures + 1)); }

while [ "$#" -gt 0 ]; do
  case "$1" in
    --warn) policy=warn ;;
    --cpu) shift; [ "$#" -gt 0 ] || { usage >&2; exit 64; }; cpu=$1 ;;
    --min-memory-gib) shift; [ "$#" -gt 0 ] || { usage >&2; exit 64; }; minimum_memory_gib=$1 ;;
    --help|-h) usage; exit 0 ;;
    *) usage >&2; exit 64 ;;
  esac
  shift
done

case "$cpu" in ''|*[!0-9]*) [ -z "$cpu" ] || { printf 'CPU must be a nonnegative integer\n' >&2; exit 64; } ;; esac
case "$minimum_memory_gib" in ''|*[!0-9]*) printf 'Memory must be a nonnegative integer\n' >&2; exit 64 ;; esac

[ "$(uname -s 2>/dev/null)" = Linux ] && pass 'Linux host' || warn 'Linux is required for comparable cycle measurements'
[ "$(uname -m 2>/dev/null)" = x86_64 ] && pass 'x86_64 host' || warn 'serialized TSC counters require x86_64'
if grep -Eqi '(microsoft|wsl)' /proc/sys/kernel/osrelease /proc/version 2>/dev/null; then
  warn 'WSL is virtualized; do not publish its cycle measurements'
elif command -v systemd-detect-virt >/dev/null 2>&1 && [ "$(systemd-detect-virt 2>/dev/null || true)" != none ]; then
  warn 'virtualized host detected; use bare metal for comparable cycle measurements'
else
  pass 'no known virtualization detected'
fi

if [ -f /.dockerenv ]; then
  pass 'container runtime active; host Docker daemon check is not applicable'
elif command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  pass 'Docker daemon available'
else
  warn 'Docker daemon is unavailable'
fi

flags=$(awk -F: '/^flags[[:space:]]*:/{print $2; exit}' /proc/cpuinfo 2>/dev/null || true)
for flag in constant_tsc nonstop_tsc rdtscp; do
  case " $flags " in *" $flag "*) pass "CPU flag $flag" ;; *) warn "CPU flag $flag not reported" ;; esac
done

governors=$(cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor 2>/dev/null | sort -u || true)
if [ "$governors" = performance ]; then pass 'CPU governor is performance'; else warn "CPU governor is not uniformly performance (${governors:-unknown})"; fi

turbo=unknown
if [ -r /sys/devices/system/cpu/intel_pstate/no_turbo ]; then
  [ "$(cat /sys/devices/system/cpu/intel_pstate/no_turbo)" = 1 ] && turbo=disabled || turbo=enabled
elif [ -r /sys/devices/system/cpu/cpufreq/boost ]; then
  [ "$(cat /sys/devices/system/cpu/cpufreq/boost)" = 0 ] && turbo=disabled || turbo=enabled
fi
[ "$turbo" = disabled ] && pass 'Turbo/boost disabled' || warn "Turbo/boost is $turbo"

memory_kib=$(awk '/^MemTotal:/{print $2}' /proc/meminfo 2>/dev/null || printf 0)
required_kib=$((minimum_memory_gib * 1024 * 1024))
[ "$memory_kib" -ge "$required_kib" ] && pass "memory >= ${minimum_memory_gib} GiB" || warn "memory below ${minimum_memory_gib} GiB"

if [ -n "$cpu" ]; then
  online=$(cat /sys/devices/system/cpu/online 2>/dev/null || true)
  if [ -d "/sys/devices/system/cpu/cpu$cpu" ]; then pass "CPU $cpu exists (online set: ${online:-unknown})"; else warn "CPU $cpu does not exist"; fi
  siblings=$(cat "/sys/devices/system/cpu/cpu$cpu/topology/thread_siblings_list" 2>/dev/null || true)
  [ -n "$siblings" ] && printf 'INFO: CPU %s thread siblings: %s; keep siblings idle\n' "$cpu" "$siblings"
fi

printf 'INFO: run containers with --cpuset-cpus=%s, never --cpus\n' "${cpu:-N}"
[ "$failures" -eq 0 ] && exit 0
[ "$policy" = warn ] && exit 0
exit 1
