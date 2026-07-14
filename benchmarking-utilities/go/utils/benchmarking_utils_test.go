package benchmark

import "testing"

func TestProportionalWarmups(t *testing.T) {
	if proportionalWarmups(1) != 1 || proportionalWarmups(100) != 10 || proportionalWarmups(1000) != 10 {
		t.Fatal("unexpected warm-up rule")
	}
}
func TestSignedDifference(t *testing.T) {
	before, after := int64(10), int64(3)
	got := signedDifference(&before, &after)
	if got == nil || *got != -7 {
		t.Fatalf("got %v", got)
	}
}
func TestMemoryDifferenceRetainsAbsoluteRuMaxRSS(t *testing.T) {
	beforePeak, afterPeak := int64(100), int64(150)
	memory := memoryDifference(memorySample{peak: &beforePeak}, memorySample{peak: &afterPeak})
	if memory.Peak == nil || *memory.Peak != 50 || memory.RuMaxRSS == nil || *memory.RuMaxRSS != 150 {
		t.Fatalf("unexpected peak memory: %+v", memory)
	}
}
func TestBenchmarkHostnameOverride(t *testing.T) {
	t.Setenv("HHE_BENCH_HOSTNAME", "physical-host-01")
	hostname := benchmarkHostname()
	if hostname == nil || *hostname != "physical-host-01" {
		t.Fatalf("hostname=%v", hostname)
	}
}
func TestGCSettingRestored(t *testing.T) {
	restore := configureGC(GCDisabled)
	restore()
}
