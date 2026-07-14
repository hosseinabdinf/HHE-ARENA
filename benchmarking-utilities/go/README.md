# Go Benchmarking Utility

This package ports frozen benchmark schema `1.0.0`.

## Measurement contract

`Run(Config, func())` locks the goroutine to its OS thread, optionally pins
that thread to a Linux CPU, verifies the active CPU, and calibrates only after
pinning. On amd64, assembly implements serialized `CPUID/RDTSC` and
`RDTSCP/CPUID`; CPUID detects RDTSCP before use. This removes the former cgo
call boundary.

`MeasureCounterOverhead` reports the median empty serialized counter boundary.
The harness measures it once per run and prints it with the summary; it is not
subtracted from raw cycles. Time, memory sampling, JSON encoding, and output are
kept outside the target operation as required by the shared contract.

Warm-ups default to 10% of measured iterations, clamped to 1-10. GC mode is
`enabled`, `disabled`, or `unknown`; changed GC percentages are restored
when the run ends. Signed RSS, Go heap allocation, and peak-RSS byte deltas are
emitted per iteration.

Strict policy stops before measurement on environment failure. Warn policy
records warnings and continues. Output failures are returned by `Run`; legacy
`Benchmark` and `BenchmarkIter` wrappers panic so failures cannot be silent.

## Environment variables

Legacy wrappers read `HHE_BENCH_OUTPUT`, `HHE_BENCH_CIPHER`,
`HHE_BENCH_BACKEND`, `HHE_BENCH_SECURITY_PROFILE`,
`HHE_BENCH_IMPLEMENTATION`, `HHE_BENCH_CPU`, `HHE_BENCH_WARMUPS`,
`HHE_BENCH_GC`, `HHE_BENCH_POLICY`, and the stable host override
`HHE_BENCH_HOSTNAME`.

See `../../docs/benchmark_schema.md` for schema and memory semantics.
