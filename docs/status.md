# Benchmark utility migration status

Last audited: 2026-07-13

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
| AES / Fregata CBS `homoSM4_CB` | AES symmetric operations and TFHE++ transciphering | **New** | Active new macros in `sym_aes.cpp` and `hom_aes.cpp`. |
| AES / Fregata other CBS benchmarks, FBS, GBS | TFHE++ primitives, SM4/Fregata modes | **Old / not integrated** | CBS `benchmark/` uses Google Benchmark; no schema call sites were found in FBS or GBS. Framework is therefore **mixed** overall. |
| AES / FbT | AES over OpenFHE | **Mixed** | The active key-scheduler measurement uses the new macro, but transcipherer setup and `HHE.Decomp()` measurements are commented out in `example/example_aes.cpp`. |
| AES / BtR (Lattigo) | BFV/BGV/CKKS and multiparty HE; CKKS cipher area | **Utility only / old** | A migrated Go utility exists in `ckks_cipher`, but no active call site invokes it. Existing Lattigo benchmarks use Go `testing.B`. |
| AES / FFT-CBS | AES, AES half-CBS, circuit bootstrapping, HP-LHE | **Mixed** | Active AES test call sites use the shared Rust schema harness; separate Criterion suites remain outside JSONL. |
| AES / Hippogryph | Hippogryph AES / odd-modulus TFHE | **New** | Active AES measurements use a thin facade over the shared Rust schema harness. |
| AES / P-encoding | P-encoding AES / odd-modulus TFHE | **New** | Active AES measurements use a thin facade over the shared Rust schema harness. |
| Elisabeth | Elisabeth clear and homomorphic paths | **Mixed** | Clear/homomorphic tests use the shared Rust schema harness; the separate Criterion benchmark remains. |
| Fasta | Fasta symmetric and HElib; Rasta HElib, packed HElib, and TFHE | **New** | All five active executables use schema-1.0.0 sessions. Reset/cleanup hooks keep input reconstruction and disposal outside timing; legacy sources are retained as sibling `_backup` files. |
| FiLIP / independent-PTX | FiLIP p-dependent setup, SKE encryption, HHE decomposition | **New** | Active `BENCHMARK_ITER` call sites use migrated C utilities. |
| FiLIP / Sorting Hat | FiLIP SKE encryption and HHE decomposition | **New** | Active `BENCHMARK_ITER` call sites use migrated C utilities. |
| FiLIP / Rust reference | FiLIP | **Old** | Criterion benchmark remains; README describes the implementation as archival because an upstream dependency is unavailable. |
| FRAST | FRAST setup, online phase, PBS generation, full benchmark | **Mixed** | Active test measurements use the shared Rust schema harness; separate Criterion benches remain. |
| HHE Framework | Agrasta, Dasta 5/6, FiLIP 1216/1280, Hera 5, Kreyvium variants, LowMC, Masta 4/5, Pasta/Pasta2 3/4, Rasta 5/6 and packed variants, HE-only paths | **New** | Seven common KAT headers expose 73 schema boundaries to 80 generated cipher/backend test-vector executables through the shared C/C++ harness. |
| HHELand / dedicated `TestNb*` paths | Hera, Pasta3, Pasta4, Rubato SKE and HHE decomposition; pure BGV and CKKS operations | **New** | Active `utils.Benchmark`/`BenchmarkIter` call sites use the migrated Go harness and honor CLI iteration overrides. |
| HHELand / standard benchmark suites | Hera, Pasta, Rubato, RtF integration, ring primitives | **Old** | `*_benchmark_test.go` and `RtF_bench_test.go` use Go `testing.B`. HHELand is therefore **mixed** overall. |
| Margrethe | Margrethe C HHE decomposition and Python SKE | **New** | C `BENCHMARK_ITER` and Python `benchmark(...)` call sites both use migrated harnesses. |
| Transistor / symmetric | Transistor symmetric clock/encryption | **New** | `symmetric/main.py` uses the migrated Python harness. |
| Transistor / homomorphic | Homomorphic Transistor / odd TFHE | **New** | The Rust implementation uses a thin facade over the shared schema harness. Combined with the symmetric path, the framework benchmark helpers are new. |
| Trivium/Kreyvium (`trkr`) | Trivium and Kreyvium TFHE-rs application paths | **Mixed** | Application test measurements use the shared Rust schema harness; separate Criterion benches remain outside JSONL. |
| YuX | Yu2x-8, Yu2x-16, Yupx-p; symmetric, key schedule, HE key encryption, transciphering and final HE decrypt operations | **New** | Test executables actively use migrated C/C++ `BENCHMARK` macros and link the new utility. |

## Framework details

### AES family

**New paths**

- OpenSSL AES and TinyAES provide reference AES encryption measurements through `BENCHMARK_ITER`.
- Full-LUT uses the new utility for symmetric key expansion, TFHE symmetric encryption, and transciphering.
- Fregata CBS `homoSM4_CB` uses the new utility for AES key expansion, symmetric encryption, plaintext encryption, table construction, and AES transciphering.

**Mixed or old paths**

