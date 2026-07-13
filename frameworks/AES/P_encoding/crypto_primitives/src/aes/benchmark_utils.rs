pub use hhe_benchmark::{print_header, print_message, print_status};

pub fn benchmark<F>(name: &str, iterations: usize, f: F)
where
    F: FnMut(),
{
    hhe_benchmark::benchmark_with_implementation(name, iterations, env!("CARGO_PKG_NAME"), f);
}
