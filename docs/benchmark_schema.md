# Benchmark JSONL contract

Each UTF-8 JSONL line is one warm-up or measured iteration and validates against `schemas/benchmark.schema.json`. Raw JSONL is the source of truth. Schema version `1.0.0` must be amended here and in the schema before a language port diverges.

`run_id` groups one invocation. Cipher and operation are extensible canonical strings; backend and security profile are nullable. Shared enums cover language, GC mode, allocator, phase, environment policy, and status. Warm-up and measured indices are zero-based within their phase; reports exclude warm-ups.

Successful records use `status: "ok"` and `error: null`; failed records require a non-empty error. Unavailable measurements are `null`, never fabricated as zero. JSONL write failure is reported on stderr with a non-zero exit because the lost line cannot describe itself.

## Schema freeze

Version `1.0.0` is frozen after completion of the Rust reference harness. Tasks 3-5 must port this contract unchanged; a necessary amendment must update both schema and documentation before port work continues.

## Time and cycles

`time_ns` is the exact integer difference between monotonic clock reads around the operation. It is never floating point, milliseconds, a decimal string, or derived from cycles. `cycles` is an unsigned difference between serialized counter reads and is null when unavailable. Both use the same operation boundary.

## Memory semantics

Memory fields are byte deltas computed as `after - before` outside timing. They are signed and must not be clamped; unavailable metrics are null.

- `rss_bytes_delta`: current resident-set change.
- `heap_bytes_delta`: live/allocated heap change reported by the allocator/runtime.
- `peak_rss_bytes_delta`: process high-water RSS change; often zero after an earlier peak.
- 
u_maxrss_bytes: absolute process high-water RSS observed after the operation, normalized to bytes from getrusage(RUSAGE_SELF).ru_maxrss. It is optional so existing schema-1.0.0 records remain valid; updated harnesses emit a value or 
ull.

Rust uses the system allocator on glibc Linux (`system_glibc`). RSS comes from `/proc/self/statm`, heap bytes from `mallinfo2().uordblks`, and peak RSS from `getrusage`. Installing jemalloc while sampling `mallinfo2` is forbidden because glibc would omit jemalloc allocations.

## C/C++ port

The C/C++ harness uses the same system-glibc memory sources and serialized TSC boundary as Rust. Legacy statement macros are retained as iteration-control wrappers around the session API so captured local variables remain valid while every warm-up and measured iteration gets its own JSONL record.

## Go port

Go heap bytes are `runtime.MemStats.Alloc`; RSS and peak RSS retain the Linux meanings above. GC mode is always explicit. The goroutine is locked to its OS thread before optional affinity pinning, verification, and calibration. Serialized amd64 assembly replaces cgo; measured empty-boundary overhead is reported but never subtracted from raw records.

## Python port

Python uses integer `perf_counter_ns` for exact time. Its heap metric is current `tracemalloc`-tracked Python allocation and excludes untraced native allocations; VMS is not heap. RSS and peak RSS retain the process meanings above. A cffi wrapper supplies explicit serialized amd64 counter reads; its measured call boundary is reported without subtraction. GC and tracing state are restored.

## Host identity

`host.hostname` is the stable benchmark-host identifier used when reports
separate complete host identities. Every harness prefers
`HHE_BENCH_HOSTNAME`; otherwise it falls back to the runtime hostname when
available. Catalog-mode CLI runs set this consistently across framework
containers from `--host-id`, whose default is the invoking system hostname.
Users should provide an explicit unique ID on cloned or generically named
hosts. The remaining host fields describe the actual execution environment
inside the container; they are not overwritten with host-OS guesses.

## Nullability and environment

Unknown or unavailable observations use null; false means a check was performed and failed. Policy is `strict` or `warn`. Strict failures prevent measured iterations. Warnings provide stable explanations.

## Minimal example

```json
{"schema_version":"1.0.0","run_id":"01JEXAMPLE","timestamp_utc":"2026-07-12T12:00:00Z","benchmark":{"cipher":"hera","operation":"encrypt","backend":"ckks","security_profile":"HHE-128","parameters":{}},"runtime":{"language":"go","implementation":"example","version":null,"gc_mode":"not_applicable","allocator":"system_glibc"},"host":{"hostname":null,"os":"linux","arch":"x86_64","cpu_model":null,"logical_cpus":null},"environment":{"policy":"warn","requested_cpu":null,"actual_cpu":null,"affinity_verified":null,"rdtscp_supported":true,"invariant_tsc":null,"nonstop_tsc":null,"governor":null,"turbo_enabled":null,"checks_passed":true,"warnings":[]},"iteration":{"phase":"measured","index":0},"measurement":{"time_ns":123456,"cycles":456789,"memory":{"rss_bytes_delta":0,"heap_bytes_delta":-1024,"peak_rss_bytes_delta":0,"ru_maxrss_bytes":12582912}},"status":"ok","error":null}
```


## Reporting

`benchmarking-utilities/report/report_generator.py` validates this contract before aggregation. It excludes warm-ups and groups measured records by cipher, language, operation, backend, security profile, GC mode, and complete host identity. Groups with at most five successful records expose median/min/max but leave quartiles and IQR unavailable; larger groups use Tukey hinges. `report.csv` contains timing and all memory statistics; `memory.csv`, `memory.tex`, and `memory.pgfplots.tex` provide a dedicated memory report. Baseline subtraction applies only to signed memory deltas, never to absolute `ru_maxrss_bytes`. Invalid, failed, duplicate-index, and index-gap conditions are surfaced rather than silently discarded.
