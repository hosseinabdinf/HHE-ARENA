package ckks_cipher

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Policy string

const (
	PolicyStrict Policy = "strict"
	PolicyWarn   Policy = "warn"
)

type GCMode string

const (
	GCEnabled  GCMode = "enabled"
	GCDisabled GCMode = "disabled"
	GCUnknown  GCMode = "unknown"
)

type phase string

const (
	phaseWarmup   phase = "warmup"
	phaseMeasured phase = "measured"
)

type Config struct {
	Cipher          string
	Operation       string
	Backend         *string
	SecurityProfile *string
	Implementation  string
	Iterations      int
	Warmups         *int
	CPU             *int
	Policy          Policy
	GCMode          GCMode
	Output          string
}

type environment struct {
	Policy           Policy   `json:"policy"`
	RequestedCPU     *int     `json:"requested_cpu"`
	ActualCPU        *int     `json:"actual_cpu"`
	AffinityVerified *bool    `json:"affinity_verified"`
	RDTSCPAvailable  *bool    `json:"rdtscp_supported"`
	InvariantTSC     *bool    `json:"invariant_tsc"`
	NonstopTSC       *bool    `json:"nonstop_tsc"`
	Governor         *string  `json:"governor"`
	TurboEnabled     *bool    `json:"turbo_enabled"`
	ChecksPassed     bool     `json:"checks_passed"`
	Warnings         []string `json:"warnings"`
}
type memoryDelta struct {
	RSS      *int64 `json:"rss_bytes_delta"`
	Heap     *int64 `json:"heap_bytes_delta"`
	Peak     *int64 `json:"peak_rss_bytes_delta"`
	RuMaxRSS *int64 `json:"ru_maxrss_bytes"`
}
type memorySample struct{ rss, heap, peak *int64 }
type record struct {
	SchemaVersion string `json:"schema_version"`
	RunID         string `json:"run_id"`
	TimestampUTC  string `json:"timestamp_utc"`
	Benchmark     struct {
		Cipher          string         `json:"cipher"`
		Operation       string         `json:"operation"`
		Backend         *string        `json:"backend"`
		SecurityProfile *string        `json:"security_profile"`
		Parameters      map[string]any `json:"parameters"`
	} `json:"benchmark"`
	Runtime struct {
		Language       string  `json:"language"`
		Implementation string  `json:"implementation"`
		Version        *string `json:"version"`
		GCMode         GCMode  `json:"gc_mode"`
		Allocator      string  `json:"allocator"`
	} `json:"runtime"`
	Host struct {
		Hostname    *string `json:"hostname"`
		OS          string  `json:"os"`
		Arch        string  `json:"arch"`
		CPUModel    *string `json:"cpu_model"`
		LogicalCPUs *int    `json:"logical_cpus"`
	} `json:"host"`
	Environment environment `json:"environment"`
	Iteration   struct {
		Phase phase `json:"phase"`
		Index int   `json:"index"`
	} `json:"iteration"`
	Measurement struct {
		TimeNS *int64      `json:"time_ns"`
		Cycles *uint64     `json:"cycles"`
		Memory memoryDelta `json:"memory"`
	} `json:"measurement"`
	Status string  `json:"status"`
	Error  *string `json:"error"`
}

var calibrationOnce sync.Once
var cyclesPerSecond *float64

