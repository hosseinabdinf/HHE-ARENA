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
type BuildSpec struct {
	Tag        string            `json:"tag"`
	Dockerfile string            `json:"dockerfile"`
	Context    string            `json:"context"`
	Args       map[string]string `json:"args,omitempty"`
}
type ImageCatalog struct {
	Version       int         `json:"version"`
	MatrixVersion int         `json:"matrix_version"`
	Tooling       BuildSpec   `json:"tooling"`
	Images        []ImageSpec `json:"images"`
}
type ImageSpec struct {
	ID            string              `json:"id"`
	Framework     string              `json:"framework"`
	Cipher        string              `json:"cipher"`
	Language      string              `json:"language"`
	Backend       string              `json:"backend"`
	Operations    []string            `json:"operations"`
	MatrixTargets []ImageMatrixTarget `json:"matrix_targets,omitempty"`
	LocalTag      string              `json:"local_tag"`
	PullRef       string              `json:"pull_ref"`
	Manifest      string              `json:"manifest"`
	Build         BuildSpec           `json:"build"`
}
type ImageMatrixTarget struct {
	VariantID  string   `json:"variant_id"`
	Cipher     string   `json:"cipher"`
	Backend    string   `json:"backend"`
	Operations []string `json:"operations"`
}
type SupportMatrix struct {
	MatrixVersion     int                `json:"matrix_version"`
	SelectorContract  SelectorContract   `json:"selector_contract"`
	BenchmarkVariants []BenchmarkVariant `json:"benchmark_variants"`
}
type SelectorContract struct {
	Wildcard       string   `json:"wildcard"`
	SelectionOrder []string `json:"selection_order"`
}
type BenchmarkVariant struct {
	ID               string          `json:"id"`
	Family           string          `json:"family"`
	Cipher           string          `json:"cipher"`
	Variant          string          `json:"variant"`
	Framework        string          `json:"framework"`
	Language         string          `json:"language"`
	ImplementationID string          `json:"implementation_id"`
	Targets          []VariantTarget `json:"targets"`
}
type VariantTarget struct {
	Path             string   `json:"path"`
	Backends         []string `json:"backends"`
	Operations       []string `json:"operations"`
	Status           string   `json:"status"`
	ImplementationID string   `json:"implementation_id,omitempty"`
}
type matrixChoice struct {
	Variant   BenchmarkVariant
	Path      string
	Backend   string
	Operation string
	Status    string
	Image     *ImageSpec
}
type imageTarget struct {
	Image            ImageSpec
	Operation        string
	VariantID        string
	DispatchCipher   string
	DispatchBackend  string
	DispatchLanguage string
}
type options struct {
	manifest, catalog, matrix, projectRoot, imagePolicy, dockerBin string
	family, path, cipher, variant, framework, language             string
	operation, backend, gc, policy, output                         string
	hostID                                                         string
	iterations, warmups, core                                      int
	preflight, appendMode, dryRun                                  bool
}
type resultBundle struct {
	Root, Raw, Report string
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
	printTitle(w)
	fmt.Fprint(w, `Usage: hhe <bench|bench-all|list|preflight> [options]
Selectors: --family --path --cipher --variant --framework --backend
           --language --operation (comma-separated; "all" selects every value)
Benchmark: --iterations N --warmup N --core N --gc MODE --policy MODE
           --host-id ID --output PATH --preflight --append --dry-run
           --manifest FILE
Container: --catalog FILE --matrix FILE --project-root DIR
           --image-policy auto|local|build|pull --docker-bin PATH
           (--output is a bundle directory containing raw/ and report/)
Manifest argument templates: {iterations} {warmups} {core} {gc} {policy}
 {output} {id} {cipher} {language} {operation} {backend} {security_profile}
`)
}
func printTitle(w io.Writer) {
	var candidates []string
	if configured := strings.TrimSpace(os.Getenv("HHE_ARENA_TITLE_FILE")); configured != "" {
		candidates = append(candidates, configured)
	}
	if executable, e := os.Executable(); e == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "ARENA.txt"))
	}
	candidates = append(candidates, "/opt/hhe/ARENA.txt", "ARENA.txt", filepath.Join("..", "..", "ARENA.txt"))
	for _, candidate := range candidates {
		data, e := os.ReadFile(candidate)
		if e == nil && strings.TrimSpace(string(data)) != "" {
			fmt.Fprintln(w, strings.TrimRight(string(data), "\r\n"))
			return
		}
	}
	fmt.Fprintln(w, "HHE-ARENA")
}
func flags(name string, errw io.Writer) (*flag.FlagSet, *options) {
	o := &options{manifest: defaultManifest, iterations: 10, warmups: -1, core: -1, gc: "unknown", policy: "warn", output: "results/raw", hostID: defaultHostID()}
	if name == "bench" {
		o.output = "benchmark.jsonl"
	}
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(errw)
	f.StringVar(&o.manifest, "manifest", o.manifest, "manifest")
	f.StringVar(&o.catalog, "catalog", "", "framework image catalog (enables host Docker orchestration)")
	f.StringVar(&o.matrix, "matrix", "", "support matrix v2 (auto-discovered beside a v2 catalog)")
	f.StringVar(&o.projectRoot, "project-root", ".", "project root for catalog build paths")
	f.StringVar(&o.imagePolicy, "image-policy", "auto", "auto, local, build, or pull")
	f.StringVar(&o.dockerBin, "docker-bin", "docker", "Docker executable")
	f.StringVar(&o.family, "family", "", "family")
	f.StringVar(&o.path, "path", "", "benchmark path")
	f.StringVar(&o.cipher, "cipher", "", "cipher")
	f.StringVar(&o.variant, "variant", "", "cipher variant")
	f.StringVar(&o.framework, "framework", "", "framework")
	f.StringVar(&o.language, "language", "", "language")
	f.StringVar(&o.operation, "operation", "", "operation")
	f.StringVar(&o.backend, "backend", "", "backend")
	f.IntVar(&o.iterations, "iterations", 10, "iterations")
	f.IntVar(&o.warmups, "warmup", -1, "warmups")
	f.IntVar(&o.core, "core", -1, "CPU")
	f.StringVar(&o.gc, "gc", "unknown", "GC mode")
	f.StringVar(&o.policy, "policy", "warn", "policy")
	f.StringVar(&o.hostID, "host-id", o.hostID, "stable benchmark host identifier")
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
	if strings.TrimSpace(o.hostID) == "" {
		return errors.New("empty --host-id (set it explicitly when the system hostname is unavailable)")
	}
	if !match("auto,local,build,pull", o.imagePolicy) {
		return errors.New("--image-policy must be auto, local, build, or pull")
	}
	return nil
}

