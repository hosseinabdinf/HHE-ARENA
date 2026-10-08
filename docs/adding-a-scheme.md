# Adding a new scheme or framework

This guide defines the integration contract for adding an HHE scheme, cipher, backend, language port, or implementation variant. Keep framework code self-contained while sharing the benchmark record contract.

For an existing C/C++ CMake target, `tools/scaffold_framework.py` creates a
Dockerfile, embedded manifest, workload-free smoke script, and integration
checklist. For example:

```sh
python tools/scaffold_framework.py example-cpp \
  --cipher Example --operation encrypt --backend native --language cpp \
  --source frameworks/example --target example_benchmark
```

The generated files stay outside the matrix and catalog until their source,
timed boundary, dependencies, runtime libraries, and result pipeline are
reviewed. The scaffold only covers C/C++ CMake projects; use the proven image
recipes and the contract below for other languages.

## 1. Decide the integration unit

Create a new top-level framework directory only when the implementation has its own upstream project, build system, dependency graph, or license. Add a scheme inside an existing framework when it uses that framework's runtime and build.

Choose canonical metadata before coding:

- cipher: stable lowercase scheme/cipher name;
- language: c, cpp, go, python, or rust;
- operation: a stable semantic name used consistently in manifests and records;
- backend: HE/runtime backend such as tfhe, helib, seal, openfhe, cuda, or native;
- security_profile: HHE-128 when the parameters satisfy the documented normalization, otherwise an honest distinct value or null;
- implementation: unique stable job ID.

Do not label a parameter set HHE-128 without checking security_level_normalization.md.

## 2. Required directory documentation

Add a README beside the implementation containing:

- upstream paper and source links;
- license and redistribution notes;
- supported OS/CPU/GPU;
- exact compiler/runtime and dependency versions;
- memory and expected runtime guidance;
- build commands from a clean checkout;
- one cheap functional command;
- one explicit benchmark command;
- generated files and cleanup instructions;
- known limitations.

Never make a broad test command the only documented smoke check.

## 3. Integrate the shared harness

Use the utility for the implementation language:

- benchmarking-utilities/c
- benchmarking-utilities/go
- benchmarking-utilities/python
- benchmarking-utilities/rust

Framework-local copies are permitted when the framework must build independently, but they must remain behaviorally identical. The target operation boundary must exclude setup, key generation unless key generation is itself the operation, memory sampling, serialization, logging, and file output.

The harness must support:

- HHE_BENCH_CIPHER
- HHE_BENCH_LANGUAGE
- HHE_BENCH_BACKEND
- HHE_BENCH_SECURITY_PROFILE
- HHE_BENCH_IMPLEMENTATION
- HHE_BENCH_ITERATIONS
- HHE_BENCH_WARMUPS
- HHE_BENCH_CPU
- HHE_BENCH_GC where applicable
- HHE_BENCH_POLICY
- HHE_BENCH_OUTPUT

Emit one JSON object per warm-up or measured iteration. Preserve exact integer time_ns, serialized cycles when available, signed memory deltas, failures, and output errors. Never clamp negative memory changes or fabricate unavailable values.

Honor `HHE_BENCH_HOSTNAME` before runtime hostname discovery so catalog
sweeps do not fragment complete-host report groups by ephemeral container ID.

Do not change the record shape locally. A necessary contract change must update schemas/benchmark.schema.json, docs/benchmark_schema.md, every language implementation, tests, and the schema version before framework code uses it.

## 4. Expose a concrete executable

The orchestrator starts executables; it does not invoke a compiler. Produce a prebuilt program or test binary with a narrow argument that selects only the intended job. Avoid an executable that unconditionally runs every scheme or parameter set.

A job should return zero only when its requested records were written successfully. Environment/preflight failures in strict mode, invalid options, operation failures, and JSONL write failures must return nonzero.

## 5. Add manifest entries

Validate against schemas/benchmark-manifest.schema.json. Add one entry per concrete cipher/language/operation/backend/security-profile job. Use unique stable IDs. Example:

    {
      "id": "rubato-go-native-encrypt",
      "cipher": "rubato",
      "language": "go",
      "operation": "SKE.Enc()",
      "backend": "native",
      "security_profile": "HHE-128",
      "executable": "/opt/hhe/frameworks/hheland-rubato",
      "args": ["--operation", "{operation}", "--iterations", "{iterations}"],
      "env": { "SCHEME_PARAMETER_SET": "rubato-2" }
    }

Use placeholders only for values supported by the executable. Do not put secrets, keys, or machine-specific temporary paths in a committed manifest.

Check:

    hhe list --manifest new-manifest.json
    hhe bench --manifest new-manifest.json --cipher rubato --dry-run

## 6. Docker integration

Use a framework-specific multi-stage image when dependencies are not already in the base tooling image:

1. pin builder and runtime image digests;
2. install and build dependencies in builder stages;
3. copy only executables and required runtime data into the final glibc image;
4. inherit or reproduce the root tooling runtime contract;
5. replace /opt/hhe/benchmarks/manifest.json;
6. keep compilers, package managers, source trees, and caches out of the final image;
7. document memory, CPU, GPU, and host mounts.

Do not Dockerize a benchmark until its native measurement boundary and JSONL output are stable.

## 7. Tests and validation

Use cheap tests that cannot trigger the HE workload:

- manifest schema validation;
- option parsing and selector tests;
- JSON serialization/schema tests using synthetic records;
- iteration/warm-up and failure-path tests with trivial callbacks;
- compile-only checks;
- CLI --dry-run;
- Docker image build and help/probe commands.

Keep the real HE functional test and benchmark as explicit opt-in commands. Document expected memory and time. Before publication, validate on the intended bare-metal host.

Acceptance checklist:

- clean-checkout build instructions work;
- cheap smoke check exists;
- executable runs exactly the selected job;
- all common CLI options reach the harness;
- raw records validate against schema 1.0.0;
- new records emit absolute u_maxrss_bytes (or 
ull) in addition to signed RSS, heap, and peak-RSS deltas;
- warm-ups are labeled and excluded by reports;
- GC and allocator semantics are honest;
- strict policy stops before measurement on failure;
- manifest selectors are unique and list correctly;
- report generator accepts synthetic output;
- framework README, root guide, Structure.md, and relevant citations are updated;
- no generated binaries, keys, raw results, or secrets are committed.

## 8. Review considerations

Reviewers should check the timed boundary first, then metadata accuracy, error propagation, and reproducibility. A numerically correct cipher with a misleading benchmark boundary is not ready for comparison.

If a framework cannot satisfy a metric, emit null and document why. If two modes are not comparable, give them different metadata so the report generator separates them.