func Run(config Config, target func()) error {
	if config.Iterations <= 0 {
		return errors.New("iterations must be positive")
	}
	if config.Cipher == "" {
		return errors.New("cipher must not be empty")
	}
	if config.Operation == "" {
		return errors.New("operation must not be empty")
	}
	if config.Implementation == "" {
		config.Implementation = "go-benchmark"
	}
	if config.Output == "" {
		config.Output = "benchmark.jsonl"
	}
	if config.Policy == "" {
		config.Policy = PolicyWarn
	}
	if config.GCMode == "" {
		config.GCMode = GCUnknown
	}

	if config.Policy != PolicyStrict && config.Policy != PolicyWarn {
		return fmt.Errorf("invalid policy %q", config.Policy)
	}
	if config.GCMode != GCEnabled && config.GCMode != GCDisabled && config.GCMode != GCUnknown {
		return fmt.Errorf("invalid GC mode %q", config.GCMode)
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	env, err := prepareEnvironment(config.CPU, config.Policy)
	if err != nil {
		return err
	}
	calibrationOnce.Do(func() { cyclesPerSecond = calibrateTSC() })
	_ = cyclesPerSecond
	counterOverhead := MeasureCounterOverhead(1000)

	restoreGC := configureGC(config.GCMode)
	defer restoreGC()

	warmups := proportionalWarmups(config.Iterations)
	if config.Warmups != nil {
		if *config.Warmups < 0 {
			return errors.New("warmups must not be negative")
		}
		warmups = *config.Warmups
	}
	file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	failed := true
	defer func() {
		if failed {
			_ = file.Close()
		}
	}()
	runID := fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	measured := make([]int64, 0, config.Iterations)

	for _, currentPhase := range []phase{phaseWarmup, phaseMeasured} {
		count := warmups
		if currentPhase == phaseMeasured {
			count = config.Iterations
		}
		for index := 0; index < count; index++ {
			before := readMemory()
			startTime := time.Now()
			var startCycles uint64
			cyclesValid := RDTSCPAvailable()
			if cyclesValid {
				startCycles = tscStart()
			}
			target()
			var endCycles uint64
			if cyclesValid {
				endCycles = tscEnd()
			}
			elapsed := time.Since(startTime).Nanoseconds()
			after := readMemory()
			if currentPhase == phaseMeasured {
				measured = append(measured, elapsed)
			}
			item := makeRecord(config, env, runID, currentPhase, index, elapsed,
				cycleDifference(cyclesValid, startCycles, endCycles), memoryDifference(before, after))
			if err := json.NewEncoder(writer).Encode(item); err != nil {
				return fmt.Errorf("write JSONL: %w", err)
			}
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush JSONL: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close JSONL: %w", err)
	}
	failed = false
	printSummary(config.Operation, measured, counterOverhead)
	return nil
}

func Benchmark(name string, target func()) { BenchmarkIter(name, 1, target) }
func BenchmarkIter(name string, iterations int, target func()) {
	config := ConfigFromEnv(name, iterations)
	if err := Run(config, target); err != nil {
		panic(fmt.Sprintf("benchmark failed: %v", err))
	}
}
func ConfigFromEnv(operation string, iterations int) Config {
	config := Config{
		Cipher:    envDefault("HHE_BENCH_CIPHER", "unknown"),
		Operation: operation, Implementation: envDefault("HHE_BENCH_IMPLEMENTATION", "go-benchmark"),
		Iterations: iterations, Policy: Policy(envDefault("HHE_BENCH_POLICY", "warn")),
		GCMode: GCMode(envDefault("HHE_BENCH_GC", "unknown")),

		Output: envDefault("HHE_BENCH_OUTPUT", "benchmark.jsonl"),
	}
	if value, err := strconv.Atoi(os.Getenv("HHE_BENCH_ITERATIONS")); err == nil && value > 0 {
		config.Iterations = value
	}
	config.Backend = envPointer("HHE_BENCH_BACKEND")
	config.SecurityProfile = envPointer("HHE_BENCH_SECURITY_PROFILE")
	if value, err := strconv.Atoi(os.Getenv("HHE_BENCH_CPU")); err == nil {
		config.CPU = &value
	}
	if value, err := strconv.Atoi(os.Getenv("HHE_BENCH_WARMUPS")); err == nil {
		config.Warmups = &value
	}
	return config
}

func prepareEnvironment(cpu *int, policy Policy) (environment, error) {
	env := environment{Policy: policy, RequestedCPU: cpu, Warnings: []string{}}
	if cpu != nil {
		if err := pinCurrentThread(*cpu); err != nil {
			return env, fmt.Errorf("pin OS thread: %w", err)
		}
	}
	env.ActualCPU = currentCPU()
	if cpu != nil && env.ActualCPU != nil {
		value := *cpu == *env.ActualCPU
		env.AffinityVerified = &value
	}
	available := RDTSCPAvailable()
	env.RDTSCPAvailable = &available
	env.InvariantTSC = cpuFlag("constant_tsc")
	env.NonstopTSC = cpuFlag("nonstop_tsc")
	env.Governor = readTrimmed("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	env.TurboEnabled = turboEnabled()
	if !available {
		env.Warnings = append(env.Warnings, "RDTSCP unavailable; cycles are null")
	}
	if env.InvariantTSC == nil || !*env.InvariantTSC {
		env.Warnings = append(env.Warnings, "invariant TSC not verified")
	}
	if env.NonstopTSC == nil || !*env.NonstopTSC {
		env.Warnings = append(env.Warnings, "nonstop TSC not verified")
	}
	if env.Governor == nil || *env.Governor != "performance" {
		env.Warnings = append(env.Warnings, "governor not performance")
	}
	if env.TurboEnabled == nil || *env.TurboEnabled {
		env.Warnings = append(env.Warnings, "Turbo enabled or unavailable")
	}
	if env.AffinityVerified != nil && !*env.AffinityVerified {
		env.Warnings = append(env.Warnings, "CPU affinity verification failed")
	}
	env.ChecksPassed = len(env.Warnings) == 0
	if policy == PolicyStrict && !env.ChecksPassed {
		return env, errors.New(strings.Join(env.Warnings, "; "))
	}
	return env, nil
}

func configureGC(mode GCMode) func() {
	switch mode {
	case GCDisabled:
		previous := debug.SetGCPercent(-1)
		return func() { debug.SetGCPercent(previous) }
	case GCEnabled:
		previous := debug.SetGCPercent(100)
		return func() { debug.SetGCPercent(previous) }
	default:
		return func() {}
	}
}
func proportionalWarmups(iterations int) int {
	value := (iterations + 9) / 10
	if value < 1 {
		return 1
	}
	if value > 10 {
		return 10
	}
	return value
}
func calibrateTSC() *float64 {
	if !RDTSCPAvailable() {
		return nil
	}
	startTime := time.Now()
	start := tscStart()
	for time.Since(startTime) < 100*time.Millisecond {
		runtime.Gosched()
	}
	elapsed := time.Since(startTime).Seconds()
	end := tscEnd()
	value := float64(end-start) / elapsed
	return &value
}
func MeasureCounterOverhead(samples int) uint64 {
	if samples <= 0 || !RDTSCPAvailable() {
		return 0
	}
	values := make([]uint64, samples)
	for index := range values {
		start := tscStart()
		values[index] = tscEnd() - start
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values[len(values)/2]
}
func cycleDifference(valid bool, start, end uint64) *uint64 {
	if !valid || end < start {
		return nil
	}
	value := end - start
	return &value
}
func signedDifference(before, after *int64) *int64 {
	if before == nil || after == nil {
		return nil
	}
	value := *after - *before
	return &value
}
func memoryDifference(before, after memorySample) memoryDelta {
	return memoryDelta{RSS: signedDifference(before.rss, after.rss), Heap: signedDifference(before.heap, after.heap), Peak: signedDifference(before.peak, after.peak), RuMaxRSS: after.peak}
}
func readMemory() memorySample {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	heap := int64(stats.Alloc)
	var sample memorySample
	sample.heap = &heap
	sample.rss = currentRSS()
	sample.peak = peakRSS()
	return sample
}

func makeRecord(config Config, env environment, runID string, currentPhase phase, index int,
	elapsed int64, cycles *uint64, memory memoryDelta) record {
	var item record
	item.SchemaVersion = "1.0.0"
	item.RunID = runID
	item.TimestampUTC = time.Now().UTC().Format(time.RFC3339Nano)
	item.Benchmark.Cipher = config.Cipher
	item.Benchmark.Operation = config.Operation
	item.Benchmark.Backend = config.Backend
	item.Benchmark.SecurityProfile = config.SecurityProfile
	item.Benchmark.Parameters = map[string]any{}
	item.Runtime.Language = "go"
	item.Runtime.Implementation = config.Implementation
	item.Runtime.GCMode = config.GCMode
	item.Runtime.Allocator = "go_runtime"
	item.Host.OS = runtime.GOOS
	item.Host.Arch = runtime.GOARCH
	if hostname, err := os.Hostname(); err == nil {
		item.Host.Hostname = &hostname
	}
	item.Host.CPUModel = cpuModel()
	cpus := runtime.NumCPU()
	item.Host.LogicalCPUs = &cpus
	item.Environment = env
	item.Iteration.Phase = currentPhase
	item.Iteration.Index = index
	item.Measurement.TimeNS = &elapsed
	item.Measurement.Cycles = cycles
	item.Measurement.Memory = memory
	item.Status = "ok"
	return item
}
func printSummary(name string, values []int64, overhead uint64) {
	if len(values) == 0 {
		return
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	fmt.Printf("-> %s: n=%d, median=%d ns, min=%d ns, max=%d ns, counter-boundary=%d cycles\n", name, len(values), values[len(values)/2], values[0], values[len(values)-1], overhead)
}
func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
func envPointer(name string) *string {
	if value := os.Getenv(name); value != "" {
		return &value
	}
	return nil
}
func readTrimmed(path string) *string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	value := strings.TrimSpace(string(data))
	return &value
}

func cpuFlag(flag string) *bool {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil
	}
	value := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "flags") {
			for _, item := range strings.Fields(line) {
				if item == flag {
					value = true
				}
			}
		}
	}
	return &value
}
func cpuModel() *string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				return &value
			}
		}
	}
	return nil
}
func turboEnabled() *bool {
	value := readTrimmed("/sys/devices/system/cpu/intel_pstate/no_turbo")
	if value == nil {
		return nil
	}
	enabled := *value == "0"
	return &enabled
}

func PrintHeader(header string) {
	fmt.Printf("\n ------------------===|\t %s \t|===------------------\n", header)
}
func PrintTimeNS(start, end time.Time) {
	diff := end.Sub(start)
	fmt.Printf("-> Time: %.4f (ms), %d (ns)\n", float64(diff.Nanoseconds())/1e6, diff.Nanoseconds())
}
func PrintCycles(start, end uint64)             { fmt.Printf("-> Cycles: %d\n", end-start) }
func PrintMemory(name string, start, end int64) { fmt.Printf("-> %s Memory: %d KB\n", name, end-start) }
