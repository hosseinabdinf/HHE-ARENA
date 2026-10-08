# Benchmark utility migration status

Last audited: 2026-10-08

This is a static source audit of framework benchmark call sites. No HE test or benchmark was executed. The status answers whether a cipher path actually uses the shared schema-1.0.0 harness; the presence of a copied utility file alone is not sufficient.

## Status definitions

- **New**: an active cipher/operation call site uses the new C/C++, Go, Python, or Rust contract and can emit one schema-valid JSONL record per warm-up or measured iteration.
- **Mixed**: a framework contains both new call sites and legacy/unintegrated benchmark paths, or only part of a cipher workflow is migrated.
- **Old**: benchmark call sites still use raw counters, console-only summaries, Criterion/Google Benchmark/Go `testing.B`, or the pre-schema local helper.
- **Not integrated**: no active call site using the common harness was found. The component may still have upstream tests or benchmarks.
- **Utility only**: a new utility exists, but no active cipher call site was found invoking it.

The new contract includes `HHE_BENCH_*` configuration, proportional or explicit warm-ups, affinity and environment policy, serialized counters where available, memory sampling outside timing, signed byte deltas, and per-iteration JSONL. See [benchmark_schema.md](benchmark_schema.md).

## Summary matrix

| Framework / component | Cipher or scheme paths | Status | Evidence and remaining gap |
|---|---|---:|---|
| AES / OpenSSL | AES-128-CBC reference encryption | **New** | Active `BENCHMARK_ITER` in `frameworks/AES/Openssl/src/aes_encrypt.c`; migrated C utility is linked. |
| AES / TinyAES | AES-128-CBC reference encryption | **New** | Active `BENCHMARK_ITER` in `frameworks/AES/Tinyaes/src/aes_encrypt.c`. |
| AES / Full-LUT | AES with TFHE full-LUT transciphering plus symmetric/key-expansion operations | **New** | Active `BENCHMARK` call sites in `frameworks/AES/Full-LUT/sources/main.cpp`. |
| AES / Fregata CBS `homoSM4_CB` | AES symmetric operations and TFHE++ transciphering | **New** | Active new macros in `sym_aes.cpp` and `hom_aes.cpp`; a compiler-free contract-v1 image exposes all five boundaries with backend-correct selectors. |
| AES / Fregata GBS | Gate-bootstrapped AES | **New** | A pinned, compiler-free, non-root contract-v1 image now exposes the one-shot GBS AES pipeline. Warm-up is zero because replaying the mutable pipeline without reconstructing ciphertext state is unsafe; original driver and CMake files are retained as `_backup`. |
| Fregata FBS/GBS SM4 | Functional/gate-bootstrapped SM4 | **Excluded** | The source boundaries remain migrated and backed up, but SM4 is intentionally excluded from the support matrix, CLI selectors, and planned image catalog. |
| AES / Fregata CBS primitive microbenchmarks | TFHE++ primitive and standard-cell microbenchmarks | **Not integrated** | `CBSmode/benchmark/` remains Google Benchmark because it is a low-level primitive suite, not a Fregata cipher entry point. The CBS `homoSM4_CB` scheme path is migrated separately. |
| AES / FbT | AES over OpenFHE | **Mixed** | The active key-scheduler measurement uses the new macro, but transcipherer setup and `HHE.Decomp()` measurements are commented out in `example/example_aes.cpp`. |
| AES / BtR (Lattigo) | AES-over-CKKS transciphering; BFV/BGV/CKKS and multiparty internals | **Mixed / dependency-blocked** | `examples/btr_cipher/main.go` actively wraps the shared Go harness, but packaging is blocked: the checkout omits `core/rlwe`, pinned upstream v6.0.0/v6.1.0 bases lack the extension's `ShallowCopy`/`BootstrapReal` APIs, and `platform_linux.go` uses removed `syscall.SYS_GETCPU`. Upstream Lattigo internal suites remain dependency-owned `testing.B` benchmarks. |
| AES / FFT-CBS | AES, AES half-CBS, circuit bootstrapping, HP-LHE | **New** | Active tests use the shared Rust harness; the four former Criterion binaries now resolve to the local schema adapter and retain `_backup` sources/manifests. A pinned, compiler-free, non-root image exposes eight non-smoke jobs across AES, half-CBS, CMUX, and HP-LHE. |
| AES / Hippogryph | Hippogryph AES / odd-modulus TFHE | **New** | Active AES measurements use a thin facade over the shared Rust schema harness. A pinned compiler-free non-root image exposes two non-smoke jobs (TFHE plaintext encryption and transciphering); no HE/HHE workload ran. |
| AES / P-encoding | P-encoding AES / odd-modulus TFHE | **New** | Active AES measurements use a thin facade over the shared Rust schema harness; the image includes its AES circuit data and both exact operations passed separate single-operation functional runs. |
| Elisabeth | Elisabeth clear and homomorphic paths | **New** | Clear/homomorphic tests and the standalone encryption/transciphering benchmark use the shared Rust schema harness; the original Criterion source is retained as `_backup`. |
| Fasta | Fasta symmetric and HElib; Rasta HElib, packed HElib, and TFHE | **New** | All five active executables use schema-1.0.0 sessions. Reset/cleanup hooks keep input reconstruction and disposal outside timing; legacy sources are retained as sibling `_backup` files. |
| FiLIP / independent-PTX | FiLIP p-dependent setup, SKE encryption, HHE decomposition | **New** | Active `BENCHMARK_ITER` call sites use migrated C utilities. |
| FiLIP / Sorting Hat | FiLIP SKE encryption and HHE decomposition | **New** | Active `BENCHMARK_ITER` call sites use migrated C utilities. |
| FiLIP / Rust reference | FiLIP encryption and transciphering | **New / dependency-blocked** | The active standalone benchmark uses the shared Rust harness and the Criterion source is retained as `_backup`; compilation still depends on an unavailable upstream FiLIP branch. |
| FRAST | FRAST setup, online phase, PBS generation, full benchmark | **New** | Active tests use the shared Rust harness; all four former Criterion binaries resolve to the local schema adapter and retain `_backup` sources/manifests. |
| HHE Framework | Agrasta, Dasta 5/6, FiLIP 1216/1280, Hera 5, Kreyvium variants, LowMC, Masta 4/5, Pasta/Pasta2 3/4, Rasta 5/6 and packed variants, HE-only paths | **New** | Seven common KAT headers expose 73 schema boundaries to 80 generated cipher/backend test-vector executables through the shared C/C++ harness. |
| HHELand / dedicated `TestNb*` paths | Hera, Pasta3, Pasta4, Rubato SKE and HHE decomposition; pure BGV and CKKS operations | **New** | Active `utils.Benchmark`/`BenchmarkIter` call sites use the migrated Go harness and honor CLI iteration overrides. |
| HHELand / standard scheme benchmark suites | Symmetric and HHE Hera, Pasta, Rubato; RtF Hera/Rubato | **New** | All scheme-level `testing.B` entry points delegate measurement to the shared Go harness and retain `_backup` sources. The copied Lattigo ring arithmetic microbenchmarks remain dependency-style and are excluded from scheme migration. |
| Margrethe | Margrethe C HHE decomposition and Python SKE | **New** | C `BENCHMARK_ITER` and Python `benchmark(...)` call sites both use migrated harnesses. |
| Transistor / symmetric | Transistor symmetric clock/encryption | **New** | `symmetric/main.py` uses the migrated Python harness. |
| Transistor / homomorphic | Homomorphic Transistor / odd TFHE | **New** | The Rust implementation uses a thin facade over the shared schema harness. Combined with the symmetric path, the framework benchmark helpers are new. |
| Trivium/Kreyvium (`trkr`) | Trivium and Kreyvium TFHE-rs application paths | **New** | Application tests use the shared Rust harness; the former Criterion application bench resolves to the local schema adapter and retains `_backup` sources/manifests. |
| YuX | Yu2x-8, Yu2x-16, Yupx-p; symmetric, key schedule, HE key encryption, transciphering and final HE decrypt operations | **New** | Test executables actively use migrated C/C++ `BENCHMARK` macros and link the new utility. |

