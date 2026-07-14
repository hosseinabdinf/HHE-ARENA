# Rust Benchmarking Utility

This crate is the reference implementation of the HHE-ARENA benchmark contract
in `../../schemas/benchmark.schema.json`.

## Reference API

Create a `BenchmarkConfig` and call `run_benchmark`. The harness pins and
verifies the requested Linux CPU before one-time calibration, performs
proportional warm-ups (10% of measured iterations, clamped to 1-10), samples
memory outside timing, and appends one JSONL record per iteration.

Strict policy aborts before measurement if required environment checks fail.
Warn policy records warnings and continues. JSONL serialization, append, or
flush failures are returned to the caller.

The legacy `benchmark(name, iterations, closure)` API remains available.
Framework-local compatibility facades call `benchmark_with_implementation` so
the consuming Cargo package is recorded instead of the utility crate. Both
entry points honor `HHE_BENCH_OUTPUT`, `HHE_BENCH_CIPHER`,
`HHE_BENCH_OPERATION`, `HHE_BENCH_IMPLEMENTATION`, `HHE_BENCH_ITERATIONS`,
`HHE_BENCH_WARMUPS`, `HHE_BENCH_BACKEND`, `HHE_BENCH_SECURITY_PROFILE`,
`HHE_BENCH_CPU`, `HHE_BENCH_POLICY`, and `HHE_BENCH_HOSTNAME`. The latter
overrides the runtime hostname for stable cross-container grouping. Output defaults to
`benchmark.jsonl`.

## Allocator contract

Rust uses the system allocator on glibc Linux. Heap sampling uses
`mallinfo2().uordblks`; do not install jemalloc unless sampling is changed to
jemalloc's statistics API. See `../../docs/benchmark_schema.md`.