- FbT has a migrated utility, but only the key-scheduler call is active; the intended HHE decomposition call is commented.
- BtR has a migrated Go utility under `ckks_cipher`, but the audit found no call site using it. Lattigo's existing BFV/BGV/CKKS and multiparty benchmarks still use native Go benchmark APIs.
- FFT-CBS, Hippogryph, and P-encoding active helper call sites now route through the shared Rust schema harness. FFT-CBS remains mixed because its separate Criterion suites are not schema-integrated.
- Fregata's separate Google Benchmark targets and its FBS/GBS components are not migrated by the `homoSM4_CB` utility.

### Fasta and Rasta

Fasta and Rasta are migrated to the C/C++ schema-1.0.0 session API. Active operations are:

- `sym/fasta_sym.c`: `ske_encrypt` and `ske_decrypt`;
- `hhe/fasta/fasta_helib.cpp`: `he_context_setup`, `he_key_switch_setup`, and `hhe_transcipher`;
- `hhe/rasta/rasta_helib.cpp`: the same setup/transcipher operations plus `cipher_matrix_setup`;
- `hhe/rasta_pack/rasta_pack_helib.cpp`: context setup, key-switch setup, and packed transciphering;
- `hhe/rasta_tfhe/rasta_tfhe.c`: TFHE context setup and Rasta transciphering.

The Fasta-local utility is now the shared C schema implementation with a local `HHE_BENCHMARK_ITER_HOOKS` wrapper. Its `before` and `after` hooks run outside the timer so mutable ciphertexts, HE contexts, keys, matrices, and temporary output can be reset or released without contaminating the measured boundary. Original sources and utility/CMake files are preserved beside the active files with `_backup` in their names.

Verification includes Docker executions of symmetric Fasta with one warmup and three measured iterations per operation. The `fasta-sym-rumaxrss.jsonl` run validated all eight records with zero issues and generated timing plus dedicated memory artifacts under `results/fasta/report-rumaxrss/`; median absolute `ru_maxrss_bytes` was 13,111,296 bytes for both operations. The utility and complete symmetric executable also compile/link under pinned GCC 14.2 with `-Werror`; active sources contain no direct legacy timing calls. HElib and TFHE executables were not run because their third-party dependencies are not present locally.

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

### HHELand

The recommended new-harness entry points are the dedicated `TestNb*` tests:

- `sym/hera/hera_test.go`: Hera `SKE.Enc()`;
- `sym/pasta/pasta_test.go`: Pasta3/Pasta4 `SKE.Enc()`;
- `sym/rubato/rubato_test.go`: Rubato `SKE.Enc()`;
- `hhe/hera/hera_test.go`: Hera `HHE.Decomp()`;
- `hhe/pasta/pasta_test.go`: Pasta3/Pasta4 `HHE.Decomp()`;
- `hhe/rubato/rubato_test.go`: Rubato `HHE.Decomp()`;
- `he/pure_bgv.go` and `he/pure_ckks.go`: HE encryption/evaluation operations.

The standard Go benchmark files are separate legacy paths and should not be confused with these `TestNb*` entry points.

### Rust-based frameworks

Elisabeth, FRAST, FFT-CBS, Hippogryph, P-encoding, homomorphic Transistor, and `trkr` now depend on `benchmarking-utilities/rust` through relative Cargo path dependencies. Their existing helper modules are thin compatibility facades, so active `benchmark(name, iterations, closure)` call sites emit schema-1.0.0 JSONL without duplicating timing code. The original seven helpers are retained as sibling `benchmark_utils_backup.rs` files; FFT-CBS also retains its original standalone bench helper as `benches/utils_backup.rs.txt`.

The shared wrapper honors `HHE_BENCH_OPERATION`, `HHE_BENCH_IMPLEMENTATION`, and `HHE_BENCH_WARMUPS` in addition to the existing common controls. A synthetic smoke test verified one warm-up plus two measured records and the requested cipher, operation, and implementation identity. No HE operation was executed.

Criterion remains a separate measurement system in Elisabeth, FRAST, FFT-CBS, and `trkr`; those framework rows are therefore mixed rather than fully new.

## CLI and manifest readiness

Benchmark-harness migration and CLI readiness are separate:

- A **new** call site can consume the common environment and emit JSONL.
- The `hhe` CLI additionally needs a prebuilt executable and a concrete version-1 manifest entry.
- The root tooling image deliberately ships an empty manifest and no framework HE binaries.
- `manifests/example.json` is illustrative, not a production sweep manifest.

Therefore, none of the framework rows should be interpreted as automatically runnable from the base Docker image. Framework-specific binaries and validated manifests still need to be built or supplied.

## Recommended migration order

Fasta/Rasta, the shared framework-local Rust helper group, and the HHE Framework common KAT/timing layer were completed on 2026-07-13. Remaining order:

1. Complete partially migrated AES paths: FbT, BtR, and remaining Fregata modes/benchmarks.
2. Convert HHELand `testing.B` suites only when they are intended as paper-source measurements; otherwise document the `TestNb*` JSONL paths as canonical.

## How to update this document

When a path is migrated:

1. verify an active cipher call site invokes the new harness;
2. use a trivial or synthetic smoke path to confirm schema-valid JSONL without running the expensive HE workload;
3. add or validate its manifest entry separately;
4. update the row and audit date here;
5. do not mark an entire framework new while legacy benchmark executables remain.