### Current packaging blockers

YuX HElib transciphering (`yux-cpp`) is not runnable: `frameworks/yux/thirdparty/build.sh` clones an unpinned HElib branch and builds NTL, GMP, and M4RI during image creation. No reproducible compiler-free runtime contract exists, so the five HElib variants remain outside the image catalog. The native YuX symmetric image remains available.

AES/BtR is separately dependency/API-blocked: the checked-in extension requires Lattigo `core/rlwe`, `ShallowCopy`, and `BootstrapReal` APIs unavailable in compatible pinned revisions, while `platform_linux.go` references removed `syscall.SYS_GETCPU`. Its recipe and manifest remain source references only and it stays outside the image catalog.

The FiLIP Rust reference unit is also blocked because its required upstream FiLIP branch is unavailable; it remains outside the image catalog. Neither blocker involved workload execution.

### Margrethe workflow validation (2026-07-23)

Both Margrethe paths completed through the catalog CLI, Docker image, embedded manifest, benchmark runner, JSONL validation, and report generation with one measured iteration and zero warmups. Margrethe SKE completed successfully (`78,112,800 ns`); single-thread Margrethe HHE completed successfully (`2,529,476,178 ns`). Each produced one valid record and zero invalid records. The run used warning preflight on Docker Desktop/WSL; invariant/nonstop TSC and performance-governor checks were not verified, so these are functional validation results, not publication-grade performance data.

## Framework details

### AES family

**New paths**

- OpenSSL AES and TinyAES provide reference AES encryption measurements through `BENCHMARK_ITER`.
- Full-LUT uses the new utility for symmetric key expansion, TFHE symmetric encryption, and transciphering.
- Fregata CBS `homoSM4_CB` uses the new utility for AES key expansion, symmetric encryption, plaintext encryption, table construction, and AES transciphering.

**Mixed or old paths**

- FbT has a migrated utility, but only the key-scheduler call is active; the intended HHE decomposition call is commented.
- BtR's `examples/btr_cipher/main.go` actively measures AES `HEDecrypt` through the migrated Go utility. Lattigo's BFV/BGV/CKKS, ring, and multiparty microbenchmarks remain upstream, dependency-owned Go benchmark paths and are not treated as HHE-Arena scheme entry points.
- BtR Docker packaging was attempted with pinned upstream Lattigo bases but is
  not advertised in `images.v2.json` until the API and missing-source blockers
  are resolved.
- FFT-CBS, Hippogryph, and P-encoding active helper call sites route through the shared Rust schema harness. FFT-CBS's four standalone benches now use the Criterion-compatible schema adapter; its contract image packages all four binaries without running them.
- Fregata GBS AES records the complete mutable cipher pipeline through a one-shot schema session and is the only FBS/GBS path intended for the support matrix. The migrated FBS/GBS SM4 sources remain archival and are excluded from CLI/image scope. Old per-round clocks remain console diagnostics only. The separate CBS Google Benchmark directory is a primitive microbenchmark suite and is not treated as a scheme entry point.

### Fasta and Rasta

Fasta and Rasta are migrated to the C/C++ schema-1.0.0 session API. Active operations are:

- `sym/fasta_sym.c`: `ske_encrypt` and `ske_decrypt`;
- `hhe/fasta/fasta_helib.cpp`: `he_context_setup`, `he_key_switch_setup`, and `hhe_transcipher`;
- `hhe/rasta/rasta_helib.cpp`: the same setup/transcipher operations plus `cipher_matrix_setup`;
- `hhe/rasta_pack/rasta_pack_helib.cpp`: context setup, key-switch setup, and packed transciphering;
- `hhe/rasta_tfhe/rasta_tfhe.c`: TFHE context setup and Rasta transciphering.

The Fasta-local utility is now the shared C schema implementation with a local `HHE_BENCHMARK_ITER_HOOKS` wrapper. Its `before` and `after` hooks run outside the timer so mutable ciphertexts, HE contexts, keys, matrices, and temporary output can be reset or released without contaminating the measured boundary. Original sources and utility/CMake files are preserved beside the active files with `_backup` in their names.

Verification includes Docker executions of symmetric Fasta with one warmup and three measured iterations per operation. The `fasta-sym-rumaxrss.jsonl` run validated all eight records with zero issues and generated timing plus dedicated memory artifacts under `results/fasta/report-rumaxrss/`; median absolute `ru_maxrss_bytes` was 13,111,296 bytes for both operations. The utility and complete symmetric executable also compile/link under pinned GCC 14.2 with `-Werror`; active sources contain no direct legacy timing calls. HElib and TFHE executables were not run because their third-party dependencies are not present locally.

The native C path is now a separate compiler-free contract-v1 image with
distinct `ske_encrypt` and `ske_decrypt` manifest jobs. Decrypt-only dispatch
constructs its required ciphertext outside the measured boundary. Its image
build, contract validation, manifest listing, and host CLI dry-run passed
without executing Fasta during this packaging step; HElib and TFHE remain
deferred.

### HHE Framework

The binary and Z_p common KAT layers now use `benchmark_kat.h`, a C++ phase
adapter over the shared C session API. The four generated test drivers repeat
each KAT over `HHE_BENCH_WARMUPS` and `HHE_BENCH_ITERATIONS`; 73 operation
call sites record `ske_encrypt`, `ske_decrypt`, `preprocess_block`,
`plain_evaluate`, `he_encrypt`, `he_key_encrypt`, `hhe_decrypt`,
`he_postprocess`, `he_evaluate`, and `final_decrypt`. Setup, diagnostics, and
correctness comparisons run outside each named operation boundary.

Eighty `testvectors.h` consumers route through the seven migrated headers for
plain, HElib, SEAL, and TFHE backends. The shared utility is linked only when
`BUILD_TESTS` is enabled. Original headers, test templates, and CMake drivers
are preserved beside the active files with `_backup` names.

Validation did not execute HE code. GCC 14.2 compiled the adapter smoke with
`-Wall -Wextra -Werror`; binary and Z_p plain headers passed syntax-only
compilation; and the synthetic adapter run produced three records accepted by
the schema/report pipeline with zero invalid records. Full HElib, SEAL, and
TFHE compilation remains dependent on the framework's third-party libraries.

The 23 SKE/plain configurations are now split into the
`hhe-framework-plain-cpp` packaging unit. Its pinned, compiler-free, non-root
image contains one prebuilt executable per configuration and 69 concrete jobs
(`ske_encrypt`, `ske_decrypt`, and `plain_evaluate` for each). The KAT adapter
honors `HHE_BENCH_OPERATION`: prerequisite and correctness work still runs,
but only the selected boundary is measured and emitted. Image construction,
contract/schema validation, manifest listing, and a 69-job host CLI dry-run
passed without executing a cipher workload. HElib, SEAL, TFHE, and HE-only
targets remain deferred.

### HHELand

The recommended new-harness entry points are the dedicated `TestNb*` tests:

- `sym/hera/hera_test.go`: Hera `SKE.Enc()`;
- `sym/pasta/pasta_test.go`: Pasta3/Pasta4 `SKE.Enc()`;
- `sym/rubato/rubato_test.go`: Rubato `SKE.Enc()`;
- `hhe/hera/hera_test.go`: Hera `HHE.Decomp()`;
- `hhe/pasta/pasta_test.go`: Pasta3/Pasta4 `HHE.Decomp()`;
- `hhe/rubato/rubato_test.go`: Rubato `HHE.Decomp()`;
- `he/pure_bgv.go` and `he/pure_ckks.go`: HE encryption/evaluation operations.

