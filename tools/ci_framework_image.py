#!/usr/bin/env python3
"""Build a catalog image and optionally run one authorized bounded smoke."""

import argparse
from contextlib import ExitStack
import json
import os
from pathlib import Path
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parents[1]


def run(*args, timeout=3600):
    print("+", " ".join(map(str, args)), flush=True)
    subprocess.run(args, cwd=ROOT, check=True, timeout=timeout)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("image_id")
    parser.add_argument("--smoke", action="store_true")
    parser.add_argument("--heavy-only", action="store_true")
    parser.add_argument("--output-dir", type=Path,
                        help="keep smoke JSONL and report here (default: temporary)")
    args = parser.parse_args()
    catalog = json.loads((ROOT / "manifests/images.v2.json").read_text())
    matrix = json.loads((ROOT / "manifests/support-matrix.v2.json").read_text())
    images = {entry["id"]: entry for entry in catalog["images"]}
    implementations = {entry["id"]: entry for entry in matrix["implementations"]}
    if args.image_id not in images:
        parser.error(f"unknown catalog image: {args.image_id}")
    image = images[args.image_id]
    implementation = implementations[args.image_id]
    if args.smoke and implementation["readiness_class"] != "safe_lightweight":
        parser.error(f"{args.image_id} is not a safe lightweight implementation")
    if args.heavy_only and implementation["readiness_class"] == "safe_lightweight":
        parser.error(f"{args.image_id} belongs in the lightweight CI job")

    tooling = catalog["tooling"]
    source_revision = subprocess.check_output(
        ["git", "rev-parse", "HEAD"], text=True, cwd=ROOT).strip()
    working_tree_dirty = bool(subprocess.check_output(
        ["git", "status", "--porcelain", "--untracked-files=all"],
        text=True, cwd=ROOT).strip())
    run("docker", "build", "--pull=false", "-f", tooling["dockerfile"],
        "-t", tooling["tag"], tooling["context"])
    build = image["build"]
    command = ["docker", "build", "--pull=false", "-f", build["dockerfile"],
               "-t", build["tag"]]
    build_args = dict(build.get("args", {}))
    build_args["IMAGE_REVISION"] = source_revision
    for name, value in sorted(build_args.items()):
        command += ["--build-arg", f"{name}={value}"]
    command.append(build["context"])
    run(*command)
    run("docker", "run", "--rm", "--network", "none", image["local_tag"],
        "list", timeout=60)
    run("docker", "run", "--rm", "--network", "none", image["local_tag"],
        "hhe-validate-framework-image", "--require-benchmarks", "--check-links",
        timeout=120)
    if args.output_dir is not None:
        result_dir = args.output_dir.resolve()
        result_dir.mkdir(parents=True, exist_ok=True)
        image_id = subprocess.check_output(
            ["docker", "image", "inspect", "--format", "{{.Id}}", image["local_tag"]],
            text=True, cwd=ROOT).strip()
        metadata = {"catalog_id": args.image_id, "local_image_id": image_id,
                    "source_revision": source_revision,
                    "working_tree_dirty": working_tree_dirty,
                    "manifest": image["manifest"]}
        (result_dir / "metadata.json").write_text(
            json.dumps(metadata, indent=2) + "\n", encoding="utf-8")

    if not args.smoke:
        return 0
    manifest = json.loads((ROOT / image["manifest"]).read_text())
    safe_operations = {entry["name"] for entry in implementation["operations"]
                       if entry.get("safe_smoke")}
    counts = {}
    for job in manifest["benchmarks"]:
        key = (job["cipher"], job["language"], job["operation"], job["backend"])
        counts[key] = counts.get(key, 0) + 1
    job = next((entry for entry in manifest["benchmarks"]
                if entry["operation"] in safe_operations
                and counts[(entry["cipher"], entry["language"],
                            entry["operation"], entry["backend"])] == 1
                and not entry.get("disabled")), None)
    if job is None:
        raise RuntimeError(f"{args.image_id}: no unique safe smoke manifest job")

    with ExitStack() as stack:
        if args.output_dir is None:
            result_dir = Path(stack.enter_context(
                tempfile.TemporaryDirectory(prefix="hhe-ci-")))
        else:
            result_dir = args.output_dir.resolve()
            result_dir.mkdir(parents=True, exist_ok=True)
        mount = f"type=bind,source={result_dir},target=/work/results"
        common = ["docker", "run", "--rm", "--network", "none",
                  "--user", f"{os.getuid()}:{os.getgid()}", "--mount", mount,
                  image["local_tag"]]
        run(*common, "bench", "--cipher", job["cipher"], "--language",
            job["language"], "--operation", job["operation"], "--backend",
            job["backend"], "--iterations", "1", "--warmup", "0",
            "--policy", "warn", "--output", "/work/results/raw.jsonl",
            timeout=600)
        run(*common, "report", "/work/results/raw.jsonl", "--output-dir",
            "/work/results/report", "--schema",
            "/opt/hhe/schemas/benchmark.schema.json", timeout=120)
        if not (result_dir / "report/report.csv").is_file():
            raise RuntimeError(f"{args.image_id}: report.csv was not produced")
        print(f"smoke results: {result_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
