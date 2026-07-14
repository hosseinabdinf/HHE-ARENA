use std::fmt::{self, Display};

pub use std::hint::black_box;

pub struct Criterion {
    _sample_size: usize,
}

impl Default for Criterion {
    fn default() -> Self {
        Self { _sample_size: 1 }
    }
}

impl Criterion {
    pub fn sample_size(mut self, sample_size: usize) -> Self {
        self._sample_size = sample_size;
        self
    }

    pub fn bench_function<I, F>(&mut self, id: I, mut routine: F)
    where
        I: Display,
        F: FnMut(&mut Bencher),
    {
        let mut bencher = Bencher::new(id.to_string());
        routine(&mut bencher);
    }

    pub fn benchmark_group<'a>(&'a mut self, name: impl Into<String>) -> BenchmarkGroup<'a> {
        BenchmarkGroup {
            _criterion: self,
            name: name.into(),
        }
    }
}

pub struct Bencher {
    operation: String,
}

impl Bencher {
    fn new(operation: String) -> Self {
        Self { operation }
    }

    pub fn iter<R, F>(&mut self, mut routine: F)
    where
        F: FnMut() -> R,
    {
        hhe_benchmark::benchmark_with_implementation(
            &self.operation,
            1,
            "criterion-compat",
            || {
                black_box(routine());
            },
        );
    }
}

pub struct BenchmarkGroup<'a> {
    _criterion: &'a mut Criterion,
    name: String,
}

impl BenchmarkGroup<'_> {
    pub fn sample_size(&mut self, _sample_size: usize) {}

    pub fn bench_function<I, F>(&mut self, id: I, mut routine: F)
    where
        I: Display,
        F: FnMut(&mut Bencher),
    {
        let mut bencher = Bencher::new(format!("{}/{}", self.name, id));
        routine(&mut bencher);
    }

    pub fn bench_with_input<I, T, F>(&mut self, id: I, input: &T, mut routine: F)
    where
        I: Display,
        F: FnMut(&mut Bencher, &T),
    {
        let mut bencher = Bencher::new(format!("{}/{}", self.name, id));
        routine(&mut bencher, input);
    }

    pub fn finish(self) {}
}

pub struct BenchmarkId {
    function: String,
    parameter: String,
}

impl BenchmarkId {
    pub fn new(function: impl Display, parameter: impl Display) -> Self {
        Self {
            function: function.to_string(),
            parameter: parameter.to_string(),
        }
    }
}

impl Display for BenchmarkId {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(formatter, "{}/{}", self.function, self.parameter)
    }
}

#[macro_export]
macro_rules! criterion_group {
    (name = $name:ident; config = $config:expr; targets = $($target:path),+ $(,)?) => {
        pub fn $name() {
            let mut criterion = $config;
            $($target(&mut criterion);)+
        }
    };
    ($name:ident, $($target:path),+ $(,)?) => {
        pub fn $name() {
            let mut criterion = $crate::Criterion::default();
            $($target(&mut criterion);)+
        }
    };
}

#[macro_export]
macro_rules! criterion_main {
    ($($group:path),+ $(,)?) => {
        fn main() {
            $($group();)+
        }
    };
}