The standard symmetric and HHE Go benchmark files, plus the RtF Hera/Rubato
integration suite, now delegate their named operations to `utils.BenchmarkIter`.
They preserve Go benchmark discovery but disable adaptive `testing.B` timing.
The `rtf_integration/ring` suite is a copied low-level Lattigo arithmetic suite,
not an HHE scheme benchmark, and remains upstream-style.

### YuX

YuX's three native symmetric executables are now separated from its five
HElib transciphering executables. The `yux-symmetric-cpp` contract image
prebuilds Yu2x-8, Yu2x-16, and Yupx-p without HElib and exposes nine jobs:
key expansion, encryption, and decryption for each variant. The shared YuX
macro honors `HHE_BENCH_OPERATION`; unselected prerequisite phases execute
once outside measurement, while only the selected phase emits JSONL. The
original macro header is retained as `utils_backup.h`. Image construction,
contract/schema validation, non-root/compiler-free inspection, listing, and
the nine-job host CLI dry-run passed without executing a cipher workload. All
HElib HE/HHE variants remain deferred.

### Full-LUT

Full-LUT AES/TFHE now has a pinned contract-v1 image and three production
jobs: native symmetric key expansion, TFHE symmetric encryption, and
transciphering. The build pins upstream TFHE commit
`ad518fe17f1b1578d0c603eb16d51cb348a4406e`, the exact revision matching the
project patch, and applies only a GCC 14 compatibility exception for the
legacy `restrict` warning. Operation selection executes unselected
prerequisites once outside measurement and emits only the requested boundary;
the original header is retained as `utils_backup.h`. Image construction,
contract/schema validation, linkage, non-root/compiler-free inspection,
three-job listing, CLI selection, and dry-run passed. No Full-LUT AES, TFHE,
or HHE workload was executed.

### Fregata CBS

Fregata CBS now has a pinned, compiler-free, non-root contract-v1 image with
two native AES jobs and three TFHE++ jobs. Selecting an operation executes only
earlier prerequisites outside measurement and exits immediately after the
requested boundary, preventing a setup-only job from falling through into the
full transciphering workload. The mutable AES transciphering job is fixed to
one measured iteration and zero warm-ups. The original source, utility header,
and corrected CMake file are retained as sibling `_backup` files. Image
construction, contract/schema validation, linkage, metadata, compiler-absence,
five-job listing, CLI selection, and dry-run passed. No AES, TFHE++, test, or
benchmark workload was executed.

### Rust-based frameworks

Elisabeth, FRAST, FFT-CBS, Hippogryph, P-encoding, homomorphic Transistor, and `trkr` now depend on `benchmarking-utilities/rust` through relative Cargo path dependencies. Their existing helper modules are thin compatibility facades, so active `benchmark(name, iterations, closure)` call sites emit schema-1.0.0 JSONL without duplicating timing code. The original seven helpers are retained as sibling `benchmark_utils_backup.rs` files; FFT-CBS also retains its original standalone bench helper as `benches/utils_backup.rs.txt`.

The shared wrapper honors `HHE_BENCH_OPERATION`, `HHE_BENCH_IMPLEMENTATION`, and `HHE_BENCH_WARMUPS` in addition to the existing common controls. A synthetic smoke test verified one warm-up plus two measured records and the requested cipher, operation, and implementation identity. No HE operation was executed.

FRAST, FFT-CBS, and `trkr` retain their existing Criterion-shaped benchmark
source API, but their dev dependency now resolves to
`benchmarking-utilities/rust-criterion-compat`. That adapter maps
`bench_function`, benchmark groups, and input benchmarks to one-iteration
shared schema sessions; it does not execute Criterion's adaptive statistical
engine. Original bench sources and Cargo manifests are retained with
`_backup` names. Elisabeth's former Criterion binary and the archival FiLIP
Criterion binary use direct shared-harness entry points.

## CLI and manifest readiness

Benchmark-harness migration and CLI readiness are separate:

- A **new** call site can consume the common environment and emit JSONL.
- The `hhe` CLI additionally needs a prebuilt executable and a concrete version-1 manifest entry.
- The root tooling image deliberately ships an empty manifest and no framework HE binaries.
- `manifests/example.json` is illustrative, not a production sweep manifest.

Therefore, none of the framework rows should be interpreted as automatically runnable from the base Docker image. Framework-specific binaries and validated manifests still need to be built or supplied.

## Ideal end-to-end goal readiness

The target product is a host-independent CLI in which a user selects a
framework, cipher, backend, operation, and benchmark policy; HHE-ARENA builds
or selects the correct pinned Docker image, runs the benchmark through the
shared contract, and returns validated raw data plus reports. A new framework
should be addable through a documented adapter, Docker build, and manifest
entry rather than changes throughout the orchestrator.

Overall readiness is not expressed as a percentage while matrix v2 remains
draft. The measurement and reporting foundation and matrix-v1 container slice
are complete; HE/HHE functional evidence, strict publication-host evidence,
image publication, and broader CI remain.
Completion means that every supported matrix choice has a matching image and
manifest, a verified contract, functional evidence appropriate to its workload,
and documented limitations. Unsupported choices retain explicit blockers.

| Order | Required capability | Current readiness | Remaining work / completion criterion |
|---:|---|---:|---|
| 1 | Shared cross-language measurement contract | **100%** | Schema 1.0.0 and maintained C/C++, Go, Python, and Rust harnesses exist. Future changes must remain backward compatible or version the schema. |
| 2 | Migrate project-owned scheme benchmark boundaries | **95%** | Non-FbT scheme paths are migrated. AES/FbT is explicitly deferred; dependency-owned primitive suites are outside the scheme support matrix unless later promoted. |
| 3 | Stable CLI selection and orchestration | **100%** | `hhe list`, `bench`, and `bench-all` consume the matrix-v2 hierarchy `family -> path -> cipher -> variant -> framework -> backend -> language`, accept `all`, show declared/effective readiness, and execute only ready targets backed by the image catalog. |
| 4 | Schema validation and report generation | **100%** | Catalog runs now standardize output as a `raw/` plus `report/` bundle and automatically invoke schema validation and CSV/LaTeX/pgfplots timing and memory reporting after every selected container succeeds. |
| 5 | Define the supported framework/cipher/backend matrix | **Inventory complete; release scope open** | Frozen v1 remains the runnable four-image slice. Draft `manifests/support-matrix.v2.json` separates 29 build/readiness units from 103 source-audited benchmark variants: 10 AES/conventional, 72 HE-friendly, and 21 pure HE. Native Fasta, HHE Framework plain/SKE, and YuX symmetric are separate from their deferred HE/HHE packages. |
| 6 | Reproducible Docker build for every supported framework | **90%** | Twenty-six of 29 v2 candidates have contract-v1 images. The three outside the catalog are dependency-blocked BtR, FiLIP/Rust reference, and YuX HElib. Heavy packaged images remain subject to their matrix execution status. |
| 7 | Production manifests for concrete binaries | **90%** | Twenty-six v2 candidates have production manifests and are listed in `manifests/images.v2.json`. Catalog presence alone does not make a deferred matrix choice runnable. |
| 8 | End-to-end container workflow | **90%** | All five matrix-v1 symmetric targets completed build/select -> warning preflight -> JSONL validation -> reports on Windows Docker Desktop and native Linux. Catalog execution wires validation/report bundles automatically and gives every language container one stable host identity. Heavy HE validation still requires explicit authorization. |
| 9 | Host-platform independence | **45%** | Windows Docker Desktop and native Linux are functionally verified. Native Linux preflight still could not confirm RDTSCP, performance governor, or Turbo state. macOS Docker and ARM remain unvalidated; publication-grade cycle measurements require a compatible, controlled host. |
| 10 | Easy addition and migration workflow | **Template available** | `tools/scaffold_framework.py` creates a C/C++ Dockerfile, manifest, workload-free smoke script, report path, and integration checklist. Other languages use the documented contract and existing recipes. |
| 11 | CI, release artifacts, and reproducibility evidence | **Build/smoke jobs added; publication open** | CI validates schemas and inventory, builds all ten lightweight images, runs one bounded safe smoke and report per image, and offers a separate workload-free heavy build job. Immutable registry digests and controlled-host HE evidence remain open. |

