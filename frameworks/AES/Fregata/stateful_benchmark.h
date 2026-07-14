#ifndef HHE_ARENA_FREGATA_STATEFUL_BENCHMARK_H
#define HHE_ARENA_FREGATA_STATEFUL_BENCHMARK_H

#include "utils.h"

static inline void hhe_stateful_benchmark_begin(
    HHEBenchmarkSession *session, const char *operation) {
  HHEBenchmarkConfig config = HHEBenchmarkConfigFromEnv(operation, 1);
  config.language = HHE_CALLER_LANGUAGE;
  config.iterations = 1;
  config.warmups = 0;
  if (HHEBenchmarkBegin(session, &config) != 0) {
    abort();
  }
  HHEBenchmarkIterationBegin(session, HHE_PHASE_MEASURED, 0);
}

static inline void hhe_stateful_benchmark_end(HHEBenchmarkSession *session) {
  if (HHEBenchmarkIterationEnd(session) != 0 ||
      HHEBenchmarkFinish(session) != 0) {
    abort();
  }
}

#endif
