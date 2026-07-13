# Docker benchmark environment

The root image packages the stable shared benchmark tooling. Its final stage is the official Python 3.13 Debian Bookworm slim image, which is glibc-based. C/C++, Go, Rust, and Python use separate builder stages; the runtime contains the linked C schema probe and static utility library, Go and Rust probes, the Python utilities and virtual environment, the report generator, and the benchmark schema. It contains no compiler toolchains.

Framework HE executables remain framework-owned because their TFHE/HElib/SEAL/Lattigo dependencies and build contracts differ. Framework-specific images install those prebuilt binaries and replace the base empty manifest used by the bundled `hhe` orchestrator.

## Build and smoke validation

```sh
docker build --pull=false -t hhe-arena:tooling .
docker run --rm hhe-arena:tooling probe
docker run --rm hhe-arena:tooling --help
```

All four official base image references include immutable multi-platform SHA-256 index digests. Rust dependencies are locked by `Cargo.lock`; the Python environment is fully version-pinned in `docker/python-requirements.lock`; Go dependencies are pinned by `go.mod` and `go.sum`; the C/C++ utility uses the pinned GCC image.

## Host requirements

Use a native x86-64 Linux host. Docker Desktop/virtualized hosts may be useful for functional checks but must not contribute comparable cycle measurements. Before collecting results:

```sh
docker/preflight.sh --cpu 2 --min-memory-gib 32
```

The read-only preflight checks Docker availability, x86-64 serialized-counter CPU flags, performance governor, disabled Turbo/boost, physical memory, and the requested logical CPU. Strict mode fails on an unmet or unobservable requirement; use `--warn` only for diagnostic runs. Keep the selected CPU and its SMT siblings idle, avoid swapping, and size memory for the chosen framework—32 GiB is only the default floor, not a promise that every HE workload fits.

Pin the container to the same logical CPU requested by the benchmark harness:

```sh
docker run --rm --cpuset-cpus=2 \
  -v "$PWD/results:/work/results" \
  hhe-arena:tooling report results/raw.jsonl --output-dir results/report \
  --schema /opt/hhe/schemas/benchmark.schema.json
```

Use `--cpuset-cpus`, never `--cpus`: a CPU quota does not pin execution and permits migration between logical CPUs. Record the image digest, host metadata, selected CPU, governor, Turbo state, memory, kernel, Docker version, and preflight output with every result archive.
