#!/usr/bin/env python3
"""Validate benchmark JSONL and generate reproducible paper artifacts."""

from __future__ import annotations

import argparse
import csv
import json
import math
import re
import sys
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path
from statistics import median
from typing import Any, Iterable, Sequence

from jsonschema import Draft202012Validator, FormatChecker


GROUP_FIELDS = (
    "cipher",
    "language",
    "operation",
    "backend",
    "security_profile",
    "gc_mode",
    "host",
)
DELTA_MEMORY_METRICS = (
    "rss_bytes_delta",
    "heap_bytes_delta",
    "peak_rss_bytes_delta",
)
MEMORY_METRICS = DELTA_MEMORY_METRICS + ("ru_maxrss_bytes",)
METRICS = ("time_ns", "cycles") + MEMORY_METRICS
CSV_FIELDS = GROUP_FIELDS + (
    "n",
    "statistics_mode",
    "incomplete",
    "failed_records",
) + tuple(
    f"{metric}_{suffix}"
    for metric in METRICS
    for suffix in ("median", "q1", "q3", "iqr", "min", "max")
) + tuple(f"{metric}_baseline_subtracted" for metric in DELTA_MEMORY_METRICS)


@dataclass(frozen=True)
class LoadedRecord:
    source: str
    line: int
    value: dict[str, Any]


def _display(value: Any) -> str:
    return "unknown" if value is None else str(value)


def _group_key(record: dict[str, Any]) -> tuple[str, ...]:
    return (
        _display(record["benchmark"]["cipher"]),
        _display(record["runtime"]["language"]),
        _display(record["benchmark"]["operation"]),
        _display(record["benchmark"]["backend"]),
        _display(record["benchmark"]["security_profile"]),
        _display(record["runtime"]["gc_mode"]),
        json.dumps(record["host"], sort_keys=True, separators=(",", ":")),
    )


def _metric(record: dict[str, Any], name: str) -> int | None:
    measurement = record["measurement"]
    if name in ("time_ns", "cycles"):
        return measurement[name]
    return measurement["memory"].get(name)


def _quartiles(values: Sequence[int]) -> tuple[float, float]:
    """Tukey hinges: medians of the lower and upper halves."""
    ordered = sorted(values)
    midpoint = len(ordered) // 2
    return float(median(ordered[:midpoint])), float(median(ordered[-midpoint:]))


def _stats(values: Sequence[int], sufficient_n: bool) -> dict[str, int | float | None]:
    if not values:
        return {key: None for key in ("median", "q1", "q3", "iqr", "min", "max")}
    result: dict[str, int | float | None] = {
        "median": median(values),
        "q1": None,
        "q3": None,
        "iqr": None,
        "min": min(values),
        "max": max(values),
    }
    if sufficient_n:
        q1, q3 = _quartiles(values)
        result.update(q1=q1, q3=q3, iqr=q3 - q1)
    return result


def load_records(
    paths: Iterable[Path], validator: Draft202012Validator
) -> tuple[list[LoadedRecord], list[dict[str, Any]]]:
    valid: list[LoadedRecord] = []
    issues: list[dict[str, Any]] = []
    for path in paths:
        with path.open(encoding="utf-8") as stream:
            for line_number, raw_line in enumerate(stream, 1):
                if not raw_line.strip():
                    continue
                try:
                    value = json.loads(raw_line)
                except json.JSONDecodeError as error:
                    issues.append({"source": str(path), "line": line_number, "kind": "invalid_json", "message": str(error)})
                    continue
                errors = sorted(validator.iter_errors(value), key=lambda error: list(error.path))
                if errors:
                    for error in errors:
                        location = ".".join(str(part) for part in error.absolute_path) or "$"
                        issues.append({"source": str(path), "line": line_number, "kind": "schema_error", "location": location, "message": error.message})
                    continue
                valid.append(LoadedRecord(str(path), line_number, value))
    return valid, issues


