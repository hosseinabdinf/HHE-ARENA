#define _GNU_SOURCE
#include "utils.h"

#include <cpuid.h>
#include <errno.h>
#include <malloc.h>
#include <sched.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>
#include <x86intrin.h>

static int read_text(const char* path, char* out, size_t size) {
  FILE* file = fopen(path, "r");
  if (!file) return 0;
  if (!fgets(out, (int)size, file)) { fclose(file); return 0; }
  fclose(file);
  out[strcspn(out, "\r\n")] = 0;
  return 1;
}
static int cpu_flag(const char* wanted) {
  FILE* file = fopen("/proc/cpuinfo", "r");
  char line[4096];
  if (!file) return -1;
  while (fgets(line, sizeof(line), file))
    if (!strncmp(line, "flags", 5) && strstr(line, wanted)) { fclose(file); return 1; }
  fclose(file); return 0;
}
static int detect_rdtscp(void) {
  unsigned int eax, ebx, ecx, edx;
  if (__get_cpuid_max(0x80000000, NULL) < 0x80000001) return 0;
  __cpuid(0x80000001, eax, ebx, ecx, edx);
  return (edx & (1U << 27)) != 0;
}
static uint64_t tsc_start(void) {
  unsigned int eax, ebx, ecx, edx;
  __cpuid(0, eax, ebx, ecx, edx);
  return __rdtsc();
}
static uint64_t tsc_end(void) {
  unsigned int aux, eax, ebx, ecx, edx;
  uint64_t value = __rdtscp(&aux);
  __cpuid(0, eax, ebx, ecx, edx);
  return value;
}
static int64_t elapsed_ns(struct timespec start, struct timespec end) {
  return (int64_t)(end.tv_sec - start.tv_sec) * 1000000000LL +
         (int64_t)(end.tv_nsec - start.tv_nsec);
}
static size_t default_warmups(size_t iterations) {
  size_t value = (iterations + 9) / 10;
  if (value < 1) value = 1;
  return value > 10 ? 10 : value;
}
static void append_warning(HHEBenchmarkSession* s, const char* warning) {
  size_t used = strlen(s->warnings);
  if (used && used + 2 < sizeof(s->warnings)) strcat(s->warnings, "; ");
  strncat(s->warnings, warning, sizeof(s->warnings) - strlen(s->warnings) - 1);
}
static HHEMemorySample memory_sample(void) {
  HHEMemorySample sample = {0};
  FILE* file = fopen("/proc/self/statm", "r");
  long pages;
  if (file) {
    if (fscanf(file, "%*s %ld", &pages) == 1) {
      sample.rss_bytes = (int64_t)pages * sysconf(_SC_PAGESIZE);
      sample.rss_valid = 1;
    }
    fclose(file);
  }
  struct mallinfo2 info = mallinfo2();
  sample.heap_bytes = (int64_t)info.uordblks;
  sample.heap_valid = 1;
  struct rusage usage;
  if (getrusage(RUSAGE_SELF, &usage) == 0) {
    sample.peak_rss_bytes = (int64_t)usage.ru_maxrss * 1024;
    sample.peak_valid = 1;
  }
  return sample;
}
static void json_string(FILE* out, const char* value) {
  if (!value) { fputs("null", out); return; }
  fputc('"', out);
  for (const unsigned char* p=(const unsigned char*)value; *p; ++p) {
    switch (*p) {
      case '"': fputs("\\\"",out); break; case '\\': fputs("\\\\",out); break;
      case '\n': fputs("\\n",out); break; case '\r': fputs("\\r",out); break;
      case '\t': fputs("\\t",out); break;
      default: if (*p < 0x20) fprintf(out,"\\u%04x",*p); else fputc(*p,out);
    }
  }
  fputc('"', out);
}
static void nullable_bool(FILE* out, int value) {
  if (value < 0) fputs("null",out); else fputs(value ? "true":"false",out);
}
static void nullable_i64(FILE* out, int valid, int64_t value) {
  if (valid) fprintf(out,"%lld",(long long)value); else fputs("null",out);
}
static void timestamp(char out[32]) {
  time_t now=time(NULL); struct tm value;
  gmtime_r(&now,&value); strftime(out,32,"%Y-%m-%dT%H:%M:%SZ",&value);
}
static void cpu_model(char out[256]) {
  FILE* f=fopen("/proc/cpuinfo","r"); char line[512]; out[0]=0;
  if (!f) return;
  while(fgets(line,sizeof(line),f)) if(!strncmp(line,"model name",10)) {
    char* colon=strchr(line,':'); if(colon){strncpy(out,colon+2,255);out[255]=0;out[strcspn(out,"\r\n")]=0;} break;
  }
  fclose(f);
}

