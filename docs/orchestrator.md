# Benchmark orchestrator

The hhe command is the single entry point for prebuilt benchmark jobs.
Framework-specific images add executables and replace
/opt/hhe/benchmarks/manifest.json; the base tooling image intentionally ships
an empty manifest because it contains no HE framework binary.

Each manifest entry is one concrete cipher/language/operation/backend job.
bench requires exactly one selected entry. bench-all executes every selected
entry in stable ID order and writes a separate, untouched JSONL file for each
job. Existing files are rejected unless --append is explicit.

Examples:

    hhe list --manifest manifests/example.json
    hhe bench --manifest manifest.json --cipher hera --language rust --operation encrypt --backend tfhe --iterations 10 --warmup 2 --core 2 --gc not_applicable --policy strict --preflight --output results/hera-rust.jsonl
    hhe bench-all --manifest manifest.json --cipher hera --iterations 10 --output results/raw

Selectors accept comma-separated values. --dry-run prints dispatches without
starting binaries. The CLI exports the common HHE_BENCH_* variables for cipher,
language, operation, backend, security profile, implementation, iterations,
warmups, CPU, GC, policy, and output. Manifest env values override those
defaults. Arguments and environment values can use the placeholders documented
by hhe help.

Use hhe preflight --policy strict --core 2 for an explicit check, or add
--preflight to bench or bench-all. Preflight runs once before a sweep. A nonzero
preflight or benchmark exit stops dispatch and is propagated. Run the container
with --cpuset-cpus matching --core; the CLI pins through each language harness
but cannot set the container CPU allocation itself.

The manifest schema is schemas/benchmark-manifest.schema.json. IDs must be
unique. Framework images should validate their manifest, copy only prebuilt
executables into the runtime layer, and keep build tools out.