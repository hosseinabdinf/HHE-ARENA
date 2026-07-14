# Linux bare metal

Linux bare metal is the intended environment for publication-oriented cycle
measurements. The host must be stable, sufficiently provisioned, and recorded
with its preflight metadata.

## Requirements

- x86_64 Linux with Docker Engine or a native toolchain;
- invariant and nonstop TSC (`constant_tsc`, `nonstop_tsc`, and `rdtscp` where
  available);
- performance CPU governor, documented Turbo/boost policy, and a known CPU
  affinity plan;
- at least 32 GB RAM for heavy HE/HHE experiments, plus scheme-specific margin;
- Docker Engine 24+ if using framework images;
- Python 3.11+ for report generation and Git for source builds.

## Preflight

Run preflight before a benchmark and preserve its output with the result:

```text
hhe preflight --policy strict --core 2
```

Strict mode should pass TSC, governor, Turbo, architecture, and resource
checks. If it fails, fix the host or use `--policy warn` only for diagnostic
work; do not silently publish the resulting cycles.

## Container workflow

```text
docker build -t hhe-arena:tooling .
python tools/validate_framework_contract.py
hhe bench --catalog manifests/images.v2.json \
  --family he_friendly --path ske --cipher Rubato --variant 128L \
  --operation "SKE.Enc()" --iterations 10 --warmup 2 \
  --core 2 --policy strict --preflight \
  --host-id lab-node-01 --output results/rubato-linux
```

For a full approved sweep, use `bench-all`; it creates one untouched JSONL
file per selected job and then generates CSV, LaTeX, PGFPlots, and memory
reports. Compare machines only when their complete host/preflight metadata
matches. Identical containers do not make arbitrary hardware equivalent.

Native Linux validation is currently deferred in this project; this guide
describes the target publication environment, not a claim that it has already
been validated here.