HHEBenchmarkConfig HHEBenchmarkConfigFromEnv(const char* operation, size_t iterations) {
  HHEBenchmarkConfig c={0};
  c.cipher=getenv("HHE_BENCH_CIPHER"); if(!c.cipher)c.cipher="unknown";
  c.operation=operation; c.backend=getenv("HHE_BENCH_BACKEND");
  c.security_profile=getenv("HHE_BENCH_SECURITY_PROFILE");
  c.implementation=getenv("HHE_BENCH_IMPLEMENTATION"); if(!c.implementation)c.implementation="c-benchmark";
#ifdef __cplusplus
  c.language="cpp";
#else
  c.language=getenv("HHE_BENCH_LANGUAGE"); if(!c.language)c.language="c";
#endif
  c.output_path=getenv("HHE_BENCH_OUTPUT"); if(!c.output_path)c.output_path="benchmark.jsonl";
  c.iterations=iterations; c.cpu=-1; c.policy=HHE_POLICY_WARN;
  const char* value=getenv("HHE_BENCH_ITERATIONS");
  if(value){size_t requested=(size_t)strtoull(value,NULL,10);if(requested>0)c.iterations=requested;}
  c.warmups=default_warmups(c.iterations);
  value=getenv("HHE_BENCH_WARMUPS"); if(value)c.warmups=(size_t)strtoull(value,NULL,10);
  value=getenv("HHE_BENCH_CPU"); if(value)c.cpu=(int)strtol(value,NULL,10);
  value=getenv("HHE_BENCH_POLICY"); if(value&&!strcmp(value,"strict"))c.policy=HHE_POLICY_STRICT;
  return c;
}

int HHEBenchmarkBegin(HHEBenchmarkSession* s, const HHEBenchmarkConfig* c) {
  memset(s,0,sizeof(*s)); s->config=*c; s->actual_cpu=-1; s->affinity_verified=-1;
  s->invariant_tsc=cpu_flag("constant_tsc"); s->nonstop_tsc=cpu_flag("nonstop_tsc");
  s->rdtscp_supported=detect_rdtscp(); s->turbo_enabled=-1;
  if(!read_text("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor",s->governor,sizeof(s->governor))) s->governor[0]=0;
  char turbo[8]; if(read_text("/sys/devices/system/cpu/intel_pstate/no_turbo",turbo,sizeof(turbo))) s->turbo_enabled=!strcmp(turbo,"0");
  if(c->cpu>=0) {
    cpu_set_t set; CPU_ZERO(&set); CPU_SET(c->cpu,&set);
    if(sched_setaffinity(0,sizeof(set),&set)!=0){perror("sched_setaffinity");s->failed=1;return -1;}
  }
  s->actual_cpu=sched_getcpu();
  if(c->cpu>=0)s->affinity_verified=s->actual_cpu==c->cpu;
  if(!s->rdtscp_supported)append_warning(s,"RDTSCP unavailable");
  if(s->invariant_tsc!=1)append_warning(s,"invariant TSC not verified");

  if(s->nonstop_tsc!=1)append_warning(s,"nonstop TSC not verified");
  if(strcmp(s->governor,"performance"))append_warning(s,"governor not performance");
  if(s->turbo_enabled!=0)append_warning(s,"Turbo enabled or unavailable");
  if(s->affinity_verified==0)append_warning(s,"CPU affinity verification failed");
  s->checks_passed=s->warnings[0]==0;
  if(c->policy==HHE_POLICY_STRICT&&!s->checks_passed){fprintf(stderr,"%s\n",s->warnings);s->failed=1;return -1;}
  s->output=fopen(c->output_path,"a"); if(!s->output){perror("benchmark output");s->failed=1;return -1;}
  s->measured_ns=(uint64_t*)calloc(c->iterations,sizeof(uint64_t)); if(!s->measured_ns){fclose(s->output);s->failed=1;return -1;}
  snprintf(s->run_id,sizeof(s->run_id),"%ld-%lld",(long)getpid(),(long long)time(NULL));
  return 0;
}