### Recommended implementation order

1. **Complete:** matrix version 1 freezes AES/OpenSSL C, HHELand/Rubato Go,
   Elisabeth-clear Rust, and Transistor-symmetric Python as the first set.
2. **Complete:** framework-image contract version 1 fixes executable and
   manifest paths, output mount, preflight, labels/digests, non-root runtime,
   validation, and failure semantics.
3. **Complete:** all four representatives have pinned multi-stage builds,
   prebuilt runtime artifacts, non-root metadata, and contract validation.
4. **Complete:** concrete per-image manifests plus
   `manifests/images.v1.json` let the CLI select and locally build, require,
   or immutably pull the appropriate image; dry-run orchestration is verified.
5. **Complete:** catalog runs automatically produce standardized raw/report
   bundles, normalize host identity across language containers, and all five
   lightweight targets completed through report generation on Windows Docker
   Desktop.
6. **Complete:** draft matrix v2 inventories 29 project-owned non-FbT
   packaging units and 103 source-audited benchmark variants. It
   defines the final family/path/cipher/variant/framework/backend/language
   selector taxonomy without creating speculative image catalog entries.
7. **Complete:** the CLI consumes the v2 selector taxonomy for listing and
   filtering; only targets backed by ready image-catalog entries may execute.
8. **Complete:** TinyAES, all 16 HHELand Rubato/Hera/Pasta symmetric
   variants, Margrethe SKE, native Fasta encrypt/decrypt, and all 23 HHE
   Framework plain variants now have contract-v1 images and manifests. YuX's
   three symmetric variants add nine operation jobs, completing the safe split
   of every previously mixed unit.
9. **Complete:** Full-LUT AES/TFHE is the next heavy contract image, with
   three backend-correct jobs and no workload execution during packaging.
10. **Complete:** Fregata CBS now packages two native AES and three TFHE++
   jobs without running its AES or homomorphic workload.
11. **Complete:** Fregata GBS AES is packaged as a pinned compiler-free
   contract image with one non-smoke job; no AES or TFHE++ workload ran.
12. **Complete (functional):** Repeated the five lightweight matrix-v1 operations
   on native Linux with warning preflight, schema validation, and reports.
13. **Complete:** FFT-CBS AES/Rust is packaged as a pinned compiler-free
   contract image with eight non-smoke jobs; no HE/HHE workload ran.
14. **Complete:** Hippogryph AES/Rust is packaged as a pinned compiler-free
   contract image with two non-smoke jobs; no HE/HHE workload ran.
15. **Complete:** P-encoding AES/Rust is packaged as a pinned compiler-free
   contract image with two non-smoke jobs; no HE/HHE workload ran.
16. **Complete:** Added `tools/validate_framework_contract.py`, the new-framework
   integration checklist, and a workload-free GitHub Actions contract gate.
17. **Complete (bounded):** Transistor HHE ran one measured iteration with
   preflight and report validation; WSL warnings mean its cycles are not publication-grade.

The immediate critical path is therefore **bounded HE/HHE validation ->
publication-host and release evidence -> immutable image publication**. Migrating more
internal microbenchmarks does not materially advance the ideal goal unless
those microbenchmarks are added to the declared support matrix.

### Elisabeth HHE packaging verification (2026-10-07)

The existing `elisabeth-hhe-rust` image was rebuilt from its pinned Dockerfile.
Its embedded framework-image contract check passed for one benchmark job, and
`hhe list` inside the image listed `Elisabeth/HHE.Decomp`. The host CLI also
listed the matching v2 matrix choice. The choice remains `deferred` and
`bench --dry-run` rejects it under the current execution gate. No Elisabeth
HHE workload was run.

### Inventory reconciliation (2026-10-07)

The v2 matrix has 29 implementation units and 103 variants; the catalog has
26 images. The host CLI lists 515 operation choices, including 120 marked
runnable. `tools/validate_project_inventory.py` now checks implementation and
variant IDs, catalog readiness, matrix dispatch selectors, and manifest jobs
in CI. It exposed and resolved the Transistor HHE backend spelling mismatch
and an FFT-CBS HP-LHE mapping that pointed at a different operation. The
unmapped HP-LHE manifest job remains packaged but deferred until its exact
variant selector is represented. These checks do not execute workloads.

### HHE Framework HE/HHE packaging and CI (2026-10-07)

The HHE Framework HE/HHE image built from exact HElib, SEAL, and TFHE source
revisions plus pinned Debian build dependencies. Its runtime contains 55
prebuilt backend executables and no GCC. The embedded manifest has 217 jobs;
contract validation, shared-library checks for all executables, and `hhe list`
passed without executing an HE/HHE workload. The catalog maps 55 variant and
backend selectors. The matrix still marks all those targets deferred, and a
CLI dry-run rejects their execution.

CI now has JSON Schema and cross-file static checks, ten lightweight
build/smoke/report jobs with retained JSONL artifacts, and a separate manual
heavy build-only workflow. The new CI smoke path was exercised locally on
OpenSSL: one valid record, zero invalid records, and a generated report.
GitHub-hosted CI has not yet run these new jobs. Strict publication preflight
failed here because the host is virtualized, RDTSCP was not reported, and the
governor and Turbo state were unconfirmed. See `docs/release-protocol.md`.

### Matrix-v1 functional smoke evidence

On 2026-10-07, the AES/OpenSSL catalog entry built and ran on native Linux
with one measured iteration and zero warm-ups. Its image passed contract
validation; the report pipeline accepted one JSONL record with zero invalid
records. The run used warning policy without preflight, so it establishes
functional catalog execution but does not validate publication conditions.
The OpenSSL Dockerfile pin was updated to the available Bookworm
`libssl-dev` version, and catalog Docker runs now use the host UID/GID so
their non-root processes can write to host-owned result bundles.

The same native Linux host then completed all five matrix-v1 catalog operations
with one measured iteration, zero warm-ups, and warning preflight. The report
validated five records with zero invalid records and generated timing and memory
artifacts under `/tmp/hhe-matrix-v1-linux-complete-20261007/`. Preflight passed
Linux, x86-64, constant/nonstop TSC, and the 32 GiB memory floor, but did not
confirm RDTSCP, a uniform performance governor, or Turbo state. CPU pinning
was not requested. Transistor emitted null cycles as designed when RDTSCP is
unavailable. These results are functional evidence, not publication data.

On 2026-07-14, Windows Docker Desktop/WSL built all four matrix-v1 images and
ran the five declared symmetric operations with one measured iteration, zero
warm-ups, CPU 2 pinning, and warning-mode preflight. The report pipeline
accepted all five records, found zero invalid or failed records, and generated
timing and memory CSV, LaTeX, and pgfplots artifacts under
`results/matrix-v1-smoke-20260714/`.

This is functional evidence only. Preflight reported virtualization, missing
invariant/nonstop TSC confirmation, unknown governor and Turbo state, and less
than 32 GiB memory. The recorded cycle and timing values must not be used as
publication comparisons. The run also exposed and fixed Transistor's
undeclared Sage dependency; its compiler-free image now verifies a
standard-Python GF(17) known-answer path during construction.