def summarize(records: Sequence[LoadedRecord], low_n_threshold: int, baseline_operation: str) -> list[dict[str, Any]]:
    measured = [record.value for record in records if record.value["iteration"]["phase"] == "measured"]
    groups: dict[tuple[str, ...], list[dict[str, Any]]] = defaultdict(list)
    for record in measured:
        groups[_group_key(record)].append(record)

    # A baseline matches every grouping dimension except operation.
    baselines: dict[tuple[str, ...], dict[str, int | float | None]] = {}
    for key, group in groups.items():
        if key[2] != baseline_operation:
            continue
        baseline_key = key[:2] + key[3:]
        ok = [record for record in group if record["status"] == "ok"]
        baselines[baseline_key] = {
            metric: (median(values) if (values := [_metric(record, metric) for record in ok if _metric(record, metric) is not None]) else None)
            for metric in DELTA_MEMORY_METRICS
        }

    rows: list[dict[str, Any]] = []
    for key in sorted(groups):
        group = groups[key]
        ok = [record for record in group if record["status"] == "ok"]
        failed_count = len(group) - len(ok)
        run_indices: dict[str, list[int]] = defaultdict(list)
        for record in group:
            run_indices[record["run_id"]].append(record["iteration"]["index"])
        incomplete = bool(failed_count)
        for indices in run_indices.values():
            incomplete = incomplete or len(indices) != len(set(indices)) or (
                bool(indices) and set(indices) != set(range(max(indices) + 1))
            )
        sufficient_n = len(ok) > low_n_threshold
        row: dict[str, Any] = dict(zip(GROUP_FIELDS, key))
        row.update(n=len(ok), statistics_mode="median_iqr_min_max" if sufficient_n else "low_n_median_min_max", incomplete=str(incomplete).lower(), failed_records=failed_count)
        for metric in METRICS:
            values = [_metric(record, metric) for record in ok]
            stats = _stats([value for value in values if value is not None], sufficient_n)
            for suffix, value in stats.items():
                row[f"{metric}_{suffix}"] = value
        baseline = baselines.get(key[:2] + key[3:])
        for metric in DELTA_MEMORY_METRICS:
            raw_median = row[f"{metric}_median"]
            baseline_median = baseline.get(metric) if baseline else None
            row[f"{metric}_baseline_subtracted"] = (
                raw_median - baseline_median
                if raw_median is not None and baseline_median is not None
                else None
            )
        rows.append(row)
    return rows


