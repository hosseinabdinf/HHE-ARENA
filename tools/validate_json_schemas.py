#!/usr/bin/env python3
"""Validate catalog, matrix, and packaged benchmark manifests against JSON Schema."""

import json
from pathlib import Path

try:
    from jsonschema import Draft202012Validator, FormatChecker
except ImportError as error:
    raise SystemExit("install jsonschema==4.23.0 or run this script in hhe-arena:tooling") from error


ROOT = Path(__file__).resolve().parents[1]
SCHEMAS = ROOT / "schemas"
MANIFESTS = ROOT / "manifests"


def main() -> int:
    checks = [
        ("image-catalog.schema.json", sorted(MANIFESTS.glob("images.v*.json"))),
        ("support-matrix.schema.json", sorted(MANIFESTS.glob("support-matrix.v*.json"))),
        (
            "benchmark-manifest.schema.json",
            sorted(
                path for path in MANIFESTS.glob("*.json")
                if not path.name.startswith(("images.", "support-matrix."))
            ),
        ),
    ]
    errors = []
    count = 0
    for schema_name, paths in checks:
        schema = json.loads((SCHEMAS / schema_name).read_text(encoding="utf-8"))
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema, format_checker=FormatChecker())
        for path in paths:
            instance = json.loads(path.read_text(encoding="utf-8"))
            count += 1
            for issue in validator.iter_errors(instance):
                location = "/".join(map(str, issue.absolute_path)) or "<root>"
                errors.append(f"{path.relative_to(ROOT)}:{location}: {issue.message}")
    if errors:
        print("JSON Schema validation failed:")
        for error in errors:
            print(f"- {error}")
        return 1
    print(f"JSON Schema validation passed: {count} document(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
