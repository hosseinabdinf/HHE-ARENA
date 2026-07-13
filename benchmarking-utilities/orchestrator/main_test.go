package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type call struct {
	name string
	args []string
	env  []string
}
type fakeRunner struct{ calls []call }

func (f *fakeRunner) Run(n string, a []string, d string, e []string, out, errw io.Writer) error {
	f.calls = append(f.calls, call{n, append([]string{}, a...), append([]string{}, e...)})
	return nil
}
func manifestFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "manifest.json")
	data := "{\"version\":1,\"preflight\":\"/bin/preflight\",\"benchmarks\":[{\"id\":\"b-rust\",\"cipher\":\"hera\",\"language\":\"rust\",\"operation\":\"decrypt\",\"backend\":\"tfhe\",\"executable\":\"/bin/rust\",\"args\":[\"--count={iterations}\"]},{\"id\":\"a-go\",\"cipher\":\"hera\",\"language\":\"go\",\"operation\":\"encrypt\",\"backend\":\"tfhe\",\"security_profile\":\"HHE-128\",\"executable\":\"/bin/go\"}]}"
	if err := os.WriteFile(p, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}
func TestBenchDispatchEnvironment(t *testing.T) {
	p := manifestFile(t)
	out := filepath.Join(t.TempDir(), "raw.jsonl")
	f := &fakeRunner{}
	var stdout, stderr bytes.Buffer
	got := run([]string{"bench", "--manifest", p, "--language", "rust", "--iterations", "7", "--warmup", "2", "--core", "3", "--gc", "not_applicable", "--policy", "strict", "--output", out}, f, &stdout, &stderr)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, stderr.String())
	}
	if len(f.calls) != 1 || f.calls[0].name != "/bin/rust" {
		t.Fatalf("calls=%+v", f.calls)
	}
	if f.calls[0].args[0] != "--count=7" {
		t.Fatal(f.calls[0].args)
	}
	for k, v := range map[string]string{"HHE_BENCH_ITERATIONS": "7", "HHE_BENCH_WARMUPS": "2", "HHE_BENCH_CPU": "3", "HHE_BENCH_OUTPUT": out} {
		if x, ok := envValue(f.calls[0].env, k); !ok || x != v {
			t.Fatalf("%s=%q,%v", k, x, ok)
		}
	}
}
func TestBenchAllStableRawFiles(t *testing.T) {
	p := manifestFile(t)
	dir := filepath.Join(t.TempDir(), "raw")
	f := &fakeRunner{}
	var stdout, stderr bytes.Buffer
	got := run([]string{"bench-all", "--manifest", p, "--output", dir}, f, &stdout, &stderr)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, stderr.String())
	}
	if len(f.calls) != 2 {
		t.Fatal(len(f.calls))
	}
	a, _ := envValue(f.calls[0].env, "HHE_BENCH_OUTPUT")
	b, _ := envValue(f.calls[1].env, "HHE_BENCH_OUTPUT")
	if !strings.HasSuffix(a, "001-a-go.jsonl") || !strings.HasSuffix(b, "002-b-rust.jsonl") {
		t.Fatalf("%q %q", a, b)
	}
}
func TestBenchNeedsOneMatch(t *testing.T) {
	p := manifestFile(t)
	var out, errw bytes.Buffer
	if got := run([]string{"bench", "--manifest", p}, &fakeRunner{}, &out, &errw); got != 65 {
		t.Fatalf("exit=%d", got)
	}
}
func TestPreflightForwarding(t *testing.T) {
	p := manifestFile(t)
	f := &fakeRunner{}
	var out, errw bytes.Buffer
	if got := run([]string{"preflight", "--manifest", p, "--policy", "warn", "--core", "4"}, f, &out, &errw); got != 0 {
		t.Fatal(got, errw.String())
	}
	if len(f.calls) != 1 || f.calls[0].name != "/bin/preflight" || strings.Join(f.calls[0].args, " ") != "--warn --cpu 4" {
		t.Fatalf("%+v", f.calls)
	}
}
func TestOutputCollision(t *testing.T) {
	p := manifestFile(t)
	out := filepath.Join(t.TempDir(), "exists.jsonl")
	if err := os.WriteFile(out, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if got := run([]string{"bench", "--manifest", p, "--language", "go", "--output", out}, &fakeRunner{}, &stdout, &stderr); got != 1 {
		t.Fatalf("exit=%d stderr=%s", got, stderr.String())
	}
}
func envValue(e []string, k string) (string, bool) {
	p := k + "="
	for i := len(e) - 1; i >= 0; i-- {
		if strings.HasPrefix(e[i], p) {
			return strings.TrimPrefix(e[i], p), true
		}
	}
	return "", false
}
