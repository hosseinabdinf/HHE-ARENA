#pragma once

#include <cstdio>
#include <cstdlib>
#include <map>
#include <memory>
#include <stdexcept>
#include <string>

#include <utils.h>

namespace hhe_benchmark {

struct KatOperationState {
  std::string operation;
  HHEBenchmarkConfig config{};
  HHEBenchmarkSession session{};
};

struct KatContext {
  bool running = false;
  bool skipping = false;
  HHEPhase phase = HHE_PHASE_WARMUP;
  size_t index = 0;
  unsigned long long next_session_id = 0;
  KatOperationState* active = nullptr;
  std::map<std::string, std::unique_ptr<KatOperationState>> operations;
};

inline KatContext& kat_context() {
  static thread_local KatContext context;
  return context;
}

inline bool finish_sessions() noexcept {
  KatContext& context = kat_context();
  bool ok = true;
  context.active = nullptr;
  context.skipping = false;
  for (auto& entry : context.operations) {
    if (HHEBenchmarkFinish(&entry.second->session) != 0) ok = false;
  }
  context.operations.clear();
  return ok;
}

inline void operation_begin(const char* operation, const char* file, int line) {
  KatContext& context = kat_context();
  if (!context.running) throw std::logic_error("KAT operation outside run_kat");
  if (context.active != nullptr || context.skipping)
    throw std::logic_error("nested KAT operation");

  const char* requested = std::getenv("HHE_BENCH_OPERATION");
  if (requested != nullptr && requested[0] != '\0' &&
      std::string(requested) != operation) {
    context.skipping = true;
    return;
  }

  const std::string key = std::string(file) + ":" + std::to_string(line);
  auto found = context.operations.find(key);
  if (found == context.operations.end()) {
    auto state = std::make_unique<KatOperationState>();
    state->operation = operation;
    state->config = HHEBenchmarkConfigFromEnv(state->operation.c_str(), 1);
    state->config.language = "cpp";
    if (std::getenv("HHE_BENCH_IMPLEMENTATION") == nullptr)
      state->config.implementation = "hhe-framework";
    if (HHEBenchmarkBegin(&state->session, &state->config) != 0)
      throw std::runtime_error("failed to initialize KAT benchmark session");

    char base_run_id[sizeof(state->session.run_id)];
    std::snprintf(base_run_id, sizeof(base_run_id), "%s",
                  state->session.run_id);
    std::snprintf(state->session.run_id, sizeof(state->session.run_id),
                  "%.40s-%llu", base_run_id, context.next_session_id++);
    found = context.operations.emplace(key, std::move(state)).first;
  }

  context.active = found->second.get();
  HHEBenchmarkIterationBegin(&context.active->session, context.phase,
                             context.index);
}

inline void operation_end() {
  KatContext& context = kat_context();
  if (context.skipping) {
    context.skipping = false;
    return;
  }
  if (context.active == nullptr)
    throw std::logic_error("KAT operation end without begin");
  KatOperationState* active = context.active;
  context.active = nullptr;
  if (HHEBenchmarkIterationEnd(&active->session) != 0)
    throw std::runtime_error("failed to write KAT benchmark record");
}

template <class Function>
bool run_kat(Function&& function) {
  KatContext& context = kat_context();
  if (context.running) throw std::logic_error("nested run_kat call");

  HHEBenchmarkConfig plan = HHEBenchmarkConfigFromEnv("kat_plan", 1);
  context.running = true;
  bool result = true;
  try {
    for (int phase = 0; phase < 2; ++phase) {
      context.phase = phase == 0 ? HHE_PHASE_WARMUP : HHE_PHASE_MEASURED;
      const size_t count = phase == 0 ? plan.warmups : plan.iterations;
      for (size_t index = 0; index < count; ++index) {
        context.index = index;
        result = function() && result;
      }
    }
  } catch (...) {
    finish_sessions();
    context.running = false;
    throw;
  }

  const bool finished = finish_sessions();
  context.running = false;
  if (!finished) throw std::runtime_error("failed to finish KAT sessions");
  return result;
}

}  // namespace hhe_benchmark

#define HHE_KAT_OPERATION_BEGIN(operation) \
  ::hhe_benchmark::operation_begin((operation), __FILE__, __LINE__)
#define HHE_KAT_OPERATION_END() ::hhe_benchmark::operation_end()
