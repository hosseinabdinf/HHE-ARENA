# Benchmark orchestrator

The hhe command is the single entry point for prebuilt benchmark jobs.
Framework-specific images add executables and replace
/opt/hhe/benchmarks/manifest.json; the base tooling image intentionally ships
an empty manifest because it contains no HE framework binary.

Each manifest entry is one concrete cipher/language/operation/backend job.
bench requires exactly one selected entry. bench-all executes every selected
entry in stable ID order and writes a separate, untouched JSONL file for each
job. Existing files are rejected unless --append is explicit.

Examples:

    hhe list --manifest manifests/example.json
    hhe bench --manifest manifest.json --cipher hera --language rust --operation encrypt --backend tfhe --iterations 10 --warmup 2 --core 2 --gc not_applicable --policy strict --host-id lab-node-01 --preflight --output results/hera-rust.jsonl
    hhe bench-all --manifest manifest.json --cipher hera --iterations 10 --output results/raw

Selectors accept comma-separated values. --dry-run prints dispatches without
starting binaries. The CLI exports the common HHE_BENCH_* variables for cipher,
language, operation, backend, security profile, implementation, iterations,
warmups, CPU, GC, policy, stable hostname, and output. Manifest env values override those
defaults. Arguments and environment values can use the placeholders documented
by hhe help.

Use hhe preflight --policy strict --core 2 for an explicit check, or add
--preflight to bench or bench-all. Preflight runs once before a sweep. A nonzero
preflight or benchmark exit stops dispatch and is propagated. Run the container
with --cpuset-cpus matching --core; the CLI pins through each language harness
but cannot set the container CPU allocation itself.

The manifest schema is schemas/benchmark-manifest.schema.json. IDs must be
unique. Framework images should validate their manifest, copy only prebuilt
executables into the runtime layer, and keep build tools out.

## Host-side Docker selection

From the repository root, pass the versioned image catalog to make the same
bench and bench-all commands select and prepare a framework image:

    hhe bench --catalog manifests/images.v1.json --cipher AES-128-CBC --operation "AES-128-CBC Encryption" --image-policy auto --dry-run --core 2 --output results/aes-run
    hhe bench-all --catalog manifests/images.v1.json --image-policy build --dry-run --host-id lab-node-01 --output results/matrix-v1
    hhe list --catalog manifests/images.v1.json

Catalog mode expands each image operation into a selectable target. bench still
requires exactly one operation; bench-all orders targets deterministically.
The CLI builds the tooling image once, builds or pulls each selected framework
image once, mounts the host output directory at /work/results, applies
--cpuset-cpus when --core is set, and invokes the embedded manifest-backed
bench command. The framework image remains responsible for executable
dispatch, so adding an image does not add framework-specific Go code.

Image policy meanings are:

- auto: use the local tag if present, otherwise pull an immutable published
  reference, otherwise build from the catalog recipe;
- local: require the local tag;
- build: build the tooling and selected framework images from the repository;
- pull: require pull_ref and retag it to the catalog's stable local tag.

For catalog runs, `--output` names a result bundle directory rather than a
JSONL file. The CLI writes one stable file per selected operation below
`raw/`, then runs the schema-validating report pipeline and writes CSV, LaTeX,
pgfplots, memory, and issue artifacts below `report/`. Reporting starts only
after every selected benchmark exits successfully. A schema or report failure
is a workflow failure, while `--dry-run` prints both Docker stages without
creating the bundle.

Catalog mode forwards one stable `HHE_BENCH_HOSTNAME` to every framework
container so complete-host grouping is not fragmented by Docker's ephemeral
container hostnames. `--host-id` defaults to the invoking system hostname;
`HHE_ARENA_HOST_ID` can set the CLI default. Use an explicit unique label
when the default is generic or shared by cloned machines. OS, architecture,
CPU model, and logical CPU count continue to describe the environment that
actually executes the benchmark.

Matrix-v1 pull references are currently unpublished and recorded as null.
Consequently, explicit pull fails with an actionable error while auto falls
back to the pinned build recipe. Use --project-root when invoking outside the
repository root. --dry-run performs no Docker operation and creates no result
directory; it prints the planned acquisition and container argv only.

The host catalog contract is schemas/image-catalog.schema.json and its first
instance is manifests/images.v1.json. Embedded benchmark manifests remain the
single source of truth inside each image.

## Matrix-v2 hierarchical selection

When a catalog declares `matrix_version: 2`, the CLI automatically loads
`support-matrix.v2.json` from the catalog directory. Use `--matrix` to
override that path, or use `list --matrix FILE` to inspect the declaration
without an image catalog.

The hierarchy is `family -> path -> cipher -> variant -> framework -> backend
-> language`; every selector accepts comma-separated values or `all`.
`--operation` remains available for choosing a concrete measured boundary.
For example:

    hhe list --catalog manifests/images.v2.json --family he_friendly --path ske --cipher Rubato --variant 128L
    hhe bench --catalog manifests/images.v2.json --family aes --path ske --cipher AES --variant AES-128-CBC --framework AES/OpenSSL --operation "AES-128-CBC Encryption" --dry-run --output results/aes
    hhe bench-all --catalog manifests/images.v2.json --family he_friendly --path ske --cipher all --variant all --output results/he-friendly-ske

Matrix listing prints variant ID, family, path, cipher, variant, framework,
backend, language, operation, declared status, and effective execution status.
A choice is `runnable` only when its matrix target is `ready`, its
`implementation_id` exists in the supplied image catalog, and that image
advertises the same operation. `bench` rejects deferred, blocked, and
uncataloged choices. `bench-all` executes only runnable choices, reports how
many matched choices it skipped, and fails when no runnable choice remains.

A framework image may advertise `matrix_targets` when one packaging unit
contains several variant-specific manifest jobs. Each entry maps a matrix
variant ID to the cipher/backend/operation values understood by the embedded
manifest. HHELand uses this mechanism so one compiler-free image exposes two
Rubato, eight Hera, and six Pasta symmetric profiles without duplicating
images.
