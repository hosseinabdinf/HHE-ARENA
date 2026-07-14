# Windows and WSL2

Windows is supported for development, Docker image construction, CLI listing,
dry-runs, schema checks, and diagnostic smoke validation. It is not currently
the publication target for cycle-count comparisons.

## Requirements

- Windows 10/11 with Docker Desktop and the WSL2 backend;
- at least 16 GB RAM for tooling and native images; heavy HE/HHE images may
  require substantially more;
- Git, PowerShell, and Python 3.11+ for local validation;
- Docker Desktop configured with enough CPUs and memory for the selected image;
- repository checkout with Docker Desktop's Linux containers enabled.

## Setup and checks

```powershell
docker version
docker build -t hhe-arena:tooling .
docker run --rm hhe-arena:tooling probe
python tools/validate_framework_contract.py
```

Inspect available matrix choices without executing anything:

```powershell
hhe list --catalog manifests/images.v2.json --family all --path all --cipher all --variant all
hhe bench --catalog manifests/images.v2.json --family aes --path ske --cipher AES --variant all --dry-run --output results/dry-run
```

## What is safe on Windows

Use `hhe list`, `--dry-run`, manifest validation, image builds, and one-off
diagnostic runs with `--policy warn`. Always use a fresh output directory.
Do not run repository-wide tests or `bench-all` unless the workload and memory
requirements have been explicitly approved.

WSL commonly reports virtualized or incomplete TSC, governor, Turbo, and memory
metadata. JSONL and reports can still be validated, but cycle counts must be
treated as diagnostic and not used as publication-grade cross-host evidence.
