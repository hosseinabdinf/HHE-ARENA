#!/usr/bin/env python3
"""Static, workload-free checks for HHE-ARENA framework packaging."""

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def main() -> int:
    catalog_path = ROOT / "manifests" / "images.v2.json"
    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
    errors: list[str] = []
    images = catalog.get("images", [])
    if not images:
        errors.append("catalog contains no images")

    for image in images:
        ident = image.get("id", "<missing-id>")
        for field in ("framework", "language", "backend", "manifest", "build"):
            if not image.get(field):
                errors.append(f"{ident}: missing catalog field {field}")

        manifest_path = ROOT / image.get("manifest", "")
        if not manifest_path.is_file():
            errors.append(f"{ident}: manifest not found: {manifest_path.relative_to(ROOT)}")
        else:
            manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
            jobs = manifest.get("benchmarks", [])
            if manifest.get("version") != 1 or not jobs:
                errors.append(f"{ident}: manifest must be version 1 with benchmark jobs")
            for job in jobs:
                for field in ("id", "cipher", "language", "operation", "backend", "executable"):
                    if not job.get(field):
                        errors.append(f"{ident}: job missing field {field}")

        build = image.get("build", {})
        dockerfile = ROOT / build.get("dockerfile", "")
        if not dockerfile.is_file():
            errors.append(f"{ident}: Dockerfile not found: {dockerfile.relative_to(ROOT)}")
        else:
            text = dockerfile.read_text(encoding="utf-8")
            if "hhe-validate-framework-image" not in text:
                errors.append(f"{ident}: Dockerfile lacks contract validation")
            if "USER 65532:65532" not in text:
                errors.append(f"{ident}: Dockerfile lacks non-root runtime user")

    if errors:
        print("framework contract validation failed")
        print("\n".join(f"- {error}" for error in errors))
        return 1
    print(f"framework contract validation passed: {len(images)} catalog image(s)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
