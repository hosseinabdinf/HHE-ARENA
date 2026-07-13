# C/C++ Benchmarking Utility

This is the C/C++ port of frozen benchmark schema `1.0.0`.

## API

`HHEBenchmarkRun` accepts a callback and context. For existing call sites,
`BENCHMARK` and `BENCHMARK_ITER` remain statement macros because C and C++
bodies commonly capture local variables. The macros now wrap the session API;
they no longer perform aggregate-only timing.

The harness:

- pins and verifies an optional Linux CPU;
- detects RDTSCP through CPUID;
- uses `CPUID/RDTSC` and `RDTSCP/CPUID`;
- checks invariant/nonstop TSC, governor, Turbo, and affinity;
- supports strict and warn policies;
- uses proportional warm-ups (10%, clamped to 1-10);
- samples memory and writes output outside timing;
- appends one JSONL record per warm-up/measured iteration;
- preserves signed byte deltas and treats output failure as fatal;
- prints median/min/max for measured iterations.

Configuration uses the `HHE_BENCH_*` environment variables documented by the
Rust utility. CMake requires C11. The schema smoke test is registered through
CTest when `BUILD_TESTING=ON`.

## Integration

```cmake
add_subdirectory(utils)
target_link_libraries(your_target PRIVATE utils)
```

See `../../docs/benchmark_schema.md` for field semantics.