### Transistor HHE functional evidence

On 2026-07-14, the Transistor homomorphic Rust/odd-modulus-TFHE image built
from its pinned Rust builder, passed framework-image contract validation, and
completed one explicitly authorized `HHE.Decomp()` iteration with zero
warm-ups, CPU 2 pinning, and a 24 GiB Docker memory cap. The schema-valid
record and generated reports are under
`results/transistor-hhe-heavy-20260714/`; the measured boundary was
291,458,474 ns and reported absolute peak RSS was 159,170,560 bytes.

This is functional evidence only. Windows Docker Desktop/WSL did not verify
invariant/nonstop TSC, governor, or Turbo state, so its cycle count is not
publication data.

## Recommended migration order

Project-level scheme migrations are complete as of 2026-07-14, with AES/FbT
intentionally left unchanged. Dependency-owned or low-level microbenchmark
suites remain upstream-style: Lattigo internals in BtR and HHELand ring,
Fregata CBS primitive Google Benchmarks, and vendored TFHE benches. These are
not HHE-ARENA scheme entry points and should only be normalized if they become
inputs to a published cross-framework comparison.

## How to update this document

When a path is migrated:

1. verify an active cipher call site invokes the new harness;
2. use a trivial or synthetic smoke path to confirm schema-valid JSONL without running the expensive HE workload;
3. add or validate its manifest entry separately;
4. update the row and audit date here;
5. do not mark an entire framework new while legacy benchmark executables remain.

## Matrix v2 completion pass (2026-10-07)

All ten safe-lightweight catalog images built and passed a one-iteration,
zero-warmup functional smoke run. Their JSONL records passed schema validation
and report generation. All sixteen heavy images built, listed their expected
jobs, and passed embedded-manifest and shared-library checks without running
benchmark workloads. The first pass found and fixed missing FRAST and trkr
Cargo lockfiles, missing runtime libraries in both FiLIP C++ images, a
Margrethe build-time RDTSCP assumption, and HHE Framework's C++ runtime link.

The orchestrator Go tests, project inventory check, and framework contract
check pass. Representative HHE Framework jobs produced four complete valid
records across SEAL and TFHE HE encryption. HElib HHE decryption produced two
valid records, then exceeded a ten-minute run bound; its container was stopped.
These samples do not promote any deferred matrix target. Strict publication
preflight still fails on the virtualized host; exact deferred HE/HHE workload
evidence and immutable registry pull references remain open. Local logs and
generated functional artifacts are kept under `/tmp` for this workspace
session and are not publication evidence.

### Deferred HE/HHE selector review (2026-10-07)

AES/Full-LUT `Transciphering` is the first target reviewed individually. Its
matrix selector and CLI dry-run resolve to
`hhe-arena:full-lut-aes-matrix2` and the expected manifest job. Functional
execution is blocked on this host because Docker socket access is denied; the
target also requires a suitable high-memory host. No readiness was changed.
Dry-run output is under `/tmp/hhe-first-deferred-dryrun`.

After Docker access was granted through the execution sandbox, a one-iteration,
zero-warmup Full-LUT `Transciphering` run started. Key generation completed,
but TFHE encryption remained CPU-bound beyond the ten-minute bound, using
about 33 GiB of RAM and emitting no complete JSONL record. The container was
stopped and verified absent. Both Full-LUT HE/HHE selectors are now deferred;
the native key-expansion selector remains ready. The empty run artifact is
under `/tmp/hhe-first-deferred-functional`.

The next target, AES/Fregata CBS `AES Transciphering`, also resolves through
the matrix and CLI dry-run to `hhe-arena:fregata-cbs-aes-matrix2` and its
matching manifest job. Its first attempt was limited to a dry-run while Docker
socket access was unavailable in the default sandbox. The dry-run output is
under `/tmp/hhe-second-deferred-dryrun`.

On retry with Docker access, all three Fregata CBS TFHE++ operations completed
one measured iteration each with zero warmups: `Encrypt Plaintext`, `Make SBox
Table`, and `AES Transciphering`. Each produced one `status=ok` record, and
each report validated one record with zero issues. The transciphering boundary
took 241.85 seconds; the other two were below one millisecond. The local image
ID was `sha256:0bd046a07c61513529116676dd587705691b26f75ab7768ffc7a1ae949238f3a`.
The VM warning policy reported no RDTSCP, non-performance governor, and unknown
Turbo state, so these are functional checks only. Artifacts are under
`/tmp/hhe-second-deferred-functional`, `/tmp/hhe-fregata-encrypt-plaintext`,
and `/tmp/hhe-fregata-sbox-table`.

AES/Fregata GBS was the next scheme. Its exact matrix selector was deferred,
so a narrow `--validate-deferred` mode was added to the host CLI. The mode
requires exact selectors, one measured iteration, zero warmups, local image
policy, a selected CPU, warning policy, and a single cataloged deferred
operation; it does not alter matrix readiness. Its CLI dry-run resolved to
`hhe-arena:fregata-gbs-aes-matrix2`. A first one-iteration attempt stayed in
the first TFHE `SubBytes` step for ten minutes. A second attempt ran for twenty
minutes: the first `SubBytes` took about seventeen minutes, round one
completed, and round two remained in `SubBytes`; no complete JSONL record was
emitted. Both containers were stopped and verified absent. The full AES
pipeline needs a faster host or a longer bounded run, so it stays
deferred. Empty artifacts are under `/tmp/hhe-fregata-gbs-functional` and
`/tmp/hhe-fregata-gbs-functional-retry`.

After these changes, orchestrator Go tests pass; inventory validation reports
29 implementations, 103 variants, 26 images, and 73 ready targets; framework
contract validation passes for all 26 images; and JSON Schema validation
passes for all 32 documents. `--validate-deferred` behavior and bounds have
dedicated Go coverage.

The catalog security-profile scan found two explicit profiles, both AES-128,
and zero HHE-128 claims. The latest full CLI listing (2026-10-08) contains 435
operation choices: 164 runnable, 269 deferred, and two blocked. A full
`bench-all --dry-run` generated dispatch plans for all 164 runnable choices,
skipped the 271 non-ready choices, created no output bundle, and started no
benchmark containers. The selector counts are also recorded in
[`support-matrix.md`](support-matrix.md).

FFT-CBS AES_TIGHT completed its three exact operations serially: `AES
SYM.Enc()`, `HHE.Decomp()`, and `AES evaluation`. Each one-iteration,
zero-warmup invocation produced 29 `status=ok` records and a report with zero
invalid records. The local image ID is
`sha256:45415db2c8886924325ceb87d92c59150e6e3a2ae3099aeeb87a05f14a5c6977`.
The host records warn that RDTSCP and performance governor checks failed, so
these results validate functionality only. The AES_TIGHT HHE target is now
ready; AES_HALF_CBS and the pure-HE FFT-CBS targets remain deferred. Artifacts
are under `/tmp/hhe-fft-cbs-tight-hhe-functional`,
`/tmp/hhe-fft-cbs-tight-ske-functional`, and
`/tmp/hhe-fft-cbs-tight-aes-eval-functional`.

FFT-CBS AES_HALF_CBS then completed the same three operations serially. Each
report validated 28 records with zero issues, and all raw rows had
`status=ok`. Its target is now ready. Artifacts are under
`/tmp/hhe-fft-cbs-half-hhe-functional`,
`/tmp/hhe-fft-cbs-half-ske-functional`, and
`/tmp/hhe-fft-cbs-half-aes-eval-functional`. The remaining pure-HE FFT-CBS
profiles are still deferred.

