# Framework image contract, version 1

A framework image is a runtime extension of the HHE-ARENA tooling image. It
contains prebuilt benchmark programs and a concrete dispatch manifest; users
must not need a compiler or framework source tree. This contract applies to
every cataloged support-matrix version and is validated during every framework
image build.

## Required filesystem layout

| Path | Requirement |
|---|---|
| `/opt/hhe/bin/hhe` | Shared CLI inherited from the tooling image. |
| `/opt/hhe/bin/hhe-tooling` | Image entry point. It propagates command exit status. |
| `/opt/hhe/bin/hhe-preflight` | Read-only host preflight named by every manifest. |
| `/opt/hhe/bin/hhe-report` | Schema validation and report generator. |
| `/opt/hhe/benchmarks/manifest.json` | Non-empty, schema-valid manifest version 1. |
| `/opt/hhe/frameworks/<implementation>/...` | Read-only prebuilt benchmark executables and runtime data. |
| `/opt/hhe/schemas/benchmark.schema.json` | Iteration record schema. |
| `/opt/hhe/schemas/benchmark-manifest.schema.json` | Dispatch manifest schema. |
| `/work/results` | Only standard result location; mount a host directory here. |

Each manifest job names one operation and an absolute executable below
`/opt/hhe/frameworks`. The executable must exist and be executable when the
image is built. Its working directory must also exist. Job IDs are unique and
stable across rebuilds. An executable that unconditionally emits multiple
operations needs an operation selector or separate wrapper before it can be a
production job.

One image may contain multiple executables or parameterized jobs. When several
matrix variants share that image, its host catalog entry uses
`matrix_targets` to map each variant ID to the exact cipher, backend, and
operation selectors in the embedded manifest.

## Runtime and metadata

The runtime layer is glibc-based, contains no compiler, runs as numeric user
and group `65532:65532`, uses `/work` as its working directory, and retains
`ENTRYPOINT ["hhe-tooling"]`. Framework images carry these OCI labels:

- `org.opencontainers.image.title`, `description`, `source`, `revision`, and
  `version`;
- `io.hhe-arena.contract.version=1`;
- `io.hhe-arena.matrix.id`, matching the versioned support-matrix entry;
- `io.hhe-arena.manifest.path=/opt/hhe/benchmarks/manifest.json`;
- `io.hhe-arena.results.path=/work/results`;
- `io.hhe-arena.resource.class`;
- `io.hhe-arena.tooling.digest`, recording the exact tooling image digest.

Development builds may use the local `hhe-arena:tooling` tag. Release builds
must pass an immutable `TOOLING_IMAGE` digest and the same digest as the label
value; a mutable tag is not reproducibility evidence.

Builder and runtime base images use immutable digests. Direct operating-system
build dependencies must also name exact package versions so repository updates
fail the build instead of silently changing its inputs.

## Dispatch and failure semantics

The normal command is `bench` or `bench-all`; users select jobs through the
embedded manifest. `--preflight` runs once before dispatch. Strict preflight,
manifest selection, benchmark process, JSONL validation, or report failure
must return nonzero and stop the workflow. Existing output is rejected unless
the user explicitly requests append. Images do not silently downgrade strict
policy, retry a failed benchmark, or clamp signed memory results.

The container only abstracts build and runtime dependencies. Comparable cycle
measurements still require compatible bare-metal Linux host metadata and CPU
pinning. Catalog dispatch supplies one stable `HHE_BENCH_HOSTNAME` to every
image; framework harnesses must preserve it rather than recording the
ephemeral container hostname. A successful run writes raw JSONL below `/work/results`; report
generation writes a sibling directory chosen by the caller.

## Build-time acceptance

Run the bundled validator after copying the manifest and executables:

```dockerfile
RUN hhe-validate-framework-image --require-benchmarks
```

This checks the manifest schema, unique IDs, fixed preflight path, executable
boundary and permissions, and working directories. The image build must also
record its pinned base digest and perform a CLI `list` check. Benchmark
execution is a later smoke step and is not part of image construction.

When a framework recipe uses the repository root as its Docker build context,
provide a Dockerfile-specific `Dockerfile.dockerignore` allowlist. The root
ignore deliberately excludes `frameworks/`; widening it would send unrelated
frameworks, vendored dependencies, and user build artifacts to every build.
