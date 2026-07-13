# Python Benchmarking Utility

This package ports frozen benchmark schema `1.0.0`.

## Measurement contract

`run_benchmark(BenchmarkConfig, callable)` adds Linux affinity and verification,
strict/warn environment checks, proportional warm-ups, explicit GC modes,
per-iteration JSONL, signed byte deltas, and median/min/max summaries. The
legacy `benchmark(name, iterations, callable)` wrapper reads the shared
`HHE_BENCH_*` environment variables.

Exact `time_ns` comes from integer `time.perf_counter_ns()`. Process/RSS
sampling, `resource.getrusage`, Python heap sampling, JSON encoding, logging,
and file output occur outside the target operation.

## Cycles

The former `hwcounter` dependency is removed because serialized reads could
not be established from repository evidence. `serialized_counter.py` uses
cffi to build a tiny amd64 wrapper with explicit `CPUID/RDTSC` and
`RDTSCP/CPUID`, guarded by CPUID feature detection. Its median empty call
boundary is reported but never subtracted from raw cycles. Unsupported or
unbuildable counters emit `null`.

## Memory and GC

RSS is current process resident memory. Peak RSS is normalized to bytes.
`heap_bytes_delta` is Python-traced allocation from `tracemalloc`; it excludes
native allocations made outside Python's traced allocator. VMS is never reported
as heap. Tracing and the prior GC enabled/disabled state are restored after the
run, including exceptional exits.

Dependencies are `psutil` and `cffi`. See
`../../docs/benchmark_schema.md` for the shared schema.
