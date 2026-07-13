"""Serialized x86 timestamp-counter access through a tiny cffi module."""

from functools import lru_cache
import platform
import statistics


@lru_cache(maxsize=1)
def _library():
    if platform.machine().lower() not in {"x86_64", "amd64"}:
        return None
    try:
        from cffi import FFI
        ffi = FFI()
        ffi.cdef("""
            int hhe_rdtscp_available(void);
            unsigned long long hhe_tsc_start(void);
            unsigned long long hhe_tsc_end(void);
        """)
        return ffi.verify(r"""
            #include <cpuid.h>
            #include <stdint.h>
            #include <x86intrin.h>
            int hhe_rdtscp_available(void) {
                unsigned int a,b,c,d;
                if (__get_cpuid_max(0x80000000, 0) < 0x80000001) return 0;
                __cpuid(0x80000001, a,b,c,d);
                return (d & (1U << 27)) != 0;
            }
            uint64_t hhe_tsc_start(void) {
                unsigned int a,b,c,d;
                __cpuid(0, a,b,c,d);
                return __rdtsc();
            }
            uint64_t hhe_tsc_end(void) {
                unsigned int aux,a,b,c,d;
                uint64_t value = __rdtscp(&aux);
                __cpuid(0, a,b,c,d);
                return value;
            }
        """, extra_compile_args=["-O2"])
    except Exception:
        return None


def available():
    library = _library()
    return library is not None and bool(library.hhe_rdtscp_available())


def start():
    library = _library()
    return int(library.hhe_tsc_start()) if available() else None


def end():
    library = _library()
    return int(library.hhe_tsc_end()) if available() else None


def measure_overhead(samples=1000):
    if samples <= 0 or not available():
        return None
    values = []
    for _ in range(samples):
        before = start()
        after = end()
        values.append(after - before)
    return int(statistics.median(values))
