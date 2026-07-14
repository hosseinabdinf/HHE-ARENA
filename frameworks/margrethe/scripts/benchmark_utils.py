"""Schema 1.0.0 reference benchmark harness for Python."""

from dataclasses import dataclass
import gc
import json
import os
import platform
import resource
import statistics
import sys
import time
import tracemalloc
from pathlib import Path
from typing import Optional

import psutil

try:
    from . import serialized_counter
except ImportError:
    import serialized_counter


@dataclass
class BenchmarkConfig:
    cipher: str
    operation: str
    iterations: int
    output: str = "benchmark.jsonl"
    backend: Optional[str] = None
    security_profile: Optional[str] = None
    implementation: str = "python-benchmark"
    warmups: Optional[int] = None
    cpu: Optional[int] = None
    policy: str = "warn"
    gc_mode: str = "unknown"


def run_benchmark(config, func):
    _validate(config)
    process = psutil.Process(os.getpid())
    environment = _prepare_environment(process, config.cpu, config.policy)
    warmups = config.warmups
    if warmups is None:
        warmups = _proportional_warmups(config.iterations)
    run_id = f"{os.getpid()}-{time.time_ns()}"
    counter_overhead = serialized_counter.measure_overhead()

    gc_was_enabled = gc.isenabled()
    tracing_was_enabled = tracemalloc.is_tracing()
    _set_gc(config.gc_mode)
    if not tracing_was_enabled:
        tracemalloc.start()

    measured_ns = []
    try:
        with open(config.output, "a", encoding="utf-8", buffering=1) as output:
            for phase, count in (("warmup", warmups), ("measured", config.iterations)):
                for index in range(count):
                    before = _memory_sample(process)
                    start_time_ns = time.perf_counter_ns()
                    start_cycles = serialized_counter.start()
                    operation_error = None
                    try:
                        func()
                    except BaseException as error:
                        operation_error = error
                    end_cycles = serialized_counter.end()
                    elapsed_ns = time.perf_counter_ns() - start_time_ns
                    after = _memory_sample(process)
                    if phase == "measured":
                        measured_ns.append(elapsed_ns)
                    record = _make_record(
                        config, environment, run_id, phase, index, elapsed_ns,
                        _cycle_delta(start_cycles, end_cycles),
                        _memory_delta(before, after), operation_error,
                    )
                    line = json.dumps(record, separators=(",", ":"), ensure_ascii=False)
                    output.write(line + "\n")
                    output.flush()
                    if operation_error is not None:
                        raise operation_error
    finally:
        if not tracing_was_enabled and tracemalloc.is_tracing():
            tracemalloc.stop()
        if gc_was_enabled:
            gc.enable()
        else:
            gc.disable()

    _print_summary(config.operation, measured_ns, counter_overhead)


def benchmark(name, iterations, func):
    config = config_from_env(name, iterations)
    run_benchmark(config, func)


def config_from_env(operation, iterations):
    requested_iterations = _env_int("HHE_BENCH_ITERATIONS")
    return BenchmarkConfig(
        cipher=os.getenv("HHE_BENCH_CIPHER", "unknown"),
        operation=operation,
        iterations=iterations if requested_iterations is None else requested_iterations,
        output=os.getenv("HHE_BENCH_OUTPUT", "benchmark.jsonl"),
        backend=os.getenv("HHE_BENCH_BACKEND"),
        security_profile=os.getenv("HHE_BENCH_SECURITY_PROFILE"),
        implementation=os.getenv("HHE_BENCH_IMPLEMENTATION", "python-benchmark"),
        warmups=_env_int("HHE_BENCH_WARMUPS"),
        cpu=_env_int("HHE_BENCH_CPU"),
        policy=os.getenv("HHE_BENCH_POLICY", "warn"),
        gc_mode=os.getenv("HHE_BENCH_GC", "unknown"),
    )


def _validate(config):
    if not config.cipher:
        raise ValueError("cipher must not be empty")
    if not config.operation:
        raise ValueError("operation must not be empty")
    if config.iterations <= 0:
        raise ValueError("iterations must be positive")
    if config.warmups is not None and config.warmups < 0:
        raise ValueError("warmups must not be negative")
    if config.policy not in {"strict", "warn"}:
        raise ValueError(f"invalid policy {config.policy!r}")
    if config.gc_mode not in {"enabled", "disabled", "unknown"}:
        raise ValueError(f"invalid GC mode {config.gc_mode!r}")


