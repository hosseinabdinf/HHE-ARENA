#include <cstdio>
#include <cstdlib>
#include <fstream>
#include <string>

#include <unistd.h>

#include "../ciphers/common/benchmark_kat.h"

int main() {
  const char* retained_output = std::getenv("HHE_KAT_SMOKE_OUTPUT");
  const bool keep_output = retained_output != nullptr;
  const std::string output =
      keep_output ? retained_output
                  : "/tmp/hhe-kat-benchmark-smoke-" +
                        std::to_string(getpid()) + ".jsonl";
  std::remove(output.c_str());
  setenv("HHE_BENCH_CIPHER", "synthetic-kat", 1);
  setenv("HHE_BENCH_BACKEND", "synthetic", 1);
  setenv("HHE_BENCH_IMPLEMENTATION", "hhe-kat-smoke", 1);
  setenv("HHE_BENCH_ITERATIONS", "5", 1);
  setenv("HHE_BENCH_WARMUPS", "1", 1);
  setenv("HHE_BENCH_POLICY", "warn", 1);
  setenv("HHE_BENCH_OUTPUT", output.c_str(), 1);

  int calls = 0;
  const bool ok = hhe_benchmark::run_kat([&]() {
    HHE_KAT_OPERATION_BEGIN("synthetic_operation_a");
    ++calls;
    HHE_KAT_OPERATION_END();
    HHE_KAT_OPERATION_BEGIN("synthetic_operation_b");
    ++calls;
    HHE_KAT_OPERATION_END();
    return true;
  });

  std::ifstream input(output);
  std::string line;
  size_t records = 0;
  while (std::getline(input, line)) {
    if (line.find("\"schema_version\":\"1.0.0\"") == std::string::npos ||
        (line.find("\"operation\":\"synthetic_operation_a\"") ==
             std::string::npos &&
         line.find("\"operation\":\"synthetic_operation_b\"") ==
             std::string::npos) ||
        line.find("\"implementation\":\"hhe-kat-smoke\"") ==
            std::string::npos)
      return 2;
    ++records;
  }
  if (!keep_output) std::remove(output.c_str());
  return ok && calls == 12 && records == 12 ? 0 : 1;
}