func defaultHostID() string {
	if configured := strings.TrimSpace(os.Getenv("HHE_ARENA_HOST_ID")); configured != "" {
		return configured
	}
	if inherited := strings.TrimSpace(os.Getenv("HHE_BENCH_HOSTNAME")); inherited != "" {
		return inherited
	}
	hostname, e := os.Hostname()
	if e != nil {
		return ""
	}
	return strings.TrimSpace(hostname)
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
func loadImageCatalog(p string) (ImageCatalog, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return ImageCatalog{}, fmt.Errorf("read image catalog: %w", e)
	}
	var c ImageCatalog
	if e = json.Unmarshal(b, &c); e != nil {
		return c, fmt.Errorf("parse image catalog: %w", e)
	}
	if c.Version != 1 || c.MatrixVersion < 1 {
		return c, errors.New("image catalog version must be 1 with a positive matrix_version")
	}
	if e = validBuild(c.Tooling); e != nil {
		return c, fmt.Errorf("invalid tooling build: %w", e)
	}
	seen := map[string]bool{}
	for i, x := range c.Images {
		if x.ID == "" || x.Framework == "" || x.Cipher == "" || x.Language == "" || x.Backend == "" || x.LocalTag == "" || len(x.Operations) == 0 {
			return c, fmt.Errorf("image %d missing required field", i)
		}
		if seen[x.ID] {
			return c, fmt.Errorf("duplicate image id %q", x.ID)
		}
		seen[x.ID] = true
		if e = validBuild(x.Build); e != nil {
			return c, fmt.Errorf("invalid build for %s: %w", x.ID, e)
		}
		ops := map[string]bool{}
		for _, op := range x.Operations {
			if op == "" || ops[op] {
				return c, fmt.Errorf("image %s has an empty or duplicate operation", x.ID)
			}
			ops[op] = true
		}
		targets := map[string]bool{}
		for _, target := range x.MatrixTargets {
			if target.VariantID == "" || target.Cipher == "" || target.Backend == "" || len(target.Operations) == 0 {
				return c, fmt.Errorf("image %s has an incomplete matrix target", x.ID)
			}
			targetKey := target.VariantID + "\x00" + strings.ToLower(target.Backend)
			if targets[targetKey] {
				return c, fmt.Errorf("image %s has duplicate matrix target %q for backend %q", x.ID, target.VariantID, target.Backend)
			}
			targets[targetKey] = true
			for _, op := range target.Operations {
				if op == "" {
					return c, fmt.Errorf("image %s matrix target %s has an empty operation", x.ID, target.VariantID)
				}
			}
		}
	}
	return c, nil
}
func loadSupportMatrix(p string) (SupportMatrix, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return SupportMatrix{}, fmt.Errorf("read support matrix: %w", e)
	}
	var m SupportMatrix
	if e = json.Unmarshal(b, &m); e != nil {
		return m, fmt.Errorf("parse support matrix: %w", e)
	}
	if m.MatrixVersion != 2 || len(m.BenchmarkVariants) == 0 {
		return m, errors.New("support matrix must be version 2 with benchmark_variants")
	}
	if m.SelectorContract.Wildcard != "all" {
		return m, errors.New("support matrix selector wildcard must be all")
	}
	wantOrder := []string{"family", "path", "cipher", "variant", "framework", "backend", "language"}
	if strings.Join(m.SelectorContract.SelectionOrder, ",") != strings.Join(wantOrder, ",") {
		return m, fmt.Errorf("unsupported selector order %q", strings.Join(m.SelectorContract.SelectionOrder, "/"))
	}
	seen := map[string]bool{}
	for i, v := range m.BenchmarkVariants {
		if v.ID == "" || v.Family == "" || v.Cipher == "" || v.Variant == "" || v.Framework == "" || v.Language == "" || v.ImplementationID == "" || len(v.Targets) == 0 {
			return m, fmt.Errorf("benchmark variant %d missing required field", i)
		}
		if seen[v.ID] {
			return m, fmt.Errorf("duplicate benchmark variant id %q", v.ID)
		}
		seen[v.ID] = true
		for _, target := range v.Targets {
			if target.Path == "" || len(target.Backends) == 0 || len(target.Operations) == 0 || target.Status == "" {
				return m, fmt.Errorf("benchmark variant %s has an incomplete target", v.ID)
			}
		}
	}
	return m, nil
}

