# Getting started from scratch

This guide takes a new user from cloning HHE-ARENA to producing validated reports. Read the framework-specific README before building an HE implementation: frameworks have different libraries, memory requirements, and licenses.

## 1. Choose the kind of run

- Tooling smoke check: safe and quick; verifies Docker, the CLI, schema, and report generator.
- Native functional check: builds one framework and runs a deliberately small test.
- Benchmark collection: may require tens of GiB and hours. Use native x86-64 Linux, an isolated CPU, and the strict preflight.
- Full sweep: runs many expensive HE jobs. Do not start this as a first test.

Results collected on Windows, WSL, Docker Desktop, a VM, or a laptop in normal use are useful for functionality but should not be published as comparable cycle measurements.

## 2. Get the repository

Install Git, then run:

    git clone https://github.com/hosseinabdinf/HHE-ARENA.git
    cd HHE-ARENA

The repository currently has no Git submodules or Git LFS configuration. Some framework directories include vendored upstream code. Check the license in each framework before redistributing it.

To update later:

    git pull --ff-only

Do not run broad commands such as go test ./..., cargo test --workspace, or ctest at the repository root. Every framework is independent and some HE tests are extremely expensive.

## 3. Prerequisites

Common requirements:

- 64-bit Linux is the supported measurement host. Windows and macOS are suitable for source inspection and limited functional checks.
- Git and enough disk space for compiler caches and framework dependencies.
- Docker 26 or newer for the common tooling image.
- For native builds: C/C++ compiler and CMake where required, Go or Rust where required, and Python 3 where required.
- HHELand and AES/BtR currently declare Go 1.24. The common tooling builders are separately pinned by the root Dockerfile.
- CUDA frameworks require the framework-documented NVIDIA driver, CUDA toolkit, and compatible GPU.
- Ciphers runs should begin with at least 64 GiB RAM; several schemes may need substantially more.

Exact native dependencies are framework-owned. Start from the closest README:

- AES family: frameworks/AES/README.md
- Elisabeth: frameworks/elisabeth/README.md
- Fasta: frameworks/fasta/README.md
- FiLIP: frameworks/filip/ind_ptx/README.md, frameworks/filip/rust/README.md, or frameworks/filip/sorting_hat/README.md
- FRAST: frameworks/FRAST/README.md
- HHE Framework: frameworks/hheFramework/Readme.md
- HHELand: frameworks/HHELand/README.md
- Margrethe: frameworks/margrethe/README.md
- Transistor: frameworks/Transistor/README.md
- Trivium/Kreyvium: frameworks/trkr/README.md
- YuX: frameworks/yux/README.md

## 4. Verify the common tooling

Build the pinned, compiler-free runtime image:

    docker build -t hhe-arena:tooling .
    docker run --rm hhe-arena:tooling probe
    docker run --rm hhe-arena:tooling --help

The base image contains the hhe CLI, schema, report generator, language probes, and Python environment. It intentionally contains no HE framework executable and therefore ships an empty benchmark manifest.

To build the CLI natively, first create the root build directory:

    mkdir -p build
    cd benchmarking-utilities/orchestrator
    go build -o ../../build/hhe .
    cd ../..

On Windows PowerShell, create it with `New-Item -ItemType Directory -Force build`, and name the output `build/hhe.exe`.

## 5. Understand the execution model

A benchmark is not discovered automatically. A manifest entry maps a stable job ID and selectors to one already-built executable. The CLI:

1. selects a concrete manifest entry;
2. optionally runs preflight once;
3. exports the common HHE_BENCH_* environment variables, including the stable
   `HHE_BENCH_HOSTNAME` identity;
4. starts the prebuilt executable;
5. preserves its raw JSONL output.

Inspect the example:

    hhe list --manifest manifests/example.json
    hhe bench --manifest manifests/example.json --cipher hera --dry-run

Use --dry-run before every new manifest entry. The example executable path is illustrative and will not exist until a framework image or native build supplies it.

## 6. Build and register one framework

Follow the framework README and compile a benchmark executable without running a broad suite. Then create a manifest based on manifests/example.json. A minimal entry is:

    {
      "version": 1,
      "preflight": "/opt/hhe/bin/hhe-preflight",
      "benchmarks": [{
        "id": "scheme-language-backend-operation",
        "cipher": "scheme",
        "language": "go",
        "operation": "encrypt",
        "backend": "native",
        "security_profile": "HHE-128",
        "executable": "/absolute/path/to/prebuilt-benchmark",
        "args": ["--operation", "{operation}"]
      }]
    }

Executable paths are interpreted on the machine running hhe. Container manifests normally use /opt/hhe/frameworks/...; native manifests may use absolute host paths. Use forward slashes in JSON paths on Windows or escape every backslash.

