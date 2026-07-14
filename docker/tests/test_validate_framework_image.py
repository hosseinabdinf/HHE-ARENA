import importlib.util
import json
from pathlib import Path
import tempfile
import unittest


SCRIPT = Path(__file__).parents[1] / "validate_framework_image.py"
SPEC = importlib.util.spec_from_file_location("validate_framework_image", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class FrameworkImageValidatorTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.schema = Path(__file__).parents[2] / "schemas" / "benchmark-manifest.schema.json"
        for relative in ("opt/hhe/bin", "opt/hhe/frameworks/demo", "work"):
            (self.root / relative).mkdir(parents=True, exist_ok=True)
        for relative in ("opt/hhe/bin/hhe-preflight", "opt/hhe/frameworks/demo/bench"):
            path = self.root / relative
            path.write_text("#!/bin/sh\n", encoding="utf-8")
            path.chmod(0o555)
        self.manifest = self.root / "manifest.json"
        self.data = {
            "version": 1,
            "preflight": "/opt/hhe/bin/hhe-preflight",
            "benchmarks": [{
                "id": "demo-c-encrypt",
                "cipher": "demo",
                "language": "c",
                "operation": "encrypt",
                "backend": "native",
                "executable": "/opt/hhe/frameworks/demo/bench",
                "workdir": "/work"
            }]
        }

    def tearDown(self):
        self.temp.cleanup()

    def write_manifest(self):
        self.manifest.write_text(json.dumps(self.data), encoding="utf-8")

    def test_accepts_contract_layout(self):
        self.write_manifest()
        self.assertEqual(MODULE.validate(self.manifest, self.schema, self.root, True), 1)

    def test_rejects_missing_executable(self):
        self.data["benchmarks"][0]["executable"] = "/opt/hhe/frameworks/demo/missing"
        self.write_manifest()
        with self.assertRaisesRegex(MODULE.ContractError, "missing or not executable"):
            MODULE.validate(self.manifest, self.schema, self.root, True)

    def test_rejects_duplicate_ids(self):
        self.data["benchmarks"].append(dict(self.data["benchmarks"][0]))
        self.write_manifest()
        with self.assertRaisesRegex(MODULE.ContractError, "IDs must be unique"):
            MODULE.validate(self.manifest, self.schema, self.root, True)


if __name__ == "__main__":
    unittest.main()
