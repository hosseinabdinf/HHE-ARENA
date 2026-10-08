#!/usr/bin/env python3
"""Check v2 matrix, image catalog, and embedded manifest relationships."""

import json
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MATRIX = ROOT / "manifests/support-matrix.v2.json"
CATALOG = ROOT / "manifests/images.v2.json"


def load_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def keyed(items, kind, errors):
    result = {}
    for item in items:
        ident = item.get("id")
        if not ident:
            errors.append(f"{kind} has a missing id")
        elif ident in result:
            errors.append(f"duplicate {kind} id: {ident}")
        else:
            result[ident] = item
    return result


def target_implementation(variant, target):
    return target.get("implementation_id", variant["implementation_id"])


def main():
    errors = []
    matrix = load_json(MATRIX)
    catalog = load_json(CATALOG)
    if matrix.get("matrix_version") != 2 or catalog.get("matrix_version") != 2:
        errors.append("matrix and catalog must both declare matrix_version 2")

    implementations = keyed(matrix["implementations"], "implementation", errors)
    variants = keyed(matrix["benchmark_variants"], "variant", errors)
    images = keyed(catalog["images"], "image", errors)
    manifests = {}

    for ident, implementation in implementations.items():
        selection = implementation["selection_status"]
        docker = implementation["docker_status"]
        manifest = implementation["manifest_status"]
        cataloged = ident in images
        if cataloged != (docker in {"ready", "built"} and manifest == "ready"):
            errors.append(
                f"{ident}: catalog presence disagrees with docker_status={docker} "
                f"and manifest_status={manifest}"
            )
        if cataloged and selection not in {"selected", "packaged_deferred"}:
            errors.append(f"{ident}: cataloged image has selection_status={selection}")

    for ident, image in images.items():
        if ident not in implementations:
            errors.append(f"{ident}: catalog image has no implementation")
        if image["build"]["tag"] != image["local_tag"]:
            errors.append(f"{ident}: build tag differs from local_tag")
        path = ROOT / image["manifest"]
        if not path.is_file():
            errors.append(f"{ident}: missing manifest {image['manifest']}")
            continue
        manifest = load_json(path)
        if manifest.get("version") != 1:
            errors.append(f"{ident}: embedded manifest is not version 1")
        jobs = keyed(manifest.get("benchmarks", []), f"{ident} manifest job", errors)
        manifests[ident] = list(jobs.values())
        if not jobs:
            errors.append(f"{ident}: manifest has no jobs")
        job_operations = {job["operation"] for job in jobs.values()}
        if set(image["operations"]) != job_operations:
            errors.append(
                f"{ident}: catalog operations differ from manifest operations: "
                f"catalog={sorted(image['operations'])}, manifest={sorted(job_operations)}"
            )

    referenced = set()
    for variant in variants.values():
        for target in variant["targets"]:
            ident = target_implementation(variant, target)
            referenced.add(ident)
            if ident not in implementations:
                errors.append(f"{variant['id']}: unknown implementation {ident}")
                continue
            if target["status"] != "ready":
                continue
            image = images.get(ident)
            if image is None:
                errors.append(f"{variant['id']}: ready target has no catalog image {ident}")
                continue
            advertised = image.get("matrix_targets") or [
                {"variant_id": variant["id"], "cipher": image["cipher"],
                 "backend": image["backend"], "operations": image["operations"]}
            ]
            for backend in target["backends"]:
                for operation in target["operations"]:
                    matches = [
                        entry for entry in advertised
                        if entry["variant_id"] == variant["id"]
                        and entry["backend"] == backend
                        and operation in entry["operations"]
                    ]
                    if not matches:
                        errors.append(
                            f"{variant['id']}: ready {backend}/{operation} is not advertised "
                            f"by {ident}"
                        )
                    elif not any(
                        job["cipher"] == entry["cipher"]
                        and job["backend"] == entry["backend"]
                        and job["language"] == image["language"]
                        and job["operation"] == operation
                        for entry in matches for job in manifests.get(ident, [])
                    ):
                        errors.append(
                            f"{variant['id']}: ready {backend}/{operation} has no "
                            f"matching manifest job in {ident}"
                        )

    for ident, image in images.items():
        if ident not in referenced:
            errors.append(f"{ident}: catalog image is not referenced by any matrix target")
        matrix_target_operations = set()
        for entry in image.get("matrix_targets", []):
            variant = variants.get(entry["variant_id"])
            if variant is None:
                errors.append(f"{ident}: unknown matrix variant {entry['variant_id']}")
                continue
            for operation in entry["operations"]:
                key = (entry["variant_id"], entry["backend"].casefold(), operation)
                if key in matrix_target_operations:
                    errors.append(
                        f"{ident}: duplicate matrix target operation "
                        f"{entry['variant_id']}/{entry['backend']}/{operation}"
                    )
                matrix_target_operations.add(key)
                if not any(
                    target_implementation(variant, target) == ident
                    and entry["backend"] in target["backends"]
                    and operation in target["operations"]
                    for target in variant["targets"]
                ):
                    errors.append(
                        f"{ident}: {entry['variant_id']} maps {entry['backend']}/{operation} "
                        "to no matrix target"
                    )
                if not any(
                    job["cipher"] == entry["cipher"]
                    and job["backend"] == entry["backend"]
                    and job["operation"] == operation
                    for job in manifests.get(ident, [])
                ):
                    errors.append(
                        f"{ident}: {entry['variant_id']} maps {entry['backend']}/{operation} "
                        "to no manifest job"
                    )

    if errors:
        print("project inventory validation failed")
        for error in errors:
            print(f"- {error}")
        return 1
    statuses = Counter(target["status"] for variant in variants.values()
                       for target in variant["targets"])
    print(
        f"project inventory validation passed: {len(implementations)} implementations, "
        f"{len(variants)} variants, {len(images)} images, "
        f"{statuses['ready']} ready targets"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
