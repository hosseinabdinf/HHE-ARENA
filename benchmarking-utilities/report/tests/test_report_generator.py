import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).parents[1] / "report_generator.py"
SPEC = importlib.util.spec_from_file_location("report_generator", MODULE_PATH)
report = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
sys.modules[SPEC.name] = report
SPEC.loader.exec_module(report)


def record(index, operation="encrypt", phase="measured", status="ok", memory=100):
    return {
        "schema_version": "1.0.0", "run_id": "run", "timestamp_utc": "2026-07-12T12:00:00Z",
        "benchmark": {"cipher": "hera", "operation": operation, "backend": "tfhe", "security_profile": "HHE-128", "parameters": {}},
        "runtime": {"language": "rust", "implementation": "test", "version": None, "gc_mode": "not_applicable", "allocator": "system_glibc"},
        "host": {"hostname": "host", "os": "linux", "arch": "x86_64", "cpu_model": None, "logical_cpus": 8},
        "environment": {"policy": "warn", "requested_cpu": 1, "actual_cpu": 1, "affinity_verified": True, "rdtscp_supported": True, "invariant_tsc": True, "nonstop_tsc": True, "governor": "performance", "turbo_enabled": False, "checks_passed": True, "warnings": []},
        "iteration": {"phase": phase, "index": index},
        "measurement": {"time_ns": 10 + index, "cycles": 20 + index, "memory": {"rss_bytes_delta": memory, "heap_bytes_delta": memory, "peak_rss_bytes_delta": memory, "ru_maxrss_bytes": 1_000_000 + memory}},
        "status": status, "error": None if status == "ok" else "failure",
    }


class ReportTests(unittest.TestCase):
    def test_schema_accepts_legacy_and_ru_maxrss_records(self):
        root = Path(__file__).parents[3]
        schema = json.loads((root / "schemas" / "benchmark.schema.json").read_text(encoding="utf-8"))
        validator = report.Draft202012Validator(schema)
        current = record(0)
        legacy = json.loads(json.dumps(current))
        legacy["measurement"]["memory"].pop("ru_maxrss_bytes")
        self.assertEqual(list(validator.iter_errors(current)), [])
        self.assertEqual(list(validator.iter_errors(legacy)), [])
    def test_low_n_warmup_exclusion_and_baseline(self):
        values = [record(0, "baseline", memory=10), record(1, "baseline", memory=20), record(0, phase="warmup", memory=999), record(0, memory=100), record(1, memory=120)]
        loaded = [report.LoadedRecord("test", index + 1, value) for index, value in enumerate(values)]
        rows = report.summarize(loaded, low_n_threshold=5, baseline_operation="baseline")
        encrypt = next(row for row in rows if row["operation"] == "encrypt")
        self.assertEqual(encrypt["n"], 2)
        self.assertEqual(encrypt["statistics_mode"], "low_n_median_min_max")
        self.assertIsNone(encrypt["time_ns_iqr"])
        self.assertEqual(encrypt["rss_bytes_delta_baseline_subtracted"], 95)
        self.assertEqual(encrypt["ru_maxrss_bytes_median"], 1_000_110)

    def test_sufficient_n_gets_iqr_and_failure_marks_incomplete(self):
        values = [record(index, memory=index) for index in range(6)] + [record(6, status="failed")]
        loaded = [report.LoadedRecord("test", index + 1, value) for index, value in enumerate(values)]
        row = report.summarize(loaded, 5, "baseline")[0]
        self.assertEqual(row["statistics_mode"], "median_iqr_min_max")
        self.assertEqual(row["time_ns_iqr"], 3)
        self.assertEqual(row["incomplete"], "true")

    def test_pipeline_writes_outputs_and_reports_invalid_json(self):
        root = Path(__file__).parents[3]
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            source = temp / "input.jsonl"
            source.write_text(json.dumps(record(0)) + "\n{bad\n", encoding="utf-8")
            output = temp / "out"
            result = report.main([str(source), "--schema", str(root / "schemas" / "benchmark.schema.json"), "--output-dir", str(output)])
            self.assertEqual(result, 2)
            self.assertTrue((output / "report.csv").is_file())
            self.assertTrue((output / "memory.csv").is_file())
            self.assertTrue((output / "memory.tex").is_file())
            self.assertTrue((output / "memory.pgfplots.tex").is_file())
            issues = json.loads((output / "issues.json").read_text(encoding="utf-8"))
            self.assertEqual(issues["validation_issues"][0]["kind"], "invalid_json")


if __name__ == "__main__":
    unittest.main()