void HHEBenchmarkIterationBegin(HHEBenchmarkSession* s,HHEPhase phase,size_t index) {
  s->phase=phase;s->index=index;s->before=memory_sample();
  clock_gettime(CLOCK_MONOTONIC,&s->start_time);
  s->cycles_valid=s->rdtscp_supported;
  if(s->cycles_valid)s->start_cycles=tsc_start();
}

int HHEBenchmarkIterationEnd(HHEBenchmarkSession* s) {
  uint64_t end_cycles=s->cycles_valid?tsc_end():0; struct timespec end;
  clock_gettime(CLOCK_MONOTONIC,&end); HHEMemorySample after=memory_sample();
  int64_t ns=elapsed_ns(s->start_time,end);
  if(s->phase==HHE_PHASE_MEASURED&&s->measured_count<s->config.iterations)s->measured_ns[s->measured_count++]=(uint64_t)ns;
  char stamp[32],model[256];timestamp(stamp);cpu_model(model);long cpus=sysconf(_SC_NPROCESSORS_ONLN);
  const char* hostname=getenv("HHE_BENCH_HOSTNAME");if(!hostname||!hostname[0])hostname=getenv("HOSTNAME");
  FILE* o=s->output;
  fputs("{\"schema_version\":\"1.0.0\",\"run_id\":",o);json_string(o,s->run_id);
  fputs(",\"timestamp_utc\":",o);json_string(o,stamp);
  fputs(",\"benchmark\":{\"cipher\":",o);json_string(o,s->config.cipher);
  fputs(",\"operation\":",o);json_string(o,s->config.operation);fputs(",\"backend\":",o);json_string(o,s->config.backend);
  fputs(",\"security_profile\":",o);json_string(o,s->config.security_profile);fputs(",\"parameters\":{}}",o);
  fputs(",\"runtime\":{\"language\":",o);json_string(o,s->config.language);fputs(",\"implementation\":",o);json_string(o,s->config.implementation);
  fputs(",\"version\":null,\"gc_mode\":\"not_applicable\",\"allocator\":\"system_glibc\"}",o);
  fputs(",\"host\":{\"hostname\":",o);json_string(o,hostname&&hostname[0]?hostname:NULL);fputs(",\"os\":\"linux\",\"arch\":",o);
#if defined(__x86_64__)
  json_string(o,"x86_64");
#else
  json_string(o,"unknown");
#endif
  fputs(",\"cpu_model\":",o);json_string(o,model[0]?model:NULL);fprintf(o,",\"logical_cpus\":%ld}",cpus);
  fprintf(o,",\"environment\":{\"policy\":\"%s\",\"requested_cpu\":",s->config.policy==HHE_POLICY_STRICT?"strict":"warn");
  if(s->config.cpu>=0)fprintf(o,"%d",s->config.cpu);else fputs("null",o);
  fprintf(o,",\"actual_cpu\":%d,\"affinity_verified\":",s->actual_cpu);nullable_bool(o,s->affinity_verified);
  fputs(",\"rdtscp_supported\":",o);nullable_bool(o,s->rdtscp_supported);fputs(",\"invariant_tsc\":",o);nullable_bool(o,s->invariant_tsc);
  fputs(",\"nonstop_tsc\":",o);nullable_bool(o,s->nonstop_tsc);fputs(",\"governor\":",o);json_string(o,s->governor[0]?s->governor:NULL);
  fputs(",\"turbo_enabled\":",o);nullable_bool(o,s->turbo_enabled);fprintf(o,",\"checks_passed\":%s,\"warnings\":[",s->checks_passed?"true":"false");
  if(s->warnings[0]) json_string(o,s->warnings);
  fputs("]}",o);
  fprintf(o,",\"iteration\":{\"phase\":\"%s\",\"index\":%zu}",s->phase==HHE_PHASE_WARMUP?"warmup":"measured",s->index);
  fprintf(o,",\"measurement\":{\"time_ns\":%lld,\"cycles\":",(long long)ns);
  if(s->cycles_valid)fprintf(o,"%llu",(unsigned long long)(end_cycles-s->start_cycles));else fputs("null",o);
  fputs(",\"memory\":{\"rss_bytes_delta\":",o);nullable_i64(o,s->before.rss_valid&&after.rss_valid,after.rss_bytes-s->before.rss_bytes);
  fputs(",\"heap_bytes_delta\":",o);nullable_i64(o,s->before.heap_valid&&after.heap_valid,after.heap_bytes-s->before.heap_bytes);
  fputs(",\"peak_rss_bytes_delta\":",o);nullable_i64(o,s->before.peak_valid&&after.peak_valid,after.peak_rss_bytes-s->before.peak_rss_bytes);
  fputs(",\"ru_maxrss_bytes\":",o);nullable_i64(o,after.peak_valid,after.peak_rss_bytes);
  fputs("}},\"status\":\"ok\",\"error\":null}\n",o);
  /* Multiple KAT sessions may share one append path. Flush each complete
     record before another buffered stream can write to that path. */
  if(fflush(o)!=0||ferror(o)){s->failed=1;return -1;} return 0;
}