AES/Hippogryph then completed `TFHE Encryption of the plaintext` and
`Transciphering` in separate one-iteration invocations. Each produced two
valid records and passed the AES known-answer output check; each report had
zero issues. The target is ready. The local image ID is
`sha256:beffa0efe1f6819623b4c45a29fd7f4629801fd1164a3cdaaa82bbd66e35f906`.
Artifacts are under `/tmp/hhe-hippogryph-he-encryption-functional` and
`/tmp/hhe-hippogryph-transciphering-functional`. These remain functional
results only because host preflight warnings preclude performance use.

P-encoding's earlier two-operation invocations exposed a shared Rust harness
bug: each call in a process inherited the selected operation label, so one
invocation could emit a second, expensive operation under the wrong name. The
shared helper now skips unselected calls. P-encoding prepares encrypted input
outside the measurement when transciphering is selected, preserving the AES
known-answer check while measuring only that operation. Rebuilt image
`sha256:e4989237e646d7a1e37191b86871e7b4bdc0010af9b5fce0dafbe26a61c5554a`
passed contract checks. Encryption completed in 6.48 ms with exactly one
valid selected-operation record. Transciphering completed in 434.97 seconds,
returned the expected AES ciphertext, and its report validated exactly one
record with zero invalid records. The target is ready for functional runs;
host preflight warnings still preclude performance claims. Artifacts are
under `/tmp/hhe-pencoding-selected/encryption.jsonl` and
`/tmp/hhe-pencoding-selected/transciphering.jsonl`.

HHE Framework's HE-only Z2/TFHE target then completed `he_encrypt`,
`he_evaluate`, and `final_decrypt` as separate one-iteration runs. Each report
validated one record with zero invalid records, and all five input rows per
operation reported `OK`. This target is ready for functional runs. Artifacts
are under `/tmp/hhe-hheframework-tfhe-z2-he-encrypt.jsonl`,
`/tmp/hhe-hheframework-tfhe-z2-he-evaluate.jsonl`, and
`/tmp/hhe-hheframework-tfhe-z2-final-decrypt.jsonl`. The host still fails
strict performance preflight.

HHE Framework's HE-only Zp variant completed `he_encrypt`, `he_evaluate`, and
`final_decrypt` separately for both SEAL and HElib, with one iteration per
run. Each operation/backend report validated three records (one per parameter
configuration) with zero invalid records, and all three parameter cases
reported `OK`. This target is ready for functional runs. Artifacts are under
`/tmp/hhe-hheframework-seal-he-encrypt.jsonl`,
`/tmp/hhe-hheframework-seal-he-evaluate.jsonl`,
`/tmp/hhe-hheframework-seal-final-decrypt.jsonl`,
`/tmp/hhe-hheframework-helib-he-encrypt.jsonl`,
`/tmp/hhe-hheframework-helib-he-evaluate.jsonl`, and
`/tmp/hhe-hheframework-helib-final-decrypt.jsonl`. These are functional
results only; the host fails strict performance preflight.

HHELand's pure-BGV target mapping now carries the selected parameter variant.
The BGV test selects the corresponding config, and the shared Go helper skips
unselected operations. The rebuilt image has separate BGV encryption and
evaluation jobs. N13QP218, N14QP438, and N15QP880 each passed both operations
in separate one-iteration runs and matched the reference output. Every report
validated one record with zero invalid records. All three BGV matrix targets
are ready. Artifacts are under `/tmp/hhe-hheland-bgv-n13-encrypt.jsonl`,
`/tmp/hhe-hheland-bgv-n13-evaluate.jsonl`, `/tmp/hhe-hheland-bgv-n14-encrypt.jsonl`,
`/tmp/hhe-hheland-bgv-n14-evaluate.jsonl`, `/tmp/hhe-hheland-bgv-n15-encrypt.jsonl`,
and `/tmp/hhe-hheland-bgv-n15-evaluate.jsonl`. The three CKKS variants are also ready, as documented below.


HHELand's pure-CKKS targets now carry the selected parameter variant and
operation. N14QP438, N15QP881, and PN16QP1761 each passed encryption, linear
evaluation, and Chebyshev evaluation as separate one-iteration runs. Each
report contained one valid record; the linear and Chebyshev checks passed, and
Chebyshev evaluation reported precision statistics. All three CKKS matrix
targets are ready for functional runs. Artifacts are under `/tmp/hhe-hheland-ckks-n14-encrypt.jsonl`,
`/tmp/hhe-hheland-ckks-n14-evaluate.jsonl`, `/tmp/hhe-hheland-ckks-n14-chebyshev.jsonl`,
`/tmp/hhe-hheland-ckks-n15-encrypt.jsonl`, `/tmp/hhe-hheland-ckks-n15-evaluate.jsonl`,
`/tmp/hhe-hheland-ckks-n15-chebyshev.jsonl`, `/tmp/hhe-hheland-ckks-pn16-encrypt.jsonl`,
`/tmp/hhe-hheland-ckks-pn16-evaluate.jsonl`, and
`/tmp/hhe-hheland-ckks-pn16-chebyshev.jsonl`. Strict performance publication
remains blocked by host preflight.


TRKR Trivium/Kreyvium catalog dispatch now maps each matrix variant to its
exact cipher name, so the selected operation matches the packaged manifest.
The first Trivium bool warmup attempt exposed a dispatch mismatch; after fixing
the mapping, the exact warmup ran for ten minutes without output or a record
and was stopped. The container is absent and the raw output file is empty at
`/tmp/hhe-trkr-trivium-bool-warmup.jsonl`. TRKR remains deferred pending a
suitable faster execution host.


HHE Framework Agrasta HHE targets now use the instrumented operation label
`he_key_encrypt` (previously listed as `he_encrypt`) and are tracked per
backend. SEAL `he_key_encrypt`, `hhe_decrypt`, and `final_decrypt` each passed
both correctness cases and produced two valid records, so the SEAL backend
target is ready. HElib key encryption emitted two valid records but its larger
correctness case exceeded ten minutes before completing, so HElib remains
deferred; TFHE has not yet been run. Artifacts are under
`/tmp/hhe-hhef-agrasta-seal-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-agrasta-seal-hhe-decrypt.jsonl`,
`/tmp/hhe-hhef-agrasta-seal-final-decrypt.jsonl`, and
`/tmp/hhe-hhef-agrasta-helib-he-key-encrypt.jsonl`.

Agrasta TFHE `he_key_encrypt` produced two schema-valid records, but the run
was stopped after ten minutes during the second correctness case's homomorphic
decryption. The complete selected invocation did not pass, so the TFHE target
remains deferred. Its partial artifact is under
`/tmp/hhe-hhef-agrasta-tfhe-he-key-encrypt.jsonl`.

Rasta-5 SEAL `he_key_encrypt` passed its first parameter case, then the larger
case remained in a single homomorphic round for more than ten minutes. Two
JSONL records emitted before the timeout validate, but the full correctness
case did not finish; Rasta-5 remains deferred. Artifact:
`/tmp/hhe-hhef-rasta5-seal-he-key-encrypt.jsonl`.


HHE Framework Pasta-3 on SEAL passed `he_key_encrypt`, `hhe_decrypt`, and
`final_decrypt` in separate runs. Each operation passed all six correctness
cases and produced six valid records with zero invalid. The SEAL backend target
is ready; Pasta-3 HElib remains deferred. Artifacts are under
`/tmp/hhe-hhef-pasta3-seal-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-pasta3-seal-hhe-decrypt.jsonl`, and
`/tmp/hhe-hhef-pasta3-seal-final-decrypt.jsonl`.


