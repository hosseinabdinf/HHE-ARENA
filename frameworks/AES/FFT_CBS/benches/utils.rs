pub fn benchmark<F>(name: &str, f: F)
where
    F: FnMut(),
{
    hhe_benchmark::benchmark_with_implementation(name, 1, env!("CARGO_PKG_NAME"), f);
}
