package main

import (
	"bytes"
	"errors"
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
type fakeRunner struct {
	calls  []call
	failAt int
}

func (f *fakeRunner) Run(n string, a []string, d string, e []string, out, errw io.Writer) error {
	f.calls = append(f.calls, call{n, append([]string{}, a...), append([]string{}, e...)})
	if f.failAt > 0 && len(f.calls) == f.failAt {
		return errors.New("synthetic command failure")
	}
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
func catalogFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "images.json")
	data := `{"version":1,"matrix_version":1,"tooling":{"tag":"tooling:test","dockerfile":"Dockerfile","context":"."},"images":[{"id":"aes-c","framework":"AES/OpenSSL","cipher":"AES-128-CBC","language":"c","backend":"openssl","operations":["encrypt"],"local_tag":"aes:test","pull_ref":null,"manifest":"manifests/aes.json","build":{"tag":"aes:test","dockerfile":"frameworks/aes/Dockerfile","context":".","args":{"TOOLING_IMAGE":"tooling:test"}}},{"id":"elisabeth-rust","framework":"Elisabeth","cipher":"Elisabeth","language":"rust","backend":"native","operations":["encrypt","decrypt"],"local_tag":"elisabeth:test","pull_ref":"registry.example/elisabeth@sha256:abc","manifest":"manifests/elisabeth.json","build":{"tag":"elisabeth:test","dockerfile":"frameworks/elisabeth/Dockerfile","context":"."}}]}`
	if err := os.WriteFile(p, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}
func matrixV2Files(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	catalog := filepath.Join(dir, "images.v2.json")
	matrix := filepath.Join(dir, "support-matrix.v2.json")
	catalogData := `{"version":1,"matrix_version":2,"tooling":{"tag":"tooling:test","dockerfile":"Dockerfile","context":"."},"images":[{"id":"aes-c","framework":"AES/OpenSSL","cipher":"AES suite","language":"c","backend":"openssl","operations":["encrypt"],"matrix_targets":[{"variant_id":"aes128-cbc-openssl","cipher":"AES-128-CBC","backend":"openssl","operations":["encrypt"]}],"local_tag":"aes:test","pull_ref":null,"manifest":"manifests/aes.json","build":{"tag":"aes:test","dockerfile":"frameworks/aes/Dockerfile","context":"."}},{"id":"rubato-go","framework":"HHELand","cipher":"Rubato","language":"go","backend":"Lattigo CKKS","operations":["HHE.Decomp()"],"matrix_targets":[{"variant_id":"rubato-128l","cipher":"Rubato","backend":"Lattigo CKKS","operations":["HHE.Decomp()"]}],"local_tag":"rubato:test","manifest":"manifests/rubato.json","build":{"tag":"rubato:test","dockerfile":"frameworks/rubato/Dockerfile","context":"."}}]}`
	matrixData := `{"matrix_version":2,"selector_contract":{"wildcard":"all","selection_order":["family","path","cipher","variant","framework","backend","language"]},"benchmark_variants":[{"id":"aes128-cbc-openssl","family":"aes","cipher":"AES","variant":"AES-128-CBC","framework":"AES/OpenSSL","language":"c","implementation_id":"aes-heavy","targets":[{"path":"ske","backends":["openssl"],"operations":["encrypt"],"status":"ready","implementation_id":"aes-c"}]},{"id":"rubato-128l","family":"he_friendly","cipher":"Rubato","variant":"128L","framework":"HHELand","language":"go","implementation_id":"rubato-go","targets":[{"path":"hhe","backends":["Lattigo CKKS"],"operations":["HHE.Decomp()"],"status":"deferred"}]}]}`
	if err := os.WriteFile(catalog, []byte(catalogData), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(matrix, []byte(matrixData), 0644); err != nil {
		t.Fatal(err)
	}
	return catalog, matrix
}
func TestCatalogAllowsOneVariantAcrossDistinctBackends(t *testing.T) {
	p := filepath.Join(t.TempDir(), "images.v2.json")
	data := `{"version":1,"matrix_version":2,"tooling":{"tag":"tooling:test","dockerfile":"Dockerfile","context":"."},"images":[{"id":"full-lut","framework":"AES/Full-LUT","cipher":"AES-128 full-LUT","language":"cpp","backend":"TFHE","operations":["key-expand","transcipher"],"matrix_targets":[{"variant_id":"aes128-full-lut","cipher":"AES-128 full-LUT","backend":"native","operations":["key-expand"]},{"variant_id":"aes128-full-lut","cipher":"AES-128 full-LUT","backend":"TFHE","operations":["transcipher"]}],"local_tag":"full-lut:test","manifest":"manifests/full-lut.json","build":{"tag":"full-lut:test","dockerfile":"frameworks/full-lut/Dockerfile","context":"."}}]}`
	if err := os.WriteFile(p, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadImageCatalog(p); err != nil {
		t.Fatal(err)
	}
}
func TestBenchDispatchEnvironment(t *testing.T) {
	p := manifestFile(t)
	out := filepath.Join(t.TempDir(), "raw.jsonl")
	f := &fakeRunner{}
	var stdout, stderr bytes.Buffer
	got := run([]string{"bench", "--manifest", p, "--language", "rust", "--iterations", "7", "--warmup", "2", "--core", "3", "--gc", "not_applicable", "--policy", "strict", "--host-id", "physical-host-01", "--output", out}, f, &stdout, &stderr)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, stderr.String())
	}
	if len(f.calls) != 1 || f.calls[0].name != "/bin/rust" {
		t.Fatalf("calls=%+v", f.calls)
	}
	if f.calls[0].args[0] != "--count=7" {
		t.Fatal(f.calls[0].args)
	}
	for k, v := range map[string]string{"HHE_BENCH_ITERATIONS": "7", "HHE_BENCH_WARMUPS": "2", "HHE_BENCH_CPU": "3", "HHE_BENCH_OUTPUT": out, "HHE_BENCH_HOSTNAME": "physical-host-01"} {
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
func TestBenchAddsJSONLExtension(t *testing.T) {
	p := manifestFile(t)
	out := filepath.Join(t.TempDir(), "raw")
	f := &fakeRunner{}
	var stdout, stderr bytes.Buffer
	if got := run([]string{"bench", "--manifest", p, "--language", "rust", "--output", out}, f, &stdout, &stderr); got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, stderr.String())
	}
	if got, _ := envValue(f.calls[0].env, "HHE_BENCH_OUTPUT"); got != out+".jsonl" {
		t.Fatalf("output=%q", got)
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
func TestContainerBenchBuildDryRun(t *testing.T) {
	p := catalogFile(t)
	f := &fakeRunner{}
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", p, "--project-root", t.TempDir(), "--cipher", "AES-128-CBC", "--operation", "encrypt", "--image-policy", "build", "--dry-run", "--core", "2", "--host-id", "physical-host-01", "--output", filepath.Join(t.TempDir(), "aes.jsonl")}, f, &out, &errw)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
	if len(f.calls) != 0 {
		t.Fatalf("dry run executed commands: %+v", f.calls)
	}
	s := out.String()
	for _, want := range []string{"docker-command=docker build", "frameworks", "docker-command=docker run --rm --env HHE_BENCH_HOSTNAME=physical-host-01 --cpuset-cpus 2", "aes:test bench", "--operation encrypt", "target=/work/results"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}
func TestContainerBenchAllCreatesStandardBundleAndReportDryRun(t *testing.T) {
	p := catalogFile(t)
	bundle := filepath.Join(t.TempDir(), "bundle")
	var out, errw bytes.Buffer
	got := run([]string{"bench-all", "--catalog", p, "--project-root", t.TempDir(), "--image-policy", "local", "--dry-run", "--output", bundle}, &fakeRunner{}, &out, &errw)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
	s := out.String()
	for _, want := range []string{
		"--output /work/results/001-aes-c-encrypt.jsonl",
		"--output /work/results/002-elisabeth-rust-decrypt.jsonl",
		"--output /work/results/003-elisabeth-rust-encrypt.jsonl",
		"aes:test report /work/results/raw/001-aes-c-encrypt.jsonl",
		"/work/results/raw/003-elisabeth-rust-encrypt.jsonl --output-dir /work/results/report",
		"result-bundle=" + bundle,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}
func TestContainerReportFailureFailsWorkflow(t *testing.T) {
	p := catalogFile(t)
	f := &fakeRunner{failAt: 3}
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", p, "--cipher", "AES-128-CBC", "--operation", "encrypt", "--image-policy", "local", "--output", filepath.Join(t.TempDir(), "bundle")}, f, &out, &errw)
	if got != 1 || !strings.Contains(errw.String(), "report generation failed") {
		t.Fatalf("exit=%d stderr=%s calls=%+v", got, errw.String(), f.calls)
	}
}
func TestContainerBenchRequiresOneOperation(t *testing.T) {
	p := catalogFile(t)
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", p, "--cipher", "Elisabeth", "--dry-run"}, &fakeRunner{}, &out, &errw)
	if got != 65 || !strings.Contains(errw.String(), "found 2") {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
}
func TestContainerPullUnavailableIsExplicit(t *testing.T) {
	p := catalogFile(t)
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", p, "--cipher", "AES-128-CBC", "--operation", "encrypt", "--image-policy", "pull", "--dry-run"}, &fakeRunner{}, &out, &errw)
	if got != 1 || !strings.Contains(errw.String(), "no immutable pull_ref is published") {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
}
func TestCatalogListShowsUnpublishedPull(t *testing.T) {
	p := catalogFile(t)
	var out, errw bytes.Buffer
	got := run([]string{"list", "--catalog", p, "--cipher", "AES-128-CBC"}, &fakeRunner{}, &out, &errw)
	if got != 0 || !strings.Contains(out.String(), "aes:test\tunpublished") {
		t.Fatalf("exit=%d stdout=%s stderr=%s", got, out.String(), errw.String())
	}
}
func TestMatrixV2ListUsesHierarchyAndShowsReadiness(t *testing.T) {
	catalog, _ := matrixV2Files(t)
	var out, errw bytes.Buffer
	got := run([]string{"list", "--catalog", catalog, "--family", "aes", "--path", "ske", "--cipher", "AES", "--variant", "AES-128-CBC", "--framework", "AES/OpenSSL", "--backend", "openssl", "--language", "c"}, &fakeRunner{}, &out, &errw)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
	for _, want := range []string{"aes128-cbc-openssl", "\taes\tske\tAES\tAES-128-CBC\tAES/OpenSSL\topenssl\tc\tencrypt\tready\trunnable"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %s", want, out.String())
		}
	}
}
func TestMatrixV2AllWildcardListsDeclaredUnavailableTargets(t *testing.T) {
	_, matrix := matrixV2Files(t)
	var out, errw bytes.Buffer
	got := run([]string{"list", "--matrix", matrix, "--family", "all", "--path", "all"}, &fakeRunner{}, &out, &errw)
	if got != 0 || !strings.Contains(out.String(), "rubato-128l\the_friendly\thhe") || !strings.Contains(out.String(), "\tdeferred\tnot-runnable") {
		t.Fatalf("exit=%d stdout=%s stderr=%s", got, out.String(), errw.String())
	}
}
func TestMatrixV2BenchRunsOnlyReadyCatalogTarget(t *testing.T) {
	catalog, _ := matrixV2Files(t)
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", catalog, "--family", "aes", "--path", "ske", "--cipher", "AES", "--variant", "AES-128-CBC", "--image-policy", "local", "--dry-run", "--output", filepath.Join(t.TempDir(), "bundle")}, &fakeRunner{}, &out, &errw)
	if got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
	if !strings.Contains(out.String(), "--output /work/results/001-aes128-cbc-openssl-encrypt.jsonl") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	if !strings.Contains(out.String(), "--cipher AES-128-CBC") {
		t.Fatalf("variant dispatch cipher missing: %s", out.String())
	}
}
func TestMatrixV2BenchRejectsDeferredTarget(t *testing.T) {
	catalog, _ := matrixV2Files(t)
	var out, errw bytes.Buffer
	got := run([]string{"bench", "--catalog", catalog, "--family", "he_friendly", "--path", "hhe", "--cipher", "Rubato", "--variant", "128L", "--dry-run"}, &fakeRunner{}, &out, &errw)
	if got != 65 || !strings.Contains(errw.String(), "none are runnable") {
		t.Fatalf("exit=%d stdout=%s stderr=%s", got, out.String(), errw.String())
	}
}

func TestMatrixV2BenchCanDryRunOneBoundedDeferredTarget(t *testing.T) {
	catalog, _ := matrixV2Files(t)
	args := []string{"bench", "--catalog", catalog, "--family", "he_friendly", "--path", "hhe", "--cipher", "Rubato", "--variant", "128L", "--framework", "HHELand", "--backend", "Lattigo CKKS", "--language", "go", "--operation", "HHE.Decomp()", "--image-policy", "local", "--iterations", "1", "--warmup", "0", "--core", "2", "--policy", "warn", "--validate-deferred", "--dry-run", "--output", filepath.Join(t.TempDir(), "deferred")}
	var out, errw bytes.Buffer
	got := run(args, &fakeRunner{}, &out, &errw)
	if got != 0 || !strings.Contains(out.String(), "rubato:test bench --cipher Rubato") || !strings.Contains(out.String(), "--iterations 1") {
		t.Fatalf("exit=%d stdout=%s stderr=%s", got, out.String(), errw.String())
	}
}

func TestValidateDeferredRequiresSingleIterationAndWarmupZero(t *testing.T) {
	catalog, _ := matrixV2Files(t)
	for _, changed := range [][]string{{"--iterations", "2"}, {"--warmup", "1"}} {
		args := []string{"bench", "--catalog", catalog, "--family", "he_friendly", "--path", "hhe", "--cipher", "Rubato", "--variant", "128L", "--framework", "HHELand", "--backend", "Lattigo CKKS", "--language", "go", "--operation", "HHE.Decomp()", "--image-policy", "local", "--iterations", "1", "--warmup", "0", "--core", "2", "--policy", "warn", "--validate-deferred", "--dry-run"}
		for i := 0; i < len(args)-1; i++ {
			if args[i] == changed[0] {
				args[i+1] = changed[1]
			}
		}
		var out, errw bytes.Buffer
		got := run(args, &fakeRunner{}, &out, &errw)
		if got != 64 || !strings.Contains(errw.String(), "requires --iterations 1 --warmup 0") {
			t.Fatalf("args=%v exit=%d stdout=%s stderr=%s", changed, got, out.String(), errw.String())
		}
	}
}
func TestHelpUsesConfiguredArenaTitle(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ARENA.txt")
	if err := os.WriteFile(p, []byte("HHE-ARENA-TITLE\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HHE_ARENA_TITLE_FILE", p)
	var out, errw bytes.Buffer
	if got := run([]string{"help"}, &fakeRunner{}, &out, &errw); got != 0 {
		t.Fatalf("exit=%d stderr=%s", got, errw.String())
	}
	if !strings.HasPrefix(out.String(), "HHE-ARENA-TITLE\nUsage: hhe") {
		t.Fatalf("unexpected help output: %q", out.String())
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
