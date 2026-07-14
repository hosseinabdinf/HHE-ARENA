# Docker benchmark environment

The root image packages the stable shared benchmark tooling. Its final stage is the official Python 3.13 Debian Bookworm slim image, which is glibc-based. C/C++, Go, Rust, and Python use separate builder stages; the runtime contains the linked C schema probe and static utility library, Go and Rust probes, the Python utilities and virtual environment, the report generator, and the benchmark schema. It contains no compiler toolchains.

Framework HE executables remain framework-owned because their TFHE/HElib/SEAL/Lattigo dependencies and build contracts differ. Framework-specific images install those prebuilt binaries and replace the base empty manifest used by the bundled `hhe` orchestrator.

All new framework images follow [framework-image-contract.md](framework-image-contract.md).
The contract fixes the runtime layout, OCI labels, non-root user, result mount,
manifest validation, and failure semantics. The first conforming recipe is
`frameworks/AES/Openssl/docker/Dockerfile`.

## AES/OpenSSL framework image

Build the tooling image, obtain its immutable registry digest for a release
build, and pass that digest both as `TOOLING_IMAGE` and `TOOLING_DIGEST`.
Development builds may use the local tag while recording its image ID:

```sh
docker build --pull=false -t hhe-arena:tooling .
docker build --pull=false \
  -f frameworks/AES/Openssl/docker/Dockerfile \
  --build-arg TOOLING_IMAGE=hhe-arena:tooling \
  --build-arg TOOLING_DIGEST=sha256:LOCAL_TOOLING_IMAGE_ID \
  --build-arg IMAGE_REVISION=GIT_COMMIT \
  --build-arg IMAGE_VERSION=0.1.0-matrix.1 \
  -t hhe-arena:openssl-aes-matrix1 .
docker run --rm hhe-arena:openssl-aes-matrix1 list
```

The Dockerfile-specific ignore file sends only the OpenSSL source and required
contract files. Image construction validates the embedded manifest and fixed
executable but does not execute the AES benchmark.

## Matrix version 1 framework images

| Image tag | Recipe | Embedded manifest | Build-only special dependency |
|---|---|---|---|
| `hhe-arena:openssl-aes-matrix1` | `frameworks/AES/Openssl/docker/Dockerfile` | `openssl-aes.v1.json` | CMake and OpenSSL headers |
| `hhe-arena:hheland-rubato-matrix1` | `frameworks/HHELand/docker/rubato/Dockerfile` | `hheland-rubato.v1.json` | Go 1.24.13 and pinned modules |
| `hhe-arena:elisabeth-clear-matrix1` | `frameworks/elisabeth/docker/Dockerfile` | `elisabeth-clear.v1.json` | Rust 1.85 and FFTW headers |
| `hhe-arena:transistor-symmetric-matrix1` | `frameworks/Transistor/symmetric/docker/Dockerfile` | `transistor-symmetric.v1.json` | GCC/C headers for the precompiled CFFI counter |

Each final image contains only its executable/runtime files, the shared tooling,
and its concrete manifest. All four recipes pass build-time contract validation
and non-root `hhe list`. These checks do not execute cipher benchmarks.

The host CLI maps matrix-v1 selectors to these recipes through
`manifests/images.v1.json`. For example, the following previews the complete
build and run argv without contacting Docker or starting AES:

```sh
hhe bench --catalog manifests/images.v1.json \
  --cipher AES-128-CBC --operation "AES-128-CBC Encryption" \
  --image-policy auto --dry-run --core 2 --host-id lab-node-01 \
  --output results/aes-run
```

See [orchestrator.md](orchestrator.md) for build, local, pull, and auto policy
semantics. Immutable release pull references remain intentionally unset until
the images are published.

The Transistor runtime uses a small standard-Python GF(17)/matrix fallback when
Sage is unavailable. Its image build verifies the fixed-key first clock output
`[15, 6, 12, 15]`, so the compiler-free runtime cannot regress to an
undeclared Sage dependency.

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

Catalog-mode `bench` and `bench-all` treat `--output` as a result bundle. Raw
JSONL is stored under `raw/`; after all selected containers succeed, the CLI
automatically validates those files and generates the complete report set
under `report/` using the framework image's inherited tooling.

The catalog CLI also forwards `--host-id` as `HHE_BENCH_HOSTNAME` to every
framework container. This prevents Docker-generated container names from
splitting otherwise identical host groups. The default is the invoking
machine's hostname, but reproducibility archives should use an explicit stable
and unique host label.
