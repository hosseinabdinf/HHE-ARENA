#!/bin/sh
set -eu

case "${1:-}" in
  bench|bench-all|list|preflight)
    exec hhe "$@"
    ;;
  report)
    shift
    exec hhe-report "$@"
    ;;
  probe)
    hhe-c-schema-smoke >/dev/null
    hhe-go-probe
    hhe-rust-probe
    python -c 'import cffi, jsonschema, psutil; print("hhe-python-probe python dependencies=ok")'
    ;;
  --help|-h|"")
    cat <<'EOF'
HHE-ARENA tooling image

Commands:
  bench [options]
  bench-all [options]
  list [options]
  preflight [options]
  report INPUT.jsonl... --output-dir DIR [report options]
  probe

Framework-specific images add prebuilt benchmark executables and replace the
base empty manifest. Run host preflight before collecting measurements.
EOF
    ;;
  *)
    exec "$@"
    ;;
esac