def write_csv(path: Path, rows: Sequence[dict[str, Any]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as stream:
        writer = csv.DictWriter(stream, fieldnames=CSV_FIELDS, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)

def write_memory_csv(path: Path, rows: Sequence[dict[str, Any]]) -> None:
    fields = GROUP_FIELDS + ("n", "statistics_mode", "incomplete") + tuple(
        f"{metric}_{suffix}"
        for metric in MEMORY_METRICS
        for suffix in ("median", "q1", "q3", "iqr", "min", "max")
    ) + tuple(
        f"{metric}_baseline_subtracted" for metric in DELTA_MEMORY_METRICS
    )
    with path.open("w", encoding="utf-8", newline="") as stream:
        writer = csv.DictWriter(stream, fieldnames=fields, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def _latex_escape(value: Any) -> str:
    text = _display(value)
    return re.sub(r"([&%$#_{}])", r"\\\1", text).replace("~", r"\textasciitilde{}")


def _number(value: Any) -> str:
    if value is None or value == "":
        return "--"
    if isinstance(value, float) and value.is_integer():
        return str(int(value))
    return str(value)


def write_latex(path: Path, rows: Sequence[dict[str, Any]]) -> None:
    columns = ("cipher", "language", "operation", "backend", "security_profile", "gc_mode", "host", "n", "time_ns_median", "time_ns_iqr", "time_ns_min", "time_ns_max", "incomplete")
    labels = ("Cipher", "Lang.", "Operation", "Backend", "Security", "GC", "Host", "$n$", "Median ns", "IQR ns", "Min ns", "Max ns", "Incomplete")
    lines = [r"\begin{tabular}{lllllllrrrrrl}", r"\toprule", " & ".join(labels) + r" \\", r"\midrule"]
    for row in rows:
        values = [_latex_escape(_number(row[column])) for column in columns]
        lines.append(" & ".join(values) + r" \\")
    lines.extend([r"\bottomrule", r"\end{tabular}", ""])
    path.write_text("\n".join(lines), encoding="utf-8")


def write_memory_latex(path: Path, rows: Sequence[dict[str, Any]]) -> None:
    columns = (
        "cipher", "language", "operation", "backend", "n",
        "ru_maxrss_bytes_median", "rss_bytes_delta_median",
        "heap_bytes_delta_median", "peak_rss_bytes_delta_median",
    )
    labels = (
        "Cipher", "Lang.", "Operation", "Backend", "$n$",
        "Median $R_{max}$ (B)", "Median RSS $\\Delta$ (B)",
        "Median heap $\\Delta$ (B)", "Median peak RSS $\\Delta$ (B)",
    )
    lines = [r"\begin{tabular}{llllrrrrr}", r"\toprule", " & ".join(labels) + r" \\", r"\midrule"]
    for row in rows:
        values = [_latex_escape(_number(row[column])) for column in columns]
        lines.append(" & ".join(values) + r" \\")
    lines.extend([r"\bottomrule", r"\end{tabular}", ""])
    path.write_text("\n".join(lines), encoding="utf-8")


def write_memory_pgfplots(path: Path, csv_name: str) -> None:
    content = rf"""% Absolute process high-water RSS from getrusage(ru_maxrss), normalized to bytes.
% Requires \usepackage{{pgfplots}} and \pgfplotsset{{compat=1.18}}
\begin{{tikzpicture}}
  \begin{{axis}}[
    ybar,
    ylabel={{Median absolute $R_{{max}}$ (bytes)}},
    x expr=\coordindex,
    xtick=data,
    xticklabels from table={{{csv_name}}}{{operation}},
    x tick label style={{rotate=45,anchor=east}},
  ]
    \addplot table[y=ru_maxrss_bytes_median,col sep=comma] {{{csv_name}}};
  \end{{axis}}
\end{{tikzpicture}}
"""
    path.write_text(content, encoding="utf-8")

def write_pgfplots(path: Path, csv_name: str) -> None:
    content = rf"""% Requires \usepackage{{pgfplots}} and \pgfplotsset{{compat=1.18}}
% Filter report.csv for a homogeneous comparison before inclusion when needed.
\begin{{tikzpicture}}
  \begin{{axis}}[
    ybar,
    ymode=log,
    ylabel={{Median time (ns)}},
    x expr=\coordindex,
    xtick=data,
    xticklabels from table={{{csv_name}}}{{operation}},
    x tick label style={{rotate=45,anchor=east}},
  ]
    \addplot table[y=time_ns_median,col sep=comma] {{{csv_name}}};
  \end{{axis}}
\end{{tikzpicture}}
"""
    path.write_text(content, encoding="utf-8")


def write_issues(path: Path, issues: Sequence[dict[str, Any]], records: Sequence[LoadedRecord]) -> None:
    failed = [
        {"source": record.source, "line": record.line, "kind": "failed_record", "run_id": record.value["run_id"], "error": record.value["error"]}
        for record in records
        if record.value["status"] == "failed"
    ]
    path.write_text(json.dumps({"validation_issues": list(issues), "failed_records": failed}, indent=2) + "\n", encoding="utf-8")


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("inputs", nargs="+", type=Path, help="JSONL input files")
    parser.add_argument("--schema", type=Path, default=Path("schemas/benchmark.schema.json"))
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--low-n-threshold", type=int, default=5, help="suppress IQR when n is at most this value")
    parser.add_argument("--baseline-operation", default="baseline", help="operation whose median memory is subtracted from matching groups")
    return parser.parse_args(argv)


def main(argv: Sequence[str] | None = None) -> int:
    args = parse_args(argv)
    if args.low_n_threshold < 1:
        raise SystemExit("--low-n-threshold must be positive")
    schema = json.loads(args.schema.read_text(encoding="utf-8"))
    validator = Draft202012Validator(schema, format_checker=FormatChecker())
    records, issues = load_records(args.inputs, validator)
    rows = summarize(records, args.low_n_threshold, args.baseline_operation)
    args.output_dir.mkdir(parents=True, exist_ok=True)
    write_csv(args.output_dir / "report.csv", rows)
    write_latex(args.output_dir / "report.tex", rows)
    write_pgfplots(args.output_dir / "report.pgfplots.tex", "report.csv")
    write_memory_csv(args.output_dir / "memory.csv", rows)
    write_memory_latex(args.output_dir / "memory.tex", rows)
    write_memory_pgfplots(args.output_dir / "memory.pgfplots.tex", "memory.csv")
    write_issues(args.output_dir / "issues.json", issues, records)
    print(f"validated={len(records)} invalid={len(issues)} groups={len(rows)}", file=sys.stderr)
    return 2 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