static int compare_u64(const void* a,const void* b){uint64_t x=*(const uint64_t*)a,y=*(const uint64_t*)b;return(x>y)-(x<y);}
int HHEBenchmarkFinish(HHEBenchmarkSession* s) {
  int result=s->failed?-1:0;
  if(s->output&&fflush(s->output)!=0)result=-1;
  if(s->output&&fclose(s->output)!=0)result=-1;
  if(s->measured_count){qsort(s->measured_ns,s->measured_count,sizeof(uint64_t),compare_u64);
    printf("-> %s: n=%zu, median=%llu ns, min=%llu ns, max=%llu ns\n",s->config.operation,s->measured_count,
      (unsigned long long)s->measured_ns[s->measured_count/2],(unsigned long long)s->measured_ns[0],
      (unsigned long long)s->measured_ns[s->measured_count-1]);}
  free(s->measured_ns);s->measured_ns=NULL;return result;
}
int HHEBenchmarkRun(const HHEBenchmarkConfig* c,HHEBenchmarkFn fn,void* context) {
  HHEBenchmarkSession s;if(HHEBenchmarkBegin(&s,c)!=0)return -1;
  for(int phase=0;phase<2&&!s.failed;++phase){size_t count=phase?c->iterations:c->warmups;
    for(size_t i=0;i<count;++i){HHEBenchmarkIterationBegin(&s,phase?HHE_PHASE_MEASURED:HHE_PHASE_WARMUP,i);fn(context);if(HHEBenchmarkIterationEnd(&s)!=0)break;}}
  return HHEBenchmarkFinish(&s);
}
void PrintHeader(const char* h){printf("\n ------------------===|\t %s \t|===------------------\n",h);}
void GetMemoryUsage(long* r,long* h,long* p){HHEMemorySample s=memory_sample();*r=(long)(s.rss_bytes/1024);*h=(long)(s.heap_bytes/1024);*p=(long)(s.peak_rss_bytes/1024);}
void PrintTimeNS(struct timespec a,struct timespec b){long long ns=(long long)elapsed_ns(a,b);printf("-> Time: %.4f (ms), %lld (ns)\n",ns/1e6,ns);}

void PrintCycles(uint64_t a,uint64_t b){printf("-> Cycles: %llu\n",(unsigned long long)(b-a));}
void PrintMemory(const char* n,long a,long b){printf("-> %s Memory: %ld KB\n",n,b-a);}
