# Benchmark report generator

This pipeline treats per-iteration JSONL as the source of truth. It validates every nonblank line against `schemas/benchmark.schema.json`, excludes warm-ups, groups measured records by cipher, language, operation, backend, security profile, GC mode, and the complete host identity, then writes:

- `report.csv`: machine-readable statistics and baseline-subtracted memory medians;
- `report.tex`: a booktabs-compatible paper table;
- `report.pgfplots.tex`: a timing pgfplots template backed by the CSV;
- `memory.csv`: focused RSS/heap delta and absolute `ru_maxrss_bytes` statistics;
- `memory.tex`: a booktabs-compatible memory table;
- `memory.pgfplots.tex`: an absolute high-water RSS plot template;
- `issues.json`: invalid input and valid records whose benchmark status is `failed`.

Install `requirements.txt`, then run from the repository root:

```text
python benchmarking-utilities/report/report_generator.py results/*.jsonl \
  --output-dir build/report \
  --baseline-operation baseline
```

Groups with at most five successful measured records are labeled `low_n_median_min_max`; their quartiles and IQR are blank. Change the cutoff with `--low-n-threshold`. Larger groups use Tukey-hinge quartiles. Failed records, duplicate measured indices within a run, or gaps in a run's measured indices mark a group incomplete.

Baseline subtraction matches the configured baseline operation on every other grouping dimension. The three `*_baseline_subtracted` columns are the group memory median minus the matching baseline memory median. They remain blank if either median is unavailable. Raw medians are always retained.

u_maxrss_bytes is the absolute process high-water resident set size after an iteration, normalized to bytes. It is summarized directly and is not baseline-subtracted. A zero peak_rss_bytes_delta can coexist with a nonzero u_maxrss_bytes when an earlier setup or warm-up already established the process high-water mark.

Schema-invalid lines are omitted from statistics, recorded in `issues.json`, and cause exit status 2 after artifacts are written. Valid `status: failed` records are reported and mark their group incomplete but do not contribute measurements.
