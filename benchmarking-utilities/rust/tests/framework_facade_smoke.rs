use hhe_benchmark::benchmark_with_implementation;
use serde_json::Value;
use std::cell::Cell;

#[test]
fn framework_facade_honors_cli_environment_and_emits_jsonl() {
    let output = std::env::temp_dir().join(format!(
        "hhe-rust-facade-smoke-{}.jsonl",
        std::process::id()
    ));
    let _ = std::fs::remove_file(&output);

    std::env::set_var("HHE_BENCH_CIPHER", "synthetic-cipher");
    std::env::set_var("HHE_BENCH_OPERATION", "synthetic-operation");
    std::env::set_var("HHE_BENCH_IMPLEMENTATION", "synthetic-framework");
    std::env::set_var("HHE_BENCH_ITERATIONS", "2");
    std::env::set_var("HHE_BENCH_WARMUPS", "1");
    std::env::set_var("HHE_BENCH_POLICY", "warn");
    std::env::set_var("HHE_BENCH_OUTPUT", &output);

    let calls = Cell::new(0);
    benchmark_with_implementation("fallback-operation", 99, "fallback-framework", || {
        calls.set(calls.get() + 1);
    });

    let records: Vec<Value> = std::fs::read_to_string(&output)
        .expect("read smoke JSONL")
        .lines()
        .map(|line| serde_json::from_str(line).expect("parse smoke record"))
        .collect();

    assert_eq!(calls.get(), 3);
    assert_eq!(records.len(), 3);
    assert_eq!(records[0]["iteration"]["phase"], "warmup");
    assert_eq!(records[1]["iteration"]["phase"], "measured");
    assert_eq!(records[2]["iteration"]["phase"], "measured");
    for record in records {
        assert_eq!(record["schema_version"], "1.0.0");
        assert_eq!(record["benchmark"]["cipher"], "synthetic-cipher");
        assert_eq!(record["benchmark"]["operation"], "synthetic-operation");
        assert_eq!(record["runtime"]["implementation"], "synthetic-framework");
    }

    std::fs::remove_file(output).expect("remove smoke JSONL");
}
