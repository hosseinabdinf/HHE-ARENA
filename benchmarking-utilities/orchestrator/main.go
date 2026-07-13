package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const defaultManifest = "/opt/hhe/benchmarks/manifest.json"

type Manifest struct {
	Version    int         `json:"version"`
	Preflight  string      `json:"preflight,omitempty"`
	Benchmarks []Benchmark `json:"benchmarks"`
}
type Benchmark struct {
	ID              string            `json:"id"`
	Cipher          string            `json:"cipher"`
	Language        string            `json:"language"`
	Operation       string            `json:"operation"`
	Backend         string            `json:"backend"`
	SecurityProfile string            `json:"security_profile,omitempty"`
	Executable      string            `json:"executable"`
	Args            []string          `json:"args,omitempty"`
	WorkDir         string            `json:"workdir,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	Disabled        bool              `json:"disabled,omitempty"`
}
type options struct {
	manifest, cipher, language, operation, backend, gc, policy, output string
	iterations, warmups, core                                          int
	preflight, appendMode, dryRun                                      bool
}
type runner interface {
	Run(string, []string, string, []string, io.Writer, io.Writer) error
}
type osRunner struct{}

func (osRunner) Run(n string, a []string, d string, e []string, out, errw io.Writer) error {
	c := exec.Command(n, a...)
	c.Dir = d
	c.Env = e
	c.Stdin = os.Stdin
	c.Stdout = out
	c.Stderr = errw
	return c.Run()
}
func main() { os.Exit(run(os.Args[1:], osRunner{}, os.Stdout, os.Stderr)) }
func run(a []string, r runner, out, errw io.Writer) int {
	if len(a) == 0 {
		usage(errw)
		return 64
	}
	switch a[0] {
	case "bench":
		return bench(a[1:], false, r, out, errw)
	case "bench-all":
		return bench(a[1:], true, r, out, errw)
	case "list":
		return list(a[1:], out, errw)
	case "preflight":
		return preflight(a[1:], r, out, errw)
	case "help", "-h", "--help":
		usage(out)
		return 0
	}
	fmt.Fprintf(errw, "unknown command %q\n", a[0])
	return 64
}
func usage(w io.Writer) {
	fmt.Fprint(w, `Usage: hhe <bench|bench-all|list|preflight> [options]
Selectors: --cipher --language --operation --backend (comma-separated)
Benchmark: --iterations N --warmup N --core N --gc MODE --policy MODE
           --output PATH --preflight --append --dry-run --manifest FILE
Manifest argument templates: {iterations} {warmups} {core} {gc} {policy}
 {output} {id} {cipher} {language} {operation} {backend} {security_profile}
`)
}
func flags(name string, errw io.Writer) (*flag.FlagSet, *options) {
	o := &options{manifest: defaultManifest, iterations: 10, warmups: -1, core: -1, gc: "unknown", policy: "warn", output: "results/raw"}
	if name == "bench" {
		o.output = "benchmark.jsonl"
	}
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(errw)
	f.StringVar(&o.manifest, "manifest", o.manifest, "manifest")
	f.StringVar(&o.cipher, "cipher", "", "cipher")
	f.StringVar(&o.language, "language", "", "language")
	f.StringVar(&o.operation, "operation", "", "operation")
	f.StringVar(&o.backend, "backend", "", "backend")
	f.IntVar(&o.iterations, "iterations", 10, "iterations")
	f.IntVar(&o.warmups, "warmup", -1, "warmups")
	f.IntVar(&o.core, "core", -1, "CPU")
	f.StringVar(&o.gc, "gc", "unknown", "GC mode")
	f.StringVar(&o.policy, "policy", "warn", "policy")
	f.StringVar(&o.output, "output", o.output, "output")
	f.BoolVar(&o.preflight, "preflight", false, "preflight")
	f.BoolVar(&o.appendMode, "append", false, "append")
	f.BoolVar(&o.dryRun, "dry-run", false, "dry run")
	return f, o
}
func valid(o *options) error {
	if o.iterations < 1 {
		return errors.New("--iterations must be positive")
	}
	if o.warmups < -1 {
		return errors.New("--warmup must be nonnegative")
	}
	if o.core < -1 {
		return errors.New("--core must be nonnegative")
	}
	if o.policy != "warn" && o.policy != "strict" {
		return errors.New("--policy must be strict or warn")
	}
	if !match("enabled,disabled,unknown,not_applicable", o.gc) {
		return errors.New("invalid --gc")
	}
	if strings.TrimSpace(o.output) == "" {
		return errors.New("empty --output")
	}
	return nil
}
func load(p string) (Manifest, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", e)
	}
	var m Manifest
	if e = json.Unmarshal(b, &m); e != nil {
		return m, fmt.Errorf("parse manifest: %w", e)
	}
	if m.Version != 1 {
		return m, fmt.Errorf("manifest version must be 1")
	}
	seen := map[string]bool{}
	for i, x := range m.Benchmarks {
		if x.ID == "" || x.Cipher == "" || x.Language == "" || x.Operation == "" || x.Backend == "" || x.Executable == "" {
			return m, fmt.Errorf("benchmark %d missing required field", i)
		}
		if seen[x.ID] {
			return m, fmt.Errorf("duplicate id %q", x.ID)
		}
		seen[x.ID] = true
	}
	return m, nil
}
func match(s, v string) bool {
	if strings.TrimSpace(s) == "" {
		return true
	}
	for _, x := range strings.Split(s, ",") {
		if strings.TrimSpace(x) == "*" || strings.EqualFold(strings.TrimSpace(x), v) {
			return true
		}
	}
	return false
}
func selectJobs(m Manifest, o *options) []Benchmark {
	var z []Benchmark
	for _, x := range m.Benchmarks {
		if !x.Disabled && match(o.cipher, x.Cipher) && match(o.language, x.Language) && match(o.operation, x.Operation) && match(o.backend, x.Backend) {
			z = append(z, x)
		}
	}
	sort.SliceStable(z, func(i, j int) bool { return z[i].ID < z[j].ID })
	return z
}
func bench(a []string, all bool, r runner, out, errw io.Writer) int {
	name := "bench"
	if all {
		name = "bench-all"
	}
	f, o := flags(name, errw)
	if f.Parse(a) != nil {
		return 64
	}
	if e := valid(o); e != nil {
		fmt.Fprintln(errw, e)
		return 64
	}
	m, e := load(o.manifest)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	z := selectJobs(m, o)
	if len(z) == 0 {
		fmt.Fprintln(errw, "no benchmark matches")
		return 65
	}
	if !all && len(z) != 1 {
		fmt.Fprintf(errw, "bench requires one match, found %d\n", len(z))
		return 65
	}
	if o.preflight {
		if e = invokePreflight(m, o, r, out, errw); e != nil {
			fmt.Fprintln(errw, e)
			return code(e)
		}
	}
	for i, x := range z {
		p := o.output
		if all {
			p = filepath.Join(p, fmt.Sprintf("%03d-%s.jsonl", i+1, safe(x.ID)))
		}
		if e = dispatch(x, o, p, r, out, errw); e != nil {
			fmt.Fprintf(errw, "%s failed: %v\n", x.ID, e)
			return code(e)
		}
	}
	return 0
}
func values(x Benchmark, o *options, p string) map[string]string {
	w, c := "", ""
	if o.warmups >= 0 {
		w = strconv.Itoa(o.warmups)
	}
	if o.core >= 0 {
		c = strconv.Itoa(o.core)
	}
	return map[string]string{"id": x.ID, "cipher": x.Cipher, "language": x.Language, "operation": x.Operation, "backend": x.Backend, "security_profile": x.SecurityProfile, "iterations": strconv.Itoa(o.iterations), "warmups": w, "core": c, "gc": o.gc, "policy": o.policy, "output": p}
}
func expand(s string, v map[string]string) string {
	for k, x := range v {
		s = strings.ReplaceAll(s, "{"+k+"}", x)
	}
	return s
}
func setenv(e []string, k, v string) []string {
	p := k + "="
	z := e[:0]
	for _, x := range e {
		if !strings.HasPrefix(x, p) {
			z = append(z, x)
		}
	}
	return append(z, p+v)
}
func dispatch(x Benchmark, o *options, p string, r runner, out, errw io.Writer) error {
	if !o.appendMode {
		if _, e := os.Stat(p); e == nil {
			return fmt.Errorf("output exists: %s", p)
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
	}
	if !o.dryRun {
		if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			return e
		}
	}
	v := values(x, o, p)
	a := make([]string, len(x.Args))
	for i, s := range x.Args {
		a[i] = expand(s, v)
	}
	e := os.Environ()
	vars := map[string]string{"HHE_BENCH_CIPHER": x.Cipher, "HHE_BENCH_LANGUAGE": x.Language, "HHE_BENCH_OPERATION": x.Operation, "HHE_BENCH_BACKEND": x.Backend, "HHE_BENCH_SECURITY_PROFILE": x.SecurityProfile, "HHE_BENCH_IMPLEMENTATION": x.ID, "HHE_BENCH_ITERATIONS": strconv.Itoa(o.iterations), "HHE_BENCH_GC": o.gc, "HHE_BENCH_POLICY": o.policy, "HHE_BENCH_OUTPUT": p}
	if o.warmups >= 0 {
		vars["HHE_BENCH_WARMUPS"] = strconv.Itoa(o.warmups)
	}
	if o.core >= 0 {
		vars["HHE_BENCH_CPU"] = strconv.Itoa(o.core)
	}
	for k, s := range x.Env {
		vars[k] = expand(s, v)
	}
	for k, s := range vars {
		e = setenv(e, k, s)
	}
	fmt.Fprintf(out, "dispatch id=%s output=%s\n", x.ID, p)
	if o.dryRun {
		fmt.Fprintf(out, "command=%s %s\n", x.Executable, strings.Join(a, " "))
		return nil
	}
	return r.Run(x.Executable, a, x.WorkDir, e, out, errw)
}

var bad = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func safe(s string) string {
	s = strings.Trim(bad.ReplaceAllString(s, "-"), "-.")
	if s == "" {
		return "benchmark"
	}
	return s
}
func invokePreflight(m Manifest, o *options, r runner, out, errw io.Writer) error {
	p := m.Preflight
	if p == "" {
		p = "/opt/hhe/bin/hhe-preflight"
	}
	var a []string
	if o.policy == "warn" {
		a = append(a, "--warn")
	}
	if o.core >= 0 {
		a = append(a, "--cpu", strconv.Itoa(o.core))
	}
	if o.dryRun {
		fmt.Fprintf(out, "preflight=%s %s\n", p, strings.Join(a, " "))
		return nil
	}
	return r.Run(p, a, "", os.Environ(), out, errw)
}
func preflight(a []string, r runner, out, errw io.Writer) int {
	f := flag.NewFlagSet("preflight", flag.ContinueOnError)
	f.SetOutput(errw)
	p := f.String("manifest", defaultManifest, "manifest")
	pol := f.String("policy", "strict", "policy")
	core := f.Int("core", -1, "CPU")
	if f.Parse(a) != nil || (*pol != "strict" && *pol != "warn") || *core < -1 {
		return 64
	}
	m, e := load(*p)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	e = invokePreflight(m, &options{policy: *pol, core: *core}, r, out, errw)
	if e != nil {
		fmt.Fprintln(errw, e)
		return code(e)
	}
	return 0
}
func list(a []string, out, errw io.Writer) int {
	f, o := flags("list", errw)
	if f.Parse(a) != nil {
		return 64
	}
	m, e := load(o.manifest)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	for _, x := range selectJobs(m, o) {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\n", x.ID, x.Cipher, x.Language, x.Operation, x.Backend, x.Executable)
	}
	return 0
}
func code(e error) int {
	var x *exec.ExitError
	if errors.As(e, &x) && x.ExitCode() > 0 {
		return x.ExitCode()
	}
	return 1
}