func matrixPath(o *options, c ImageCatalog) string {
	if strings.TrimSpace(o.matrix) != "" {
		return o.matrix
	}
	if c.MatrixVersion == 2 {
		return filepath.Join(filepath.Dir(o.catalog), "support-matrix.v2.json")
	}
	return ""
}
func validBuild(b BuildSpec) error {
	if b.Tag == "" || b.Dockerfile == "" || b.Context == "" {
		return errors.New("tag, dockerfile, and context are required")
	}
	return nil
}
func match(s, v string) bool {
	if strings.TrimSpace(s) == "" {
		return true
	}
	for _, x := range strings.Split(s, ",") {
		if strings.TrimSpace(x) == "*" || strings.EqualFold(strings.TrimSpace(x), "all") || strings.EqualFold(strings.TrimSpace(x), v) {
			return true
		}
	}
	return false
}
func hasMatrixSelectors(o *options) bool {
	return strings.TrimSpace(o.family) != "" || strings.TrimSpace(o.path) != "" || strings.TrimSpace(o.variant) != "" || strings.TrimSpace(o.framework) != ""
}

func selectMatrixChoices(m SupportMatrix, c *ImageCatalog, o *options) []matrixChoice {
	images := map[string]*ImageSpec{}
	if c != nil {
		for i := range c.Images {
			images[c.Images[i].ID] = &c.Images[i]
		}
	}
	var z []matrixChoice
	for _, v := range m.BenchmarkVariants {
		if !match(o.family, v.Family) || !match(o.cipher, v.Cipher) || !match(o.variant, v.Variant) || !match(o.framework, v.Framework) || !match(o.language, v.Language) {
			continue
		}
		for _, target := range v.Targets {
			if !match(o.path, target.Path) {
				continue
			}
			implementationID := target.ImplementationID
			if implementationID == "" {
				implementationID = v.ImplementationID
			}
			for _, backend := range target.Backends {
				if !match(o.backend, backend) {
					continue
				}
				for _, operation := range target.Operations {
					if !match(o.operation, operation) {
						continue
					}
					z = append(z, matrixChoice{Variant: v, Path: target.Path, Backend: backend, Operation: operation, Status: target.Status, Image: images[implementationID]})
				}
			}
		}
	}
	sort.SliceStable(z, func(i, j int) bool {
		a, b := z[i], z[j]
		return strings.Join([]string{a.Variant.Family, a.Path, a.Variant.Cipher, a.Variant.Variant, a.Variant.Framework, a.Backend, a.Variant.Language, a.Operation}, "\x00") < strings.Join([]string{b.Variant.Family, b.Path, b.Variant.Cipher, b.Variant.Variant, b.Variant.Framework, b.Backend, b.Variant.Language, b.Operation}, "\x00")
	})
	return z
}