List and dry-run the entry:

    hhe list --manifest my-manifest.json
    hhe bench --manifest my-manifest.json --cipher scheme --language go --operation encrypt --backend native --iterations 2 --warmup 1 --policy warn --output results/smoke.jsonl --dry-run

Then remove --dry-run only when the selected executable and workload are understood. The output path must not already exist unless --append is deliberate.

For a concrete HHELand Rubato example, see the Rubato section below.

## 7. HHELand Rubato functional CLI example

HHELand requires Go 1.24. From the repository root, compile the symmetric Rubato test binary without running it:

    cd frameworks/HHELand
    go test -c -o ../../build/hheland-rubato-sym.test ./sym/rubato
    cd ../..

On Windows, use the .exe suffix. Create a manifest entry whose executable points to that file and whose arguments are:

    ["-test.run=^TestNbRubato$", "-test.v=true"]

Use cipher rubato, language go, operation SKE.Enc(), and backend native. Start with two measured iterations and one warm-up. This exercises the symmetric cipher only. The HHE/transciphering Rubato test is a different, expensive job and should have a separate manifest entry.

## 8. Preflight and CPU isolation

On native Linux, run:

    docker/preflight.sh --cpu 2 --min-memory-gib 32

For a diagnostic functional run, add --warn. For measurements, fix every warning and use strict policy. Run containers with --cpuset-cpus=2, never --cpus, and pass --core 2 to hhe so the container allocation and language harness agree.

Record the CPU model, logical CPU, SMT sibling, kernel, governor, Turbo/boost state, memory, Docker/image digest, compiler/runtime versions, framework revision, and manifest with every result archive.

## 9. Run one job or a sweep

One exact job:

    hhe bench --manifest manifest.json --cipher rubato --language go --operation "SKE.Enc()" --backend native --iterations 10 --warmup 2 --core 2 --gc enabled --policy strict --host-id lab-node-01 --preflight --output results/rubato.jsonl

A manifest-driven sweep:

    hhe bench-all --manifest manifest.json --cipher rubato --iterations 10 --core 2 --policy strict --host-id lab-node-01 --preflight --output results/raw

bench requires exactly one match. bench-all sorts by stable job ID and creates one JSONL file per job. Do not combine unlike GC modes, hosts, security profiles, or backends in later comparisons.

`--host-id` defaults to the invoking system hostname. Set it explicitly to a
stable unique label when running through Docker or when machines use generic
or cloned hostnames; every language harness records the same value.

## 10. Validate and generate reports

Using the Docker image:

    docker run --rm -v "$PWD/results:/work/results" hhe-arena:tooling report results/raw.jsonl --output-dir results/report --schema /opt/hhe/schemas/benchmark.schema.json

Multiple JSONL paths are allowed. The report pipeline validates every line, excludes warm-ups, separates all grouping dimensions, flags failed/incomplete records, and creates:

- report.csv
- report.tex
- report.pgfplots.tex
- memory.csv
- memory.tex
- memory.pgfplots.tex
- issues.json

An invalid record causes a nonzero report exit after diagnostic artifacts are written. Raw JSONL is the source of truth; never replace it with only a summary table.

## 11. Troubleshooting

- No benchmark matches: run hhe list with the same manifest and remove selectors one at a time.
- bench found several matches: add selectors or use bench-all.
- Output exists: choose a new path; use --append only when records truly belong to the same raw file.
- Executable not found: remember that paths are resolved where hhe runs, not where the manifest was authored.
- Strict preflight fails: correct the host condition; do not silently switch a publication run to warn.
- cycles is null: serialized counters are unavailable or the host is unsupported. time_ns can still be valid for a functional run.
- Docker base manifest is empty: this is intentional. Supply a framework-specific image or mount a manifest and its prebuilt executable.
- Build exhausts memory or appears stuck: stop it and follow the framework README; HE key generation and bootstrapping can be very costly.

## 12. Documentation map

- docs/orchestrator.md: complete CLI and manifest behavior
- docs/docker.md: reproducible container and host requirements
- docs/benchmark_schema.md: exact JSONL timing and memory semantics
- schemas/benchmark.schema.json: record contract
- schemas/benchmark-manifest.schema.json: manifest contract
- benchmarking-utilities/report/README.md: statistics and report details
- docs/adding-a-scheme.md: contributor checklist for a new scheme
- docs/status.md: framework/cipher migration and CLI-readiness audit
- security_level_normalization.md: HHE-128 normalization

Before a large run, commit or archive the exact manifest and confirm a small job produces schema-valid JSONL.
