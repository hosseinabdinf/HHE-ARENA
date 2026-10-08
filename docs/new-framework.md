# New framework integration checklist

This checklist is the minimum contract for adding a framework to HHE-ARENA.
CI builds and bounded-smokes declared safe lightweight images; HE/HHE image
jobs remain workload-free. `tools/scaffold_framework.py` creates the initial
Dockerfile, manifest, smoke script, and checklist for C/C++ CMake targets.

## Required changes

1. Migrate active benchmark boundaries to the language adapter and retain each
   replaced helper/source beside it with an `_backup` suffix.
2. Add a version-1 manifest under `manifests/` using the shared JSONL schema.
   Every job needs a stable id, cipher, language, operation, backend, and
   executable path. Keep heavy jobs explicitly non-smoke.
3. Add a Dockerfile under the framework directory. It must pin its builder
   inputs, copy prebuilt executables into a compiler-free runtime, invoke
   `hhe-validate-framework-image --require-benchmarks`, and run as UID/GID
   `65532:65532`.
4. Register the image and its concrete variants in `manifests/images.v2.json`
   and `manifests/support-matrix.v2.json`.
5. Add a bounded smoke or dry-run command only when it is safe; never make CI
   execute HE/HHE work. Reports are generated later from validated JSONL by
   `benchmarking-utilities/report/report_generator.py`.

## Local gate

From the repository root, run:

```text
python tools/validate_framework_contract.py
python tools/validate_project_inventory.py
python tools/validate_json_schemas.py
```

These verify schemas, catalog/manifest/matrix selectors, Dockerfile validator
hooks, and non-root runtime declarations without executing a workload. The
schema script requires `jsonschema==4.23.0`; run it inside the tooling image
if the host Python has an older package.