func runnableMatrixTargets(choices []matrixChoice) []imageTarget {
	var z []imageTarget
	seen := map[string]bool{}
	for _, choice := range choices {
		if choice.Status != "ready" || choice.Image == nil {
			continue
		}
		if len(choice.Image.MatrixTargets) > 0 {
			for _, advertised := range choice.Image.MatrixTargets {
				if advertised.VariantID != choice.Variant.ID {
					continue
				}
				for _, operation := range advertised.Operations {
					if operation == choice.Operation && strings.EqualFold(advertised.Backend, choice.Backend) {
						key := choice.Variant.ID + "\x00" + strings.ToLower(choice.Backend) + "\x00" + operation
						if !seen[key] {
							z = append(z, imageTarget{Image: *choice.Image, Operation: operation, VariantID: choice.Variant.ID, DispatchCipher: advertised.Cipher, DispatchBackend: advertised.Backend, DispatchLanguage: choice.Image.Language})
							seen[key] = true
						}
					}
				}
			}
			continue
		}
		for _, operation := range choice.Image.Operations {
			if operation != choice.Operation {
				continue
			}
			key := choice.Variant.ID + "\x00" + operation
			if !seen[key] {
				z = append(z, imageTarget{Image: *choice.Image, Operation: operation, VariantID: choice.Variant.ID})
				seen[key] = true
			}
		}
	}
	return z
}
func selectImageTargets(c ImageCatalog, o *options) []imageTarget {
	var z []imageTarget
	for _, x := range c.Images {
		if !match(o.cipher, x.Cipher) || !match(o.language, x.Language) || !match(o.backend, x.Backend) {
			continue
		}
		for _, op := range x.Operations {
			if match(o.operation, op) {
				z = append(z, imageTarget{Image: x, Operation: op})
			}
		}
	}
	sort.SliceStable(z, func(i, j int) bool {
		if z[i].Image.ID == z[j].Image.ID {
			return z[i].Operation < z[j].Operation
		}
		return z[i].Image.ID < z[j].Image.ID
	})
	return z
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
	if o.catalog != "" {
		return containerBench(all, o, r, out, errw)
	}
	if o.matrix != "" || hasMatrixSelectors(o) {
		fmt.Fprintln(errw, "matrix-v2 selectors require --catalog so only runnable image targets can execute")
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
	if !all {
		o.output = jsonlPath(o.output)
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
func jsonlPath(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".jsonl") {
		return path
	}
	return path + ".jsonl"
}

func containerBench(all bool, o *options, r runner, out, errw io.Writer) int {
	c, e := loadImageCatalog(o.catalog)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	var z []imageTarget
	if p := matrixPath(o, c); p != "" {
		m, matrixErr := loadSupportMatrix(p)
		if matrixErr != nil {
			fmt.Fprintln(errw, matrixErr)
			return 66
		}
		choices := selectMatrixChoices(m, &c, o)
		if len(choices) == 0 {
			fmt.Fprintf(errw, "no matrix-v2 benchmark matches family=%q path=%q cipher=%q variant=%q framework=%q backend=%q language=%q operation=%q\n", o.family, o.path, o.cipher, o.variant, o.framework, o.backend, o.language, o.operation)
			return 65
		}
		z = runnableMatrixTargets(choices)
		if len(z) == 0 {
			fmt.Fprintf(errw, "matched %d matrix-v2 choice(s), but none are runnable; use list --catalog %s with the same selectors to inspect deferred, blocked, or uncataloged targets\n", len(choices), o.catalog)
			return 65
		}
		if all && len(z) < len(choices) {
			fmt.Fprintf(out, "matrix-selection=%d runnable=%d skipped=%d\n", len(choices), len(z), len(choices)-len(z))
		}
	} else {
		if hasMatrixSelectors(o) {
			fmt.Fprintln(errw, "--family, --path, --variant, and --framework require a matrix-v2 catalog or explicit --matrix")
			return 64
		}
		z = selectImageTargets(c, o)
	}
	if len(z) == 0 {
		fmt.Fprintf(errw, "no container benchmark matches cipher=%q language=%q operation=%q backend=%q; use list --catalog %s to inspect supported combinations\n", o.cipher, o.language, o.operation, o.backend, o.catalog)
		return 65
	}
	if !all && len(z) != 1 {
		fmt.Fprintf(errw, "bench requires one container match, found %d\n", len(z))
		return 65
	}
	root, e := filepath.Abs(o.projectRoot)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	bundle, e := newResultBundle(o.output)
	if e != nil {
		fmt.Fprintln(errw, e)
		return 66
	}
	toolingBuilt := false
	acquired := map[string]bool{}
	rawOutputs := make([]string, 0, len(z))
	for i, t := range z {
		if !acquired[t.Image.ID] {
			if e = acquireImage(c.Tooling, t.Image, o, root, &toolingBuilt, r, out, errw); e != nil {
				fmt.Fprintf(errw, "%s image acquisition failed: %v\n", t.Image.ID, e)
				return code(e)
			}
			acquired[t.Image.ID] = true
		}
		targetID := t.Image.ID
		if t.VariantID != "" {
			targetID = t.VariantID
		}
		hostOutput := filepath.Join(bundle.Raw, fmt.Sprintf("%03d-%s-%s.jsonl", i+1, safe(targetID), safe(t.Operation)))
		if e = runContainer(t, o, hostOutput, r, out, errw); e != nil {
			fmt.Fprintf(errw, "%s failed: %v\n", t.Image.ID, e)
			return code(e)
		}
		rawOutputs = append(rawOutputs, hostOutput)
	}
	if e = runContainerReport(z[0].Image.LocalTag, bundle, rawOutputs, o, r, out, errw); e != nil {
		fmt.Fprintf(errw, "report generation failed: %v\n", e)
		return code(e)
	}
	fmt.Fprintf(out, "result-bundle=%s\n", bundle.Root)
	return 0
}

func newResultBundle(root string) (resultBundle, error) {
	absolute, e := filepath.Abs(root)
	if e != nil {
		return resultBundle{}, fmt.Errorf("resolve result bundle: %w", e)
	}
	if info, statErr := os.Stat(absolute); statErr == nil && !info.IsDir() {
		return resultBundle{}, fmt.Errorf("result bundle is not a directory: %s", absolute)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return resultBundle{}, fmt.Errorf("inspect result bundle: %w", statErr)
	}
	return resultBundle{Root: absolute, Raw: filepath.Join(absolute, "raw"), Report: filepath.Join(absolute, "report")}, nil
}

func runContainerReport(image string, bundle resultBundle, rawOutputs []string, o *options, r runner, out, errw io.Writer) error {
	if !o.dryRun {
		if e := os.MkdirAll(bundle.Report, 0755); e != nil {
			return e
		}
	}
	a := []string{"run", "--rm", "--mount", "type=bind,source=" + bundle.Root + ",target=/work/results", image, "report"}
	for _, raw := range rawOutputs {
		relative, e := filepath.Rel(bundle.Root, raw)
		if e != nil || strings.HasPrefix(relative, "..") {
			return fmt.Errorf("raw result is outside bundle: %s", raw)
		}
		a = append(a, "/work/results/"+filepath.ToSlash(relative))
	}
	a = append(a, "--output-dir", "/work/results/report", "--schema", "/opt/hhe/schemas/benchmark.schema.json")
	return dockerCommand(o, "", a, r, out, errw)
}

func acquireImage(tooling BuildSpec, x ImageSpec, o *options, root string, toolingBuilt *bool, r runner, out, errw io.Writer) error {
	policy := o.imagePolicy
	if policy == "auto" && o.dryRun {
		fmt.Fprintf(out, "image-check=%s image inspect %s\n", o.dockerBin, x.LocalTag)
		if x.PullRef != "" {
			fmt.Fprintln(out, "image-fallback=pull")
			if e := dockerCommand(o, root, []string{"pull", x.PullRef}, r, out, errw); e != nil {
				return e
			}
			return dockerCommand(o, root, []string{"tag", x.PullRef, x.LocalTag}, r, out, errw)
		} else {
			fmt.Fprintln(out, "image-fallback=build")
			return buildImage(tooling, x, o, root, toolingBuilt, r, out, errw)
		}
	}
	if policy == "auto" {
		if e := r.Run(o.dockerBin, []string{"image", "inspect", x.LocalTag}, root, os.Environ(), io.Discard, io.Discard); e == nil {
			fmt.Fprintf(out, "image=%s source=local\n", x.LocalTag)
			return nil
		}
		if x.PullRef != "" {
			policy = "pull"
		} else {
			policy = "build"
		}
	}
	switch policy {
	case "local":
		if o.dryRun {
			fmt.Fprintf(out, "image-check=%s image inspect %s\n", o.dockerBin, x.LocalTag)
			return nil
		}
		if e := r.Run(o.dockerBin, []string{"image", "inspect", x.LocalTag}, root, os.Environ(), io.Discard, errw); e != nil {
			return e
		}
		fmt.Fprintf(out, "image=%s source=local\n", x.LocalTag)
		return nil
	case "pull":
		if x.PullRef == "" {
			return fmt.Errorf("no immutable pull_ref is published for %s; use auto, local, or build", x.ID)
		}
		if e := dockerCommand(o, root, []string{"pull", x.PullRef}, r, out, errw); e != nil {
			return e
		}
		return dockerCommand(o, root, []string{"tag", x.PullRef, x.LocalTag}, r, out, errw)
	case "build":
		return buildImage(tooling, x, o, root, toolingBuilt, r, out, errw)
	default:
		return fmt.Errorf("unsupported image policy %q", policy)
	}
}
func buildImage(tooling BuildSpec, x ImageSpec, o *options, root string, toolingBuilt *bool, r runner, out, errw io.Writer) error {
	if !*toolingBuilt {
		if e := dockerCommand(o, root, buildArgs(tooling, root), r, out, errw); e != nil {
			return fmt.Errorf("tooling build: %w", e)
		}
		*toolingBuilt = true
	}
	if e := dockerCommand(o, root, buildArgs(x.Build, root), r, out, errw); e != nil {
		return fmt.Errorf("framework build: %w", e)
	}
	return nil
}
func buildArgs(b BuildSpec, root string) []string {
	a := []string{"build", "--pull=false", "-t", b.Tag, "-f", filepath.Join(root, b.Dockerfile)}
	keys := make([]string, 0, len(b.Args))
	for k := range b.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		a = append(a, "--build-arg", k+"="+b.Args[k])
	}
	return append(a, filepath.Join(root, b.Context))
}
func dockerCommand(o *options, root string, a []string, r runner, out, errw io.Writer) error {
	fmt.Fprintf(out, "docker-command=%s %s\n", o.dockerBin, strings.Join(a, " "))
	if o.dryRun {
		return nil
	}
	return r.Run(o.dockerBin, a, root, os.Environ(), out, errw)
}
func runContainer(t imageTarget, o *options, hostOutput string, r runner, out, errw io.Writer) error {
	absolute, e := filepath.Abs(hostOutput)
	if e != nil {
		return e
	}
	hostDir, containerOutput := filepath.Dir(absolute), "/work/results/"+filepath.Base(absolute)
	if !o.dryRun {
		if e = os.MkdirAll(hostDir, 0755); e != nil {
			return e
		}
	}
	a := []string{"run", "--rm", "--env", "HHE_BENCH_HOSTNAME=" + o.hostID}
	if o.core >= 0 {
		a = append(a, "--cpuset-cpus", strconv.Itoa(o.core))
	}
	cipher, language, backend := t.Image.Cipher, t.Image.Language, t.Image.Backend
	if t.DispatchCipher != "" {
		cipher = t.DispatchCipher
	}
	if t.DispatchLanguage != "" {
		language = t.DispatchLanguage
	}
	if t.DispatchBackend != "" {
		backend = t.DispatchBackend
	}
	a = append(a, "--mount", "type=bind,source="+hostDir+",target=/work/results", t.Image.LocalTag, "bench",
		"--cipher", cipher, "--language", language, "--operation", t.Operation, "--backend", backend,
		"--iterations", strconv.Itoa(o.iterations), "--gc", o.gc, "--policy", o.policy, "--host-id", o.hostID, "--output", containerOutput)
	if o.warmups >= 0 {
		a = append(a, "--warmup", strconv.Itoa(o.warmups))
	}
	if o.core >= 0 {
		a = append(a, "--core", strconv.Itoa(o.core))
	}
	if o.preflight {
		a = append(a, "--preflight")
	}
	if o.appendMode {
		a = append(a, "--append")
	}
	return dockerCommand(o, "", a, r, out, errw)
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
	vars := map[string]string{"HHE_BENCH_CIPHER": x.Cipher, "HHE_BENCH_LANGUAGE": x.Language, "HHE_BENCH_OPERATION": x.Operation, "HHE_BENCH_BACKEND": x.Backend, "HHE_BENCH_SECURITY_PROFILE": x.SecurityProfile, "HHE_BENCH_IMPLEMENTATION": x.ID, "HHE_BENCH_ITERATIONS": strconv.Itoa(o.iterations), "HHE_BENCH_GC": o.gc, "HHE_BENCH_POLICY": o.policy, "HHE_BENCH_OUTPUT": p, "HHE_BENCH_HOSTNAME": o.hostID}
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
	if o.catalog != "" {
		c, e := loadImageCatalog(o.catalog)
		if e != nil {
			fmt.Fprintln(errw, e)
			return 66
		}
		if p := matrixPath(o, c); p != "" {
			m, matrixErr := loadSupportMatrix(p)
			if matrixErr != nil {
				fmt.Fprintln(errw, matrixErr)
				return 66
			}
			printMatrixChoices(selectMatrixChoices(m, &c, o), out)
			return 0
		}
		if hasMatrixSelectors(o) {
			fmt.Fprintln(errw, "--family, --path, --variant, and --framework require a matrix-v2 catalog or explicit --matrix")
			return 64
		}
		for _, t := range selectImageTargets(c, o) {
			pull := t.Image.PullRef
			if pull == "" {
				pull = "unpublished"
			}
			fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", t.Image.ID, t.Image.Cipher, t.Image.Language, t.Operation, t.Image.Backend, t.Image.LocalTag, pull)
		}
		return 0
	}
	if o.matrix != "" {
		m, e := loadSupportMatrix(o.matrix)
		if e != nil {
			fmt.Fprintln(errw, e)
			return 66
		}
		printMatrixChoices(selectMatrixChoices(m, nil, o), out)
		return 0
	}
	if hasMatrixSelectors(o) {
		fmt.Fprintln(errw, "matrix-v2 selectors require --matrix or a matrix-v2 --catalog")
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
func printMatrixChoices(choices []matrixChoice, out io.Writer) {
	for _, choice := range choices {
		execution := "not-runnable"
		if len(runnableMatrixTargets([]matrixChoice{choice})) == 1 {
			execution = "runnable"
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			choice.Variant.ID, choice.Variant.Family, choice.Path, choice.Variant.Cipher,
			choice.Variant.Variant, choice.Variant.Framework, choice.Backend,
			choice.Variant.Language, choice.Operation, choice.Status, execution)
	}
}
func code(e error) int {
	var x *exec.ExitError
	if errors.As(e, &x) && x.ExitCode() > 0 {
		return x.ExitCode()
	}
	return 1
}
