# Criterion compatibility adapter

This private crate implements only the Criterion API surface used by the
project-owned FRAST, FFT-CBS, and `trkr` benchmark binaries. Each
`bench_function` or `bench_with_input` call delegates to
`hhe_benchmark::benchmark_with_implementation` and emits schema-1.0.0 JSONL.

The adapter is intentionally not Criterion's adaptive statistics engine.
Heavy operations default to one measured iteration plus the shared harness
warm-up policy; `HHE_BENCH_ITERATIONS` and `HHE_BENCH_WARMUPS` remain the
authoritative overrides. Benchmark groups are encoded in operation names as
`group/id`.

Do not extend this into a general Criterion replacement. Add only API methods
required by a project-owned benchmark migration and keep the compile-only
example in `examples/api_compile.rs` representative.
