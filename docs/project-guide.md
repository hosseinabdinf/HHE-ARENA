# HHE-ARENA project guide

HHE-ARENA is a Docker-enabled, manifest-driven benchmark arena for hybrid
homomorphic encryption (HHE), transciphering, and reference symmetric
implementations. Its purpose is fair comparison across languages, cipher
families, and HE backends—not the introduction of a new cipher or attack.

## What is implemented

The shared C/C++, Go, Python, and Rust harnesses emit schema `1.0.0` JSONL
records. Each measured record identifies the cipher, operation, backend,
language, implementation, iteration phase, timing, memory deltas, GC mode, and
host metadata. The orchestrator selects one concrete job or a matrix sweep;
the report pipeline validates JSONL before producing CSV, LaTeX, PGFPlots, and
memory artifacts.

Framework images are multi-stage builds: compilers and source trees stay in the
builder layer, while the runtime contains only prebuilt executables and the
common `hhe-tooling` entrypoint. Contract images validate their embedded
manifest during construction and run as UID/GID `65532:65532`.

## Platform requirements

See the separate [platform documentation index](README.md) for Windows/WSL2, Docker, and Linux bare-metal setup requirements.

## Quick start

```text
docker build -t hhe-arena:tooling .
docker run --rm hhe-arena:tooling probe
docker run --rm hhe-arena:tooling --help
python tools/validate_framework_contract.py
```

To inspect the current v2 hierarchy:

```text
hhe list --catalog manifests/images.v2.json --family all --path all --cipher all --variant all
```

Use `--dry-run` before any execution. A real run should select one operation,
use a small iteration count first, and write results to a fresh directory:

```text
hhe bench --catalog manifests/images.v2.json \
  --family he_friendly --path ske --cipher Rubato --variant 128L \
  --operation "SKE.Enc()" --iterations 2 --warmup 1 \
  --policy warn --preflight --output results/rubato-check
```

`bench-all` is for approved sweeps. It builds or selects images, writes one
raw JSONL file per job, and generates reports only after all selected jobs
finish successfully.

## Safety and reproducibility policy

Do not run repository-wide tests or benchmarks by default. Full HE/HHE jobs
can require substantial memory and hours. Start with schema checks, image
contract validation, `hhe list`, and `--dry-run`; request explicit approval
before bounded HE/HHE execution.

Windows Docker Desktop is useful for functional validation, but WSL commonly
cannot verify invariant/nonstop TSC, governor, or Turbo state. Such runs may be
schema-valid and useful for debugging, but their cycle counts are not
publication-grade. Publication comparisons require matching preflight metadata
on compatible bare-metal Linux hosts. Native Linux validation is currently
deferred.

## Adding a framework

Follow [new-framework.md](new-framework.md) and
[adding-a-scheme.md](adding-a-scheme.md). The minimum integration consists of
an active shared-harness call site, preserved `_backup` source, a version-1
manifest, a compiler-free non-root Docker runtime, and catalog/support-matrix
entries. The static contract gate checks these packaging invariants without
executing the benchmark.

## Current scope

AES/FbT remains intentionally unchanged. BtR is dependency-blocked. HE/HHE
workloads for the packaged heavy images are listed and selectable but remain
non-smoke unless explicitly authorized. See [status.md](status.md) and
[support-matrix.md](support-matrix.md) for the authoritative readiness and
variant inventory.