Pasta-3 HElib `he_key_encrypt`, `hhe_decrypt`, and `final_decrypt` each
passed all six correctness cases in separate serial runs. Each report validated
six records with zero invalid. The HElib backend is ready, so both Pasta-3 HHE
backends are now ready for functional runs. Artifacts are under
`/tmp/hhe-hhef-pasta3-helib-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-pasta3-helib-hhe-decrypt.jsonl`, and
`/tmp/hhe-hhef-pasta3-helib-final-decrypt.jsonl`. Strict performance
publication remains blocked by host preflight.


HHE Framework Pasta-4 SEAL `he_key_encrypt`, `hhe_decrypt`, and
`final_decrypt` each passed all six correctness cases in separate serial runs.
Each report validated six records with zero invalid. The SEAL backend is ready;
Pasta-4 HElib remains deferred. Artifacts are under
`/tmp/hhe-hhef-pasta4-seal-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-pasta4-seal-hhe-decrypt.jsonl`, and
`/tmp/hhe-hhef-pasta4-seal-final-decrypt.jsonl`. Strict performance
publication remains blocked by host preflight.


HHE Framework Pasta-4 HElib `he_key_encrypt`, `hhe_decrypt`, and
`final_decrypt` each passed all six correctness cases in separate serial runs.
Each report validated six records with zero invalid. The HElib backend is ready,
completing the Pasta-4 HHE backend matrix. Artifacts are under
`/tmp/hhe-hhef-pasta4-helib-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-pasta4-helib-hhe-decrypt.jsonl`, and
`/tmp/hhe-hhef-pasta4-helib-final-decrypt.jsonl`. Strict performance
publication remains blocked by host preflight.


HHE Framework Pasta2-3 SEAL `he_key_encrypt`, `hhe_decrypt`, and
`final_decrypt` each passed all six correctness cases in separate serial runs.
Each report validated six records with zero invalid. The SEAL backend is ready;
Pasta2-3 HElib remains deferred. Artifacts are under
`/tmp/hhe-hhef-pasta2-3-seal-he-key-encrypt.jsonl`,
`/tmp/hhe-hhef-pasta2-3-seal-hhe-decrypt.jsonl`, and
`/tmp/hhe-hhef-pasta2-3-seal-final-decrypt.jsonl`. Strict performance
publication remains blocked by host preflight.


Pasta2-3 plain SKE `ske_encrypt` completed all three correctness cases. Its
report validated three records with zero invalid. Artifact: `/tmp/hhe-hhef-pasta2-3-ske-encrypt.jsonl`. This is functional evidence for the encryption operation; no additional heavy HHE schemes were started after Pasta2.


### Deferred selector dispatch audit (2026-10-08)

The host CLI listed 435 matrix choices: 164 ready, 269 deferred, and 2
blocked. A `bench-all --dry-run` prepared 164 dispatches and skipped 271
non-ready choices; it created no result bundle and launched no containers.
Then each deferred operation was checked separately with exact selectors and
`--validate-deferred --dry-run`. Dispatch plans succeeded for 253 of 269.
The 16 uncataloged choices are BtR and YuX HElib, both dependency blocked; no
cataloged mapping failures remain. No benchmark workloads were run.

The audit corrected Fasta/Rasta, FiLIP C++, and FRAST catalog mappings, aligned
Elisabeth's HHE backend, and added exact HHELand RtF and FFT-CBS variant and
operation routing. HHELand now filters nested benchmark metrics by the exact
operation selected by the host runner. Inventory validation detects duplicate
variant/backend/operation mappings. Framework contract and inventory checks
pass for all 26 images and 29 implementations; all edited manifests parse and
`git diff --check` passes. The pinned schema check could not be repeated in
this environment because `jsonschema==4.23.0` and the tooling image are
unavailable. Further project phases await user direction. Heavy HHE schemes
remain stopped after Pasta2.

### Host-independent release preparation (2026-10-08)

The release protocol now records the proposed first-release scope: 26 cataloged
implementation images, with functional support claims limited to ready matrix
targets. Deferred targets remain unclaimed, and the three dependency-blocked
units plus AES/FbT remain excluded. This is still a proposal; matrix v2 remains
draft pending owner approval and completion of the release gates.

Current inventory is 29 implementation records, 103 variants, 73 ready
targets, 87 deferred targets, and 2 blocked targets. The catalog contains 26
images and no immutable pull references. Only the OpenSSL and TinyAES manifest
jobs declare `AES-128`; there are no `HHE-128` claims. A release evidence
record checklist now specifies commit and scope, per-target manifest selectors,
registry digests, raw result/report checksums, correctness outcomes, exclusions,
and host metadata.

Schema validation passes all 32 JSON documents; framework contract validation
passes all 26 catalog images; inventory validation and `git diff --check` pass.
No readiness state changed. Remaining gates are controlled-host correctness
and performance evidence, first-release scope approval, and registry
credentials and destination. Heavy HHE execution remains stopped after Pasta2.

A read-only local image audit found all 26 catalog tags present. Each image has
the required OCI contract labels, the expected matrix/manifest/results/resource
metadata, the `65532:65532` runtime user, and the `hhe-tooling` entrypoint. These
are local image IDs only; the catalog still has zero published pull references.

### FFT-CBS deferred pure-HE runtime samples (2026-10-08)

After user approval to resume, all ten deferred pure-HE FFT-CBS matrix
selectors ran sequentially as bounded functional invocations (one measured
iteration, zero warmups, CPU 0). They covered three CMUX selectors, two
improved HP-LHE selectors, and five high-precision HP-LHE selectors. Their
reports validated 2, 2, 2, 8, 4, 12, 6, 4, 16, and 8 records respectively,
with zero invalid records. CMUX1 was repeated once after fixing a Criterion
adapter operation-name mismatch; that check also validated two records.

The image was rebuilt with the exact variant and operation-prefix routing.
At the time of these sample runs, the sources printed maximum-error
diagnostics without asserting a correctness threshold. The samples therefore
establish runtime and JSONL behavior only; all ten targets remain deferred and
no performance claim follows. Run bundles
are `/tmp/hhe-fft-cbs-cmux1-retry-20261008`,
`/tmp/hhe-fft-cbs-cmux2-20261008`, `/tmp/hhe-fft-cbs-cmux3-20261008`,
`/tmp/hhe-fft-hplhe-improved-2-2-e1-20261008`,
`/tmp/hhe-fft-hplhe-improved-2-2-e2-20261008`,
`/tmp/hhe-fft-hplhe-highprec-3-3-e1-20261008`,
`/tmp/hhe-fft-hplhe-highprec-3-3-e2-20261008`,
`/tmp/hhe-fft-hplhe-highprec-3-3-e3-20261008`,
`/tmp/hhe-fft-hplhe-highprec-4-4-e1-20261008`, and
`/tmp/hhe-fft-hplhe-highprec-4-4-e2-20261008`. The CMUX1 adapter recheck is
under `/tmp/hhe-fft-cbs-cmux1-prefix-check-20261008`.

The pinned tooling image validates all 32 JSON documents. That run found
six RtF operation entries encoded as strings; these are now proper operation
objects. Project inventory validation passes (29 implementations, 103
variants, 26 images, 73 ready targets), and `git diff --check` passes. The
current VM still fails strict publication preflight. Heavy HHE scheme runs
remain stopped after Pasta2.

### FFT-CBS correctness check review (2026-10-08)

A temporary assertion on individual decrypted GGSW matrix coefficients used
half of each gadget spacing as a standalone correctness threshold. That is
not a valid scheme-level correctness test: the helper is a noise diagnostic,
and correctness must be checked on the consuming operation's output. The
assertion was removed. Two CMUX1 runs failed that diagnostic assertion at row
6 (the second reported `2296560889102336` against `2251799813685248`), but
those failures do not establish an incorrect final output. Their timing JSONL
records are not correctness evidence. CMUX1 remains deferred until an
end-to-end decrypted output check passes. No other FFT-CBS selector or HHE
scheme was run for this diagnostic.

