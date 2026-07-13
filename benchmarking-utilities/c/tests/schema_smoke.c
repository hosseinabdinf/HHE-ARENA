#include "../utils/utils.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void increment(void* context) { ++*(int*)context; }

int main(void) {
  const char* path = "hhe-c-schema-smoke.jsonl";
  HHEBenchmarkConfig config = HHEBenchmarkConfigFromEnv("smoke", 2);
  config.cipher = "test";
  config.implementation = "c-schema-smoke";
  config.output_path = path;
  config.warmups = 1;
  config.cpu = -1;
  config.policy = HHE_POLICY_WARN;
  int count = 0;
  if (HHEBenchmarkRun(&config, increment, &count) != 0 || count != 3) return 1;
  FILE* file = fopen(path, "r");
  if (!file) return 2;
  char line[8192];
  int lines = 0;
  while (fgets(line, sizeof(line), file)) {
    if (!strstr(line, "\"schema_version\":\"1.0.0\"") ||
        !strstr(line, "\"time_ns\":") || !strstr(line, "\"memory\":") ||
        !strstr(line, "\"ru_maxrss_bytes\":"))
      return 3;
    ++lines;
  }
  fclose(file);
  remove(path);
  return lines == 3 ? 0 : 4;
}
