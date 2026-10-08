#!/usr/bin/env python3
"""Validate the filesystem portion of the HHE-ARENA framework-image contract."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

from jsonschema import Draft202012Validator


class ContractError(ValueError):
    pass


def _inside(root: Path, absolute_path: str) -> Path:
    if not absolute_path.startswith("/"):
        raise ContractError(f"contract path must be absolute: {absolute_path}")
    candidate = root / absolute_path.lstrip("/")
    if os.path.commonpath((str(root.resolve()), str(candidate.resolve()))) != str(root.resolve()):
        raise ContractError(f"contract path escapes image root: {absolute_path}")
    return candidate


def validate(manifest_path: Path, schema_path: Path, root: Path,
             require_benchmarks: bool, check_links: bool = False) -> int:
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    Draft202012Validator.check_schema(schema)
    Draft202012Validator(schema).validate(manifest)

    benchmarks = manifest["benchmarks"]
    if require_benchmarks and not benchmarks:
        raise ContractError("framework image manifest must contain at least one benchmark")
    ids = [entry["id"] for entry in benchmarks]
    if len(ids) != len(set(ids)):
        raise ContractError("benchmark IDs must be unique")

    preflight = manifest.get("preflight")
    if preflight != "/opt/hhe/bin/hhe-preflight":
        raise ContractError("preflight must be /opt/hhe/bin/hhe-preflight")
    preflight_file = _inside(root, preflight)
    if not preflight_file.is_file() or not os.access(preflight_file, os.X_OK):
        raise ContractError(f"preflight is missing or not executable: {preflight}")

    for entry in benchmarks:
        executable = entry["executable"]
        if not executable.startswith("/opt/hhe/frameworks/"):
            raise ContractError(f"benchmark executable is outside /opt/hhe/frameworks: {executable}")
        executable_file = _inside(root, executable)
        if not executable_file.is_file() or not os.access(executable_file, os.X_OK):
            raise ContractError(f"benchmark executable is missing or not executable: {executable}")
        workdir = entry.get("workdir", "/work")
        workdir_path = _inside(root, workdir)
        if not workdir_path.is_dir():
            raise ContractError(f"benchmark workdir is missing: {workdir}")

    if check_links:
        for executable in sorted({entry["executable"] for entry in benchmarks}):
            executable_file = _inside(root, executable)
            with executable_file.open("rb") as binary:
                is_elf = binary.read(4) == b"\x7fELF"
            if not is_elf:
                continue
            result = subprocess.run(["ldd", str(executable_file)],
                                    capture_output=True, text=True, check=False)
            output = result.stdout + result.stderr
            if "not a dynamic executable" in output:
                continue
            if result.returncode != 0 or "not found" in output:
                raise ContractError(f"unresolved shared library for {executable}: {output.strip()}")

    return len(benchmarks)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", default="/opt/hhe/benchmarks/manifest.json")
    parser.add_argument("--schema", default="/opt/hhe/schemas/benchmark-manifest.schema.json")
    parser.add_argument("--root", default="/", help="alternate image root for tests")
    parser.add_argument("--require-benchmarks", action="store_true")
    parser.add_argument("--check-links", action="store_true",
                        help="check ELF executable shared libraries without running benchmarks")
    args = parser.parse_args()
    try:
        count = validate(Path(args.manifest), Path(args.schema), Path(args.root),
                         args.require_benchmarks, args.check_links)
    except (ContractError, OSError, json.JSONDecodeError, ValueError) as error:
        print(f"framework image contract validation failed: {error}", file=sys.stderr)
        return 1
    print(f"framework image contract validation passed: {count} benchmark job(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
