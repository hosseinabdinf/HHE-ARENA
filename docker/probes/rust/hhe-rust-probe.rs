use hhe_benchmark::Policy;

fn main() {
    println!("hhe-rust-probe rust={} policy={:?}", env!("CARGO_PKG_VERSION"), Policy::Warn);
}