### FFT-CBS CMUX1 output check (2026-10-08)

The invalid per-row assertion was removed. The benchmark now checks the
decrypted final LWE after subtracting the CMUX lookup result from the encrypted
input; for CMUX1's all-ones input and identity lookup, that difference must
decode to zero. The rebuilt image passed its 11-job contract check. One exact,
bounded CMUX1 run (one iteration, zero warmups, CPU 0) passed the output check
and generated two valid records with zero invalid. CMUX1 is now ready; the
other nine pure-HE FFT-CBS selectors remain deferred. Artifact:
`/tmp/hhe-fft-cbs-cmux1-e2e-20261008`. No further HHE scheme was run.
Current inventory is 75 ready, 85 deferred, and 2 blocked matrix targets; the
host CLI reports 166 ready, 267 deferred, and 2 blocked operation choices.

### FFT-CBS CMUX2 output check (2026-10-08)

The same final-output assertion passed for the exact CMUX2 selector in one
bounded run (one iteration, zero warmups, CPU 0). Its report validated two
records with zero invalid. CMUX2 is ready; CMUX3 and the seven HP-LHE selectors
remain deferred. Artifact: `/tmp/hhe-fft-cbs-cmux2-e2e-20261008`. No further
HHE scheme was run. Current inventory is 76 ready, 84 deferred, and 2 blocked
matrix targets; the host CLI reports 167 ready, 266 deferred, and 2 blocked
operation choices.

### FFT-CBS HP-LHE e2 output check (2026-10-08)

The exact `IMPROVED_WOPBS_2_2/e2` selector passed the final decrypted-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated four records with zero invalid. Both standard improved HP-LHE
extraction profiles are now ready. Artifact:
`/tmp/hhe-fft-hplhe-improved-2-2-e2-e2e-20261008`. Five high-precision
selectors remain deferred. Current inventory is 79 ready, 81 deferred, and 2
blocked matrix targets; the host CLI reports 170 ready, 263 deferred, and 2
blocked operation choices.

### FFT-CBS CMUX3 output check (2026-10-08)

The same final-output assertion passed for the exact CMUX3 selector in one
bounded run (one iteration, zero warmups, CPU 0). Its report validated two
records with zero invalid. All three CMUX selectors are now ready; seven
HP-LHE selectors remain deferred. Artifact:
`/tmp/hhe-fft-cbs-cmux3-e2e-20261008`. No further HHE scheme was run.
Current inventory is 77 ready, 83 deferred, and 2 blocked matrix targets; the
host CLI reports 168 ready, 265 deferred, and 2 blocked operation choices.

### FFT-CBS HP-LHE e1 output check (2026-10-08)

Added a decrypted final-output assertion to the ordinary and high-precision
HP-LHE benchmark paths. The rebuilt image passed its 11-job contract check.
The exact `IMPROVED_WOPBS_2_2/e1` selector passed in one bounded run (one
iteration, zero warmups, CPU 0); its report validated eight records with zero
invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-improved-2-2-e1-e2e-20261008`. Six HP-LHE selectors
remain deferred. Current inventory is 78 ready, 82 deferred, and 2 blocked
matrix targets; the host CLI reports 169 ready, 264 deferred, and 2 blocked
operation choices.

### FFT-CBS HP-LHE high-precision 3_3/e1 output check (2026-10-08)

The exact `HIGHPREC_IMPROVED_WOPBS_3_3/e1` selector passed the final-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated 12 records with zero invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-highprec-3-3-e1-e2e-20261008`. Four high-precision
selectors remain deferred. Current inventory is 80 ready, 80 deferred, and 2
blocked matrix targets; the host CLI reports 171 ready, 262 deferred, and 2
blocked operation choices.

### FFT-CBS HP-LHE high-precision 3_3/e2 output check (2026-10-08)

The exact `HIGHPREC_IMPROVED_WOPBS_3_3/e2` selector passed the final-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated six records with zero invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-highprec-3-3-e2-e2e-20261008`. Three high-precision
selectors remain deferred. Current inventory is 81 ready, 79 deferred, and 2
blocked matrix targets; the host CLI reports 172 ready, 261 deferred, and 2
blocked operation choices.

### FFT-CBS HP-LHE high-precision 3_3/e3 output check (2026-10-08)

The exact `HIGHPREC_IMPROVED_WOPBS_3_3/e3` selector passed the final-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated four records with zero invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-highprec-3-3-e3-e2e-20261008`. Two high-precision
selectors remain deferred. Current inventory is 82 ready, 78 deferred, and 2
blocked matrix targets; the host CLI reports 173 ready, 260 deferred, and 2
blocked operation choices.

### FFT-CBS HP-LHE high-precision 4_4/e1 output check (2026-10-08)

The exact `HIGHPREC_IMPROVED_WOPBS_4_4/e1` selector passed the final-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated 16 records with zero invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-highprec-4-4-e1-e2e-20261008`. Only the `4_4/e2`
high-precision selector remains deferred in the FFT-CBS subset. Current
inventory is 83 ready, 77 deferred, and 2 blocked matrix targets; the host CLI
reports 174 ready, 259 deferred, and 2 blocked operation choices.

### FFT-CBS HP-LHE high-precision 4_4/e2 output check (2026-10-08)

The exact `HIGHPREC_IMPROVED_WOPBS_4_4/e2` selector passed the final-output
assertion in one bounded run (one iteration, zero warmups, CPU 0). Its report
validated eight records with zero invalid. Promoted only this selector. Artifact:
`/tmp/hhe-fft-hplhe-highprec-4-4-e2-e2e-20261008`. All ten deferred
FFT-CBS pure-HE selectors now have output-level correctness evidence. Current
inventory is 84 ready, 76 deferred, and 2 blocked matrix targets; the host CLI
reports 175 ready, 258 deferred, and 2 blocked operation choices.

### FRAST pure-HE profile output checks (2026-10-08)

Fixed exact-operation routing for the FRAST Criterion groups by setting each
manifest prefix to the actual group ID. Added key-switch and PBS plaintext
assertions in `bench_gen_pbs.rs`, and added a decrypted-zero assertion to the
online benchmark using its zero-input fixture. The FRAST image rebuilt and its
13-job framework contract passed. Elisabeth and Kreyvium each passed their
`KS`, `GenPBS`, and `online` exact selectors serially; each run produced one
valid record and zero invalid. The FRAST profile had already passed its
`KS` and `GenPBS` selectors. Promoted all three PBS profile targets. Run bundles:
`/tmp/hhe-frast-pbs-ks-frast-20261008`,
`/tmp/hhe-frast-pbs-gen-frast-20261008`, and
`/tmp/hhe-frast-pbs-{ks,gen,online}-{elisabeth,kreyvium}-20261008`. The host
still fails strict publication preflight, so these functional samples do not
support performance claims. No additional SKE or HHE scheme was run. Current
inventory is 87 ready, 73 deferred, and 2 blocked matrix targets; the host CLI
reports 183 ready, 250 deferred, and 2 blocked operation choices.
Final local checks passed: inventory validation (29 implementations, 103
variants, 26 images), JSON Schema validation (32 documents), manifest JSON
parsing, and `git diff --check`. No containers remain running. The 73 deferred
targets comprise 68 HE-friendly HHE targets, 3 AES HHE targets, and 2 SKE
targets. HHE execution stays at the user's Pasta2 stop boundary; the remaining
SKE checks are beyond the requested current SKE evidence. There are no deferred
pure-HE targets. The two blocked targets remain the documented dependency
blockers.
