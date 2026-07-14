# Docker execution

HHE-ARENA uses a two-layer image model. `hhe-arena:tooling` contains the CLI,
preflight, schema validator, and report generator. Framework images add only
prebuilt benchmark executables plus an embedded manifest.

## Requirements

- Docker Engine 24+ or Docker Desktop with Linux containers;
- access to the repository root as the build context;
- enough Docker CPU/RAM allocation for the selected framework;
- a writable host directory for `/work/results`.

Build and inspect the tooling image:

```text
docker build -t hhe-arena:tooling .
docker run --rm hhe-arena:tooling --help
docker run --rm hhe-arena:tooling probe
```

Framework images are built from catalog recipes. For example:

```text
docker build -f frameworks/AES/P_encoding/docker/Dockerfile \
  -t hhe-arena:p-encoding-aes-matrix2 .
docker run --rm --entrypoint hhe-tooling \
  hhe-arena:p-encoding-aes-matrix2 list
```

The image contract requires an embedded version-1 manifest, executable paths,
`hhe-validate-framework-image --require-benchmarks` during construction, and
UID/GID `65532:65532` at runtime. Validate all catalog entries with:

```text
python tools/validate_framework_contract.py
```

## Running and collecting reports

Use catalog mode from the repository root. `--dry-run` performs no Docker
operation. A real run mounts the output directory and creates `raw/` and
`report/` only after successful dispatch:

```text
hhe bench --catalog manifests/images.v2.json \
  --family he_friendly --path ske --cipher Rubato --variant 128L \
  --operation "SKE.Enc()" --iterations 2 --warmup 1 \
  --policy warn --preflight --output results/rubato-check
```

Use `--cpuset-cpus=N` at the Docker layer when pinning a host CPU, matching the
CLI's `--core N`. Never use `--cpus` as a substitute for affinity. Heavy HE/HHE
jobs are non-smoke and require explicit authorization.
