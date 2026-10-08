#!/usr/bin/env python3
"""Create a C/C++ CMake framework-image integration skeleton."""

import argparse
import json
from pathlib import Path
import re
import shlex


ROOT = Path(__file__).resolve().parents[1]
GCC = "gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("id", help="new lowercase slug, for example my-cipher-cpp")
    parser.add_argument("--cipher", required=True)
    parser.add_argument("--operation", required=True)
    parser.add_argument("--backend", required=True)
    parser.add_argument("--language", choices=("c", "cpp"), required=True)
    parser.add_argument("--source", required=True,
                        help="existing CMake source directory inside the repository")
    parser.add_argument("--target", required=True, help="CMake executable target")
    args = parser.parse_args()
    if not re.fullmatch(r"[a-z][a-z0-9-]*", args.id):
        parser.error("id must be a lowercase dash-separated slug")
    if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_-]*", args.target):
        parser.error("target must be a simple CMake target name")
    source = (ROOT / args.source).resolve()
    if ROOT not in source.parents or not (source / "CMakeLists.txt").is_file():
        parser.error("--source must be an existing CMake project in the repository")
    source_rel = source.relative_to(ROOT).as_posix()
    if not re.fullmatch(r"[A-Za-z0-9_./-]+", source_rel):
        parser.error("--source path contains characters unsupported by the Docker template")
    destination = ROOT / "frameworks" / args.id / "docker"
    manifest_path = ROOT / "manifests" / f"{args.id}.v2.json"
    if destination.exists() or manifest_path.exists():
        parser.error("destination already exists; choose a fresh id")

    executable = f"/opt/hhe/frameworks/{args.id}/{args.target}"
    manifest = {
        "$schema": "../schemas/benchmark-manifest.schema.json",
        "version": 1,
        "preflight": "/opt/hhe/bin/hhe-preflight",
        "benchmarks": [{
            "id": args.id,
            "cipher": args.cipher,
            "language": args.language,
            "operation": args.operation,
            "backend": args.backend,
            "executable": executable,
            "workdir": "/work",
        }],
    }
    dockerfile = f'''# syntax=docker/dockerfile:1.7
ARG GCC_IMAGE={GCC}
ARG TOOLING_IMAGE=hhe-arena:tooling
FROM ${{GCC_IMAGE}} AS builder
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \\
 && apt-get install -y --no-install-recommends cmake=3.25.1-1 \\
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY {source_rel}/ {source_rel}/
COPY benchmarking-utilities/c/utils/ benchmarking-utilities/c/utils/
# Adapt configure flags and runtime libraries to this framework's dependency graph.
RUN cmake -S {source_rel} -B /build -DCMAKE_BUILD_TYPE=Release \\
 && cmake --build /build --target {args.target} --parallel 2 \\
 && install -D -m 0555 /build/{args.target} /out/{args.target}

FROM ${{TOOLING_IMAGE}} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA {args.id}" \\
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \\
      org.opencontainers.image.revision="${{IMAGE_REVISION}}" \\
      org.opencontainers.image.version="${{IMAGE_VERSION}}" \\
      io.hhe-arena.contract.version="1" \\
      io.hhe-arena.matrix.id="{args.id}" \\
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \\
      io.hhe-arena.results.path="/work/results" \\
      io.hhe-arena.resource.class="review-required" \\
      io.hhe-arena.tooling.digest="${{TOOLING_DIGEST}}"
USER root
COPY --from=builder /out/{args.target} {executable}
COPY manifests/{args.id}.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \\
 && mkdir -p /work/results \\
 && chown -R 65532:65532 /work \\
 && hhe-validate-framework-image --require-benchmarks \\
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
'''
    readme = f'''# {args.id} integration

Source: `{source_rel}`. Review the generated build flags and add pinned builder
dependencies and required runtime libraries. Confirm that `{args.target}`
emits schema 1.0.0 JSONL using the shared harness and selects only the intended
`{args.operation}` boundary. Do not add this scaffold to the matrix or catalog
until its binary and contract pass.

Build and workload-free contract check:

```sh
docker build --pull=false -t hhe-arena:tooling .
docker build --pull=false -f frameworks/{args.id}/docker/Dockerfile \\
  -t hhe-arena:{args.id}-matrix2 .
docker run --rm --network none hhe-arena:{args.id}-matrix2 list
```

`smoke.sh` runs the workload-free list check by default. After the operation is
declared safe, `smoke.sh --safe` runs one measured iteration and validates its
JSONL through the report generator. Heavy HE/HHE work remains opt-in on
suitable hardware. Add matrix targets, a catalog entry,
security evidence, documentation, and CI classification after validation.
Then run all three `tools/validate_*.py` static checks.
'''
    smoke = f'''#!/usr/bin/env bash
set -euo pipefail
image=hhe-arena:{args.id}-matrix2
docker run --rm --network none "$image" list
if [[ "${{1:-}}" != "--safe" ]]; then
  exit 0
fi
output_dir=$(mktemp -d)
trap 'rm -rf "$output_dir"' EXIT
mount="type=bind,source=$output_dir,target=/work/results"
docker run --rm --network none --user "$(id -u):$(id -g)" \\
  --mount "$mount" "$image" bench --cipher {shlex.quote(args.cipher)} \\
  --language {shlex.quote(args.language)} --operation {shlex.quote(args.operation)} \\
  --backend {shlex.quote(args.backend)} --iterations 1 --warmup 0 --policy warn \\
  --output /work/results/raw.jsonl
docker run --rm --network none --user "$(id -u):$(id -g)" \\
  --mount "$mount" "$image" report /work/results/raw.jsonl \\
  --output-dir /work/results/report \\
  --schema /opt/hhe/schemas/benchmark.schema.json
'''
    destination.mkdir(parents=True)
    (destination / "Dockerfile").write_text(dockerfile, encoding="utf-8")
    (destination.parent / "README.md").write_text(readme, encoding="utf-8")
    smoke_path = destination.parent / "smoke.sh"
    smoke_path.write_text(smoke, encoding="utf-8")
    smoke_path.chmod(0o755)
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
    print(f"created {destination / 'Dockerfile'} and {manifest_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
