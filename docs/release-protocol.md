# Release and measurement protocol

## Current boundary

Matrix v2 remains a draft. The catalog has local build recipes, but no registry
digests. A successful image build or functional run is not a published
performance result. Strict preflight on 2026-10-08 detected a virtualized host
and failed to confirm RDTSCP, the performance governor, and Turbo state. Docker,
32 GiB minimum memory, and the selected CPU checks passed. This host can
provide functional evidence only.

The three dependency-blocked units are BtR, FiLIP/Rust reference, and YuX
HElib. AES/FbT is outside the declared v2 scope. Deferred choices stay
deferred until their exact manifest jobs have functional evidence. Per user
instruction, do not run further heavy HHE schemes beyond the completed
Pasta2 work. An earlier FiLIP SKE validation attempt used the full benchmark
binary and unintentionally ran HHE setup and decomposition; that attempt is
recorded as invalid evidence and was not repeated. The corrected SKE-only image
emits only its selected operation. Other authorized heavy scheme runs must be
sequential, one scheme at a time.

## Proposed first-release scope

The current candidate is the 26 implementation images already represented in
`manifests/images.v2.json`. Only matrix targets marked `ready` are eligible for
functional support claims; deferred targets remain visible but are excluded
from those claims. The three dependency-blocked units stay outside the image
catalog, and AES/FbT remains outside matrix v2. This is a proposal, not a
frozen release decision: the matrix remains `draft` until the owner approves
the scope and the release gates pass.

The current matrix contains 29 implementation records, 103 variants, 74 ready
targets, 86 deferred targets, and 2 blocked targets. The image catalog has 26
images and no immutable pull references. Its embedded manifests contain two
explicit `AES-128` profiles (OpenSSL and TinyAES) and no `HHE-128` claims.
Keep security profiles unset where evidence is absent.

## Functional acceptance for each supported image

1. Validate all JSON files and matrix/catalog/manifest relationships with
   `python tools/validate_json_schemas.py`,
   `python tools/validate_framework_contract.py`, and
   `python tools/validate_project_inventory.py`.
   The schema script requires `jsonschema==4.23.0`, which is installed in the
   tooling image and CI environment.
2. Build the image from its catalog recipe. The lightweight CI workflow runs one
   safe job per lightweight image with one measured iteration and zero warmups,
   then validates JSONL and generates a report. The separate heavy image job
   builds and lists an image without starting its benchmark.
3. On suitable hardware, inspect each deferred choice with `hhe list` and
   `hhe bench --dry-run` before an explicit bounded run. Archive the raw JSONL,
   generated report, manifest, image ID, source revision, and host metadata.
4. Promote a matrix target to `ready` only after its exact backend, variant,
   operation, and manifest selector pass. Re-run inventory validation after
   changing readiness.

## Comparable performance runs

Use native x86-64 Linux with an isolated CPU. Run
`bash docker/preflight.sh --cpu N --min-memory-gib 32` in strict mode and stop
if it exits nonzero. Record the CPU model, kernel, logical and physical CPU
topology, governor, Turbo setting, RDTSCP/invariant TSC flags, RAM, Docker
version, image digest, and source revision. Keep the selected CPU and sibling
threads idle, avoid swapping, and use `--cpuset-cpus=N` with the same harness
`--core N`. Use one fixed host ID across all containers.

For each comparable job, use the same security profile and semantic operation.
Run two warmups and ten measured iterations per invocation, then repeat three
invocations in a recorded order. Retain all raw records, including failures;
report both per-invocation and pooled summaries without combining different
hosts, backends, security profiles, or GC modes. A warning-policy result is
diagnostic only. Record any framework-specific memory or runtime limit before
the run; do not reduce iteration counts after seeing a result without marking
the changed protocol.

## Security profile audit

`security_level_normalization.md` requires both the HE and symmetric layers
to reach 128 bits, plus explicit audit metadata, before a row is labelled
`HHE-128`. The current cataloged manifests make **zero** HHE-128 claims. Two
AES-128 jobs use the conventional AES profile. `manifests/example.json` has an
illustrative HHE-128 label and is not a catalog image. Preserve null profiles
for unverified framework jobs. Before assigning HHE-128, record the exact
variant, HE parameters and estimator or standard evidence, symmetric security
claim, correctness evidence, and review decision in the release archive.

The 2026-10-07 catalog-manifest audit read all 26 manifest files and found two
explicit profiles: AES-128 in the OpenSSL and TinyAES jobs. It found zero
HHE-128 labels. All other cataloged jobs leave the security profile unset;
this audit does not infer HE security from parameter names or paper claims.

## Immutable publication

After all supported targets pass, build from a clean source revision, supply
the real tooling digest to each framework build, and push images to a chosen
registry. Capture each registry-provided `repository@sha256:...` reference and
put it in the corresponding catalog `pull_ref`. A local image ID is not a
registry pull reference. Re-run the static checks and a pull/list check using
the digest references. Freeze matrix v2 only when its declared release scope
has matching catalog entries, manifests, functional evidence, and immutable
pull references. Publish the archive with a checksum manifest and document
which platforms were actually verified; this repository currently has only
native Linux and Windows Docker functional evidence.

## Release evidence record

Create one release record alongside the archived run artifacts. It must identify
the exact commit and matrix scope and link each included target to its evidence.
For every target, record the image ID, registry `repository@sha256:...` digest,
embedded manifest job ID, variant/backend/operation selectors, raw JSONL and
report checksums, correctness result, and any known limitation. For each run,
record the host ID and preflight result; for publication comparisons, include
the full CPU, OS, governor, Turbo, TSC, memory, Docker, and affinity metadata
specified above. Record excluded targets with their dependency or evidence
blocker. Do not use a local image ID as a published digest or infer correctness
from a valid JSONL record.