def _prepare_environment(process, requested_cpu, policy):
    warnings = []
    affinity_verified = None
    if requested_cpu is not None:
        try:
            process.cpu_affinity([requested_cpu])
            affinity_verified = process.cpu_affinity() == [requested_cpu]
        except (AttributeError, OSError, ValueError) as error:
            raise RuntimeError(f"CPU affinity failed: {error}") from error
    try:
        actual_cpu = process.cpu_num()
    except (AttributeError, OSError):
        actual_cpu = None
    if requested_cpu is not None and actual_cpu is not None:
        affinity_verified = affinity_verified and actual_cpu == requested_cpu

    rdtscp = serialized_counter.available()
    invariant_tsc = _cpu_flag("constant_tsc")
    nonstop_tsc = _cpu_flag("nonstop_tsc")
    governor = _read_text("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
    turbo_enabled = _turbo_enabled()
    if not rdtscp:
        warnings.append("RDTSCP unavailable; cycles are null")
    if invariant_tsc is not True:
        warnings.append("invariant TSC not verified")
    if nonstop_tsc is not True:
        warnings.append("nonstop TSC not verified")
    if governor != "performance":
        warnings.append("governor not performance")
    if turbo_enabled is not False:
        warnings.append("Turbo enabled or unavailable")
    if affinity_verified is False:

        warnings.append("CPU affinity verification failed")
    checks_passed = not warnings
    if policy == "strict" and not checks_passed:
        raise RuntimeError("; ".join(warnings))
    return {
        "policy": policy,
        "requested_cpu": requested_cpu,
        "actual_cpu": actual_cpu,
        "affinity_verified": affinity_verified,
        "rdtscp_supported": rdtscp,
        "invariant_tsc": invariant_tsc,
        "nonstop_tsc": nonstop_tsc,
        "governor": governor,
        "turbo_enabled": turbo_enabled,
        "checks_passed": checks_passed,
        "warnings": warnings,
    }


def _set_gc(mode):
    if mode == "enabled":
        gc.enable()
    elif mode == "disabled":
        gc.disable()


def _memory_sample(process):
    try:
        rss = int(process.memory_info().rss)
    except (psutil.Error, OSError):
        rss = None
    try:
        peak = int(resource.getrusage(resource.RUSAGE_SELF).ru_maxrss)
        if not sys.platform.startswith("darwin"):
            peak *= 1024
    except (ValueError, OSError):
        peak = None
    try:
        heap = int(tracemalloc.get_traced_memory()[0])
    except RuntimeError:
        heap = None
    return rss, heap, peak


def _memory_delta(before, after):
    return {
        "rss_bytes_delta": _signed_delta(before[0], after[0]),
        "heap_bytes_delta": _signed_delta(before[1], after[1]),
        "peak_rss_bytes_delta": _signed_delta(before[2], after[2]),
        "ru_maxrss_bytes": after[2],
    }


def _signed_delta(before, after):
    if before is None or after is None:
        return None
    return int(after) - int(before)


def _cycle_delta(start, end):
    if start is None or end is None or end < start:
        return None
    return end - start


def _hostname():
    hostname = os.getenv("HHE_BENCH_HOSTNAME")
    if hostname:
        return hostname
    try:
        return platform.node() or None
    except OSError:
        return None


def _make_record(config, environment, run_id, phase, index, elapsed_ns,
                 cycles, memory, operation_error):
    error_text = None if operation_error is None else f"{type(operation_error).__name__}: {operation_error}"
    hostname = _hostname()
    return {
        "schema_version": "1.0.0",
        "run_id": run_id,
        "timestamp_utc": _timestamp_utc(),
        "benchmark": {
            "cipher": config.cipher,
            "operation": config.operation,
            "backend": config.backend,
            "security_profile": config.security_profile,
            "parameters": {},
        },
        "runtime": {
            "language": "python",
            "implementation": config.implementation,
            "version": platform.python_version(),
            "gc_mode": config.gc_mode,
            "allocator": "python_runtime",
        },
        "host": {
            "hostname": hostname,
            "os": platform.system().lower() or sys.platform,
            "arch": platform.machine() or "unknown",
            "cpu_model": _cpu_model(),
            "logical_cpus": os.cpu_count(),
        },
        "environment": environment,
        "iteration": {"phase": phase, "index": index},
        "measurement": {
            "time_ns": elapsed_ns,
            "cycles": cycles,
            "memory": memory,
        },
        "status": "ok" if operation_error is None else "failed",
        "error": error_text,
    }


def _timestamp_utc():
    seconds, nanos = divmod(time.time_ns(), 1_000_000_000)
    base = time.strftime("%Y-%m-%dT%H:%M:%S", time.gmtime(seconds))
    return f"{base}.{nanos:09d}Z"


def _proportional_warmups(iterations):
    return min(10, max(1, (iterations + 9) // 10))


def _print_summary(name, values, counter_overhead):
    if not values:
        return
    ordered = sorted(values)
    print(
        f"-> {name}: n={len(ordered)}, median={int(statistics.median(ordered))} ns, "
        f"min={ordered[0]} ns, max={ordered[-1]} ns, "
        f"counter-boundary={counter_overhead if counter_overhead is not None else 'null'} cycles"
    )


def _read_text(path):
    try:
        return Path(path).read_text(encoding="utf-8").strip()
    except OSError:
        return None


def _cpu_flag(flag):
    text = _read_text("/proc/cpuinfo")
    if text is None:
        return None
    return any(
        flag in line.split()
        for line in text.splitlines()
        if line.startswith("flags")
    )


def _cpu_model():
    text = _read_text("/proc/cpuinfo")
    if text is None:
        return None
    for line in text.splitlines():
        if line.startswith("model name") and ":" in line:
            return line.split(":", 1)[1].strip()
    return None


def _turbo_enabled():
    value = _read_text("/sys/devices/system/cpu/intel_pstate/no_turbo")
    if value == "0":
        return True
    if value == "1":
        return False
    return None


def _env_int(name):
    value = os.getenv(name)
    if value is None:
        return None
    try:
        return int(value)
    except ValueError as error:
        raise ValueError(f"{name} must be an integer") from error


def print_header(header):
    print(f"\n ------------------===|\t {header} \t|===------------------")


def print_status(message):
    print(f"--> {message}", end="", flush=True)


def print_message(message):
    print(f"--> {message}")
