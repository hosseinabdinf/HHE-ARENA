#ifndef HHE_BENCHMARK_UTILS_H
#define HHE_BENCHMARK_UTILS_H

#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <time.h>

#ifdef __cplusplus
#define HHE_CALLER_LANGUAGE "cpp"
#else
#define HHE_CALLER_LANGUAGE "c"
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef enum { HHE_POLICY_STRICT, HHE_POLICY_WARN } HHEPolicy;
typedef enum { HHE_PHASE_WARMUP, HHE_PHASE_MEASURED } HHEPhase;

typedef struct {
  int64_t rss_bytes;
  int64_t heap_bytes;
  int64_t peak_rss_bytes;
  int rss_valid;
  int heap_valid;
  int peak_valid;
} HHEMemorySample;

typedef struct {
  const char* cipher;
  const char* operation;
  const char* backend;
  const char* security_profile;
  const char* implementation;
  const char* language;
  const char* output_path;
  size_t iterations;
  size_t warmups;
  int cpu;
  HHEPolicy policy;
} HHEBenchmarkConfig;

typedef struct {
  HHEBenchmarkConfig config;
  FILE* output;
  char run_id[64];
  char governor[64];
  char warnings[512];
  int actual_cpu;
  int affinity_verified;
  int rdtscp_supported;
  int invariant_tsc;
  int nonstop_tsc;
  int turbo_enabled;
  int checks_passed;
  HHEPhase phase;
  size_t index;
  struct timespec start_time;
  uint64_t start_cycles;
  int cycles_valid;
  HHEMemorySample before;
  uint64_t* measured_ns;
  size_t measured_count;
  int failed;
} HHEBenchmarkSession;

typedef void (*HHEBenchmarkFn)(void* context);

HHEBenchmarkConfig HHEBenchmarkConfigFromEnv(const char* operation,
                                               size_t iterations);
int HHEBenchmarkBegin(HHEBenchmarkSession* session,
                      const HHEBenchmarkConfig* config);
void HHEBenchmarkIterationBegin(HHEBenchmarkSession* session, HHEPhase phase,
                                size_t index);
int HHEBenchmarkIterationEnd(HHEBenchmarkSession* session);
int HHEBenchmarkFinish(HHEBenchmarkSession* session);
int HHEBenchmarkRun(const HHEBenchmarkConfig* config, HHEBenchmarkFn function,
                    void* context);

void PrintHeader(const char* header);
void GetMemoryUsage(long* rss_kib, long* heap_kib, long* peak_rss_kib);
void PrintTimeNS(struct timespec start, struct timespec end);
void PrintCycles(uint64_t start, uint64_t end);
void PrintMemory(const char* name, long start, long end);

#ifdef __cplusplus
}
#endif

/*
 * Compatibility choice: legacy statement macros remain wrappers around the
 * session API because C/C++ call sites capture locals. The session API owns
 * warm-up/measured separation, per-iteration timing, JSONL, and summaries.
 */
#define HHE_BENCHMARK_ITER(name, count_arg, code)                           \
  do {                                                                       \
    HHEBenchmarkConfig _hhe_cfg =                                             \
        HHEBenchmarkConfigFromEnv((name), (size_t)(count_arg));              \
    _hhe_cfg.language = HHE_CALLER_LANGUAGE;              \
    HHEBenchmarkSession _hhe_session;                                         \
    if (HHEBenchmarkBegin(&_hhe_session, &_hhe_cfg) != 0) abort();              \
    for (int _hhe_phase = 0; _hhe_phase < 2 && !_hhe_session.failed;          \
         ++_hhe_phase) {                                                      \
      size_t _hhe_count = _hhe_phase == 0 ? _hhe_cfg.warmups                 \
                                           : _hhe_cfg.iterations;             \
      for (size_t _hhe_i = 0; _hhe_i < _hhe_count; ++_hhe_i) {               \
        HHEBenchmarkIterationBegin(                                           \
            &_hhe_session, _hhe_phase == 0 ? HHE_PHASE_WARMUP                \
                                            : HHE_PHASE_MEASURED,             \
            _hhe_i);                                                          \
        do { code; } while (0);                                               \
        if (HHEBenchmarkIterationEnd(&_hhe_session) != 0) break;              \
      }                                                                       \
    }                                                                         \
    if (HHEBenchmarkFinish(&_hhe_session) != 0)                               \
      abort();                     \
  } while (0)

#define BENCHMARK(name, ...) HHE_BENCHMARK_ITER((name), 1, __VA_ARGS__)
#define BENCHMARK_ITER(name, count_arg, ...)                                 \
  HHE_BENCHMARK_ITER((name), (count_arg), __VA_ARGS__)


/* Fasta/Rasta compatibility wrapper: reset and cleanup run outside timing. */
#define HHE_BENCHMARK_ITER_HOOKS(name, count_arg, before, code, after)       \
  do {                                                                       \
    HHEBenchmarkConfig _hhe_cfg =                                             \
        HHEBenchmarkConfigFromEnv((name), (size_t)(count_arg));              \
    _hhe_cfg.language = HHE_CALLER_LANGUAGE;                                 \
    HHEBenchmarkSession _hhe_session;                                         \
    if (HHEBenchmarkBegin(&_hhe_session, &_hhe_cfg) != 0) abort();           \
    for (int _hhe_phase = 0; _hhe_phase < 2 && !_hhe_session.failed;         \
         ++_hhe_phase) {                                                      \
      size_t _hhe_count = _hhe_phase == 0 ? _hhe_cfg.warmups                 \
                                           : _hhe_cfg.iterations;             \
      for (size_t _hhe_i = 0; _hhe_i < _hhe_count; ++_hhe_i) {               \
        do { before; } while (0);                                             \
        HHEBenchmarkIterationBegin(                                           \
            &_hhe_session, _hhe_phase == 0 ? HHE_PHASE_WARMUP                \
                                            : HHE_PHASE_MEASURED,             \
            _hhe_i);                                                          \
        do { code; } while (0);                                               \
        int _hhe_iteration_result = HHEBenchmarkIterationEnd(&_hhe_session);  \
        do { after; } while (0);                                              \
        if (_hhe_iteration_result != 0) break;                                \
      }                                                                       \
    }                                                                         \
    if (HHEBenchmarkFinish(&_hhe_session) != 0) abort();                      \
  } while (0)

#endif
