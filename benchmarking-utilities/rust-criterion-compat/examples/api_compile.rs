use criterion::{black_box, criterion_group, criterion_main, BenchmarkId, Criterion};

fn api_surface(criterion: &mut Criterion) {
    criterion.bench_function("direct", |bencher| bencher.iter(|| black_box(1_u64)));

    let input = 7_u64;
    let mut group = criterion.benchmark_group("group");
    group.sample_size(10);
    group.bench_function(BenchmarkId::new("function", "parameter"), |bencher| {
        bencher.iter(|| black_box(input))
    });
    group.bench_with_input(
        BenchmarkId::new("input", input),
        &input,
        |bencher, value| bencher.iter(|| black_box(*value)),
    );
    group.finish();
}

criterion_group!(
    name = benches;
    config = Criterion::default().sample_size(100);
    targets = api_surface
);
criterion_main!(benches);
