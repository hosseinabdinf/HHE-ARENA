# Benchmark utility migration status

Last audited: 2026-07-14

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
| AES / P-encoding | P-encoding AES / odd-modulus TFHE | **New** | Active AES measurements use a thin facade over the shared Rust schema harness. |
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

### Current packaging blockers`r`n`r`nYuX HElib transciphering (`yux-cpp`) is not runnable: `frameworks/yux/thirdparty/build.sh` clones an unpinned HElib branch and builds NTL, GMP, and M4RI during image creation. No reproducible compiler-free runtime contract exists, so the five HElib variants remain outside the image catalog. The native YuX symmetric image remains available.`r`n`r`nAES/BtR is separately dependency/API-blocked: the checked-in extension requires Lattigo `core/rlwe`, `ShallowCopy`, and `BootstrapReal` APIs unavailable in compatible pinned revisions, while `platform_linux.go` references removed `syscall.SYS_GETCPU`. Its recipe and manifest remain source references only and it stays outside the image catalog.`r`n`r`nThe FiLIP Rust reference unit is also blocked because its required upstream FiLIP branch is unavailable; it remains outside the image catalog. Neither blocker involved workload execution.`r`n`r`n### Margrethe workflow validation (2026-07-23)
`rnBoth Margrethe paths completed through the catalog CLI, Docker image, embedded manifest, benchmark runner, JSONL validation, and report generation with one measured iteration and zero warmups. Margrethe SKE completed successfully (`78,112,800 ns`); single-thread Margrethe HHE completed successfully (`2,529,476,178 ns`). Each produced one valid record and zero invalid records. The run used warning preflight on Docker Desktop/WSL; invariant/nonstop TSC and performance-governor checks were not verified, so these are functional validation results, not publication-grade performance data.

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

**Estimated overall readiness: approximately 89%.** This is a planning
estimate, not a test-coverage metric. The measurement and reporting foundation
and matrix-v1 container/dispatch slice are complete. Native-host portability
evidence, release publication, and CI remain.
Completion means that every framework in the declared support matrix has a
reproducible image, production manifest, lightweight end-to-end smoke, and
documented limitations.

| Order | Required capability | Current readiness | Remaining work / completion criterion |
|---:|---|---:|---|
| 1 | Shared cross-language measurement contract | **100%** | Schema 1.0.0 and maintained C/C++, Go, Python, and Rust harnesses exist. Future changes must remain backward compatible or version the schema. |
| 2 | Migrate project-owned scheme benchmark boundaries | **95%** | Non-FbT scheme paths are migrated. AES/FbT is explicitly deferred; dependency-owned primitive suites are outside the scheme support matrix unless later promoted. |
| 3 | Stable CLI selection and orchestration | **100%** | `hhe list`, `bench`, and `bench-all` consume the matrix-v2 hierarchy `family -> path -> cipher -> variant -> framework -> backend -> language`, accept `all`, show declared/effective readiness, and execute only ready targets backed by the image catalog. |
| 4 | Schema validation and report generation | **100%** | Catalog runs now standardize output as a `raw/` plus `report/` bundle and automatically invoke schema validation and CSV/LaTeX/pgfplots timing and memory reporting after every selected container succeeds. |
| 5 | Define the supported framework/cipher/backend matrix | **100%** | Frozen v1 remains the runnable four-image slice. Draft `manifests/support-matrix.v2.json` separates 29 build/readiness units from 103 source-audited benchmark variants: 10 AES/conventional, 72 HE-friendly, and 21 pure HE. Native Fasta, HHE Framework plain/SKE, and YuX symmetric are separate from their deferred HE/HHE packages. |
| 6 | Reproducible Docker build for every supported framework | **52%** | Seventeen of 29 v2 candidates now have contract-v1 images. Hippogryph and P-encoding AES/Rust are packaged heavy images; remaining candidates are deferred until their packaging tasks are scheduled. |
| 7 | Production manifests for concrete binaries | **52%** | Seventeen v2 candidates have production manifests and are listed in `manifests/images.v2.json`. Hippogryph adds two jobs, bringing the catalog to 123 runnable operation choices. Heavy choices remain explicitly non-smoke. |
| 8 | End-to-end container workflow | **85%** | All five matrix-v1 symmetric targets completed build/select -> warning preflight -> CPU-pinned process -> JSONL validation -> reports on Windows Docker Desktop. Catalog execution now wires validation/report bundles automatically and gives every language container one stable host identity. Repeat on native Linux. Heavy HE validation still requires explicit authorization. |
| 9 | Host-platform independence | **30%** | Windows Docker Desktop is functionally verified, but its WSL preflight failed publication assumptions. Native Linux, macOS Docker, and ARM remain unvalidated; publication-grade cycle measurements remain limited to compatible bare-metal Linux hosts. |
| 10 | Easy addition and migration workflow | **75%** | Adding-scheme and schema documentation plus shared adapters exist. Add a scaffold/template that creates the Dockerfile, manifest, smoke target, report wiring, and acceptance checklist together. |
| 11 | CI, release artifacts, and reproducibility evidence | **15%** | Add CI that builds each lightweight image target, validates manifests, runs only safe synthetic/plain smokes, and publishes pinned image digests. Record full HE validation separately on suitable hardware. |

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
12. **Deferred by user:** Repeat the lightweight runnable matrix workflow on
   native Linux; do not schedule this until explicitly resumed.
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

The immediate critical path is therefore **heavy image contract packaging
without workload execution -> native Linux portability -> CI and release
evidence**. Migrating more
internal microbenchmarks does not materially advance the ideal goal unless
those microbenchmarks are added to the declared support matrix.

### Matrix-v1 functional smoke evidence

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
