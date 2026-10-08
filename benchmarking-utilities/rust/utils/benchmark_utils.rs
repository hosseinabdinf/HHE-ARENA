use serde::Serialize;
use std::fs::{self, OpenOptions};
use std::io::{self, BufWriter, Write};
use std::path::PathBuf;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Policy {
    Strict,
    Warn,
}
#[derive(Clone, Copy, Debug, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Phase {
    Warmup,
    Measured,
}

#[derive(Clone, Debug)]
pub struct BenchmarkConfig {
    pub cipher: String,
    pub operation: String,
    pub backend: Option<String>,
    pub security_profile: Option<String>,
    pub implementation: String,
    pub iterations: usize,
    pub warmups: Option<usize>,
    pub cpu: Option<usize>,
    pub policy: Policy,
    pub output: PathBuf,
}
impl BenchmarkConfig {
    pub fn new(
        cipher: impl Into<String>,
        operation: impl Into<String>,
        iterations: usize,
        output: impl Into<PathBuf>,
    ) -> Self {
        Self {
            cipher: cipher.into(),
            operation: operation.into(),
            backend: None,
            security_profile: None,
            implementation: env!("CARGO_PKG_NAME").into(),
            iterations,
            warmups: None,
            cpu: None,
            policy: Policy::Warn,
            output: output.into(),
        }
    }
}

#[derive(Clone, Debug, Serialize)]
struct Environment {
    policy: Policy,
    requested_cpu: Option<usize>,
    actual_cpu: Option<usize>,
    affinity_verified: Option<bool>,
    rdtscp_supported: Option<bool>,
    invariant_tsc: Option<bool>,
    nonstop_tsc: Option<bool>,
    governor: Option<String>,
    turbo_enabled: Option<bool>,
    checks_passed: bool,
    warnings: Vec<String>,
}
#[derive(Clone, Copy, Debug, Default, Serialize)]
struct MemoryDelta {
    rss_bytes_delta: Option<i64>,
    heap_bytes_delta: Option<i64>,
    peak_rss_bytes_delta: Option<i64>,
    ru_maxrss_bytes: Option<i64>,
}
#[derive(Clone, Copy, Debug, Default)]
struct MemorySample {
    rss: Option<i64>,
    heap: Option<i64>,
    peak: Option<i64>,
}
#[derive(Serialize)]
struct Record<'a> {
    schema_version: &'static str,
    run_id: &'a str,
    timestamp_utc: String,
    benchmark: Benchmark<'a>,
    runtime: Runtime<'a>,
    host: Host,
    environment: &'a Environment,
    iteration: Iteration,
    measurement: Measurement,
    status: &'static str,
    error: Option<String>,
}
#[derive(Serialize)]
struct Benchmark<'a> {
    cipher: &'a str,
    operation: &'a str,
    backend: &'a Option<String>,
    security_profile: &'a Option<String>,
    parameters: serde_json::Value,
}
#[derive(Serialize)]
struct Runtime<'a> {
    language: &'static str,
    implementation: &'a str,
    version: Option<&'static str>,
    gc_mode: &'static str,
    allocator: &'static str,
}
#[derive(Serialize)]
struct Host {
    hostname: Option<String>,
    os: &'static str,
    arch: &'static str,
    cpu_model: Option<String>,
    logical_cpus: Option<usize>,
}
#[derive(Serialize)]
struct Iteration {
    phase: Phase,
    index: usize,
}
#[derive(Serialize)]
struct Measurement {
    time_ns: Option<u64>,
    cycles: Option<u64>,
    memory: MemoryDelta,
}

pub fn run_benchmark<F>(config: &BenchmarkConfig, mut operation: F) -> io::Result<()>
where
    F: FnMut(),
{
    if config.iterations == 0 {
        return Err(io::Error::new(
            io::ErrorKind::InvalidInput,
            "iterations must be positive",
        ));
    }
    let environment = prepare_environment(config.cpu, config.policy)?;
    let _cycles_per_second = calibrate_once();
    let warmups = config
        .warmups
        .unwrap_or_else(|| proportional_warmups(config.iterations));
    let run_id = format!("{}-{}", std::process::id(), unix_nanos());
    let file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(&config.output)?;
    let mut writer = BufWriter::new(file);
    let mut measured_ns = Vec::with_capacity(config.iterations);
    for phase in [Phase::Warmup, Phase::Measured] {
        let count = if matches!(phase, Phase::Warmup) {
            warmups
        } else {
            config.iterations
        };
        for index in 0..count {
            let before = memory_sample();
            let start_time = Instant::now();
            let start_cycles = tsc_start();
            operation();
            let end_cycles = tsc_end();
            let time_ns = u64::try_from(start_time.elapsed().as_nanos()).ok();
            let after = memory_sample();
            if matches!(phase, Phase::Measured) {
                if let Some(ns) = time_ns {
                    measured_ns.push(ns);
                }
            }
            let record = make_record(
                config,
                &run_id,
                &environment,
                phase,
                index,
                time_ns,
                cycle_delta(start_cycles, end_cycles),
                memory_delta(before, after),
            );
            serde_json::to_writer(&mut writer, &record)
                .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;
            writer.write_all(b"\n")?;
        }
    }
    writer.flush()?;
    print_summary(&config.operation, &mut measured_ns);
    Ok(())
}

pub fn benchmark<F>(name: &str, iterations: usize, f: F)
where
    F: FnMut(),
{
    benchmark_with_implementation(name, iterations, env!("CARGO_PKG_NAME"), f)
}

pub fn benchmark_with_implementation<F>(name: &str, iterations: usize, implementation: &str, f: F)
where
    F: FnMut(),
{
    // A single process can contain several benchmark calls. When the runner
    // selects one operation, only that call may execute or emit records.
    // Running the other closures unmeasured still performs the expensive work
    // and can make a supposedly bounded HHE operation run for hours.
    if let Ok(selected_operation) = std::env::var("HHE_BENCH_OPERATION") {
        let selected_prefix = std::env::var("HHE_BENCH_OPERATION_PREFIX").unwrap_or_default();
        if selected_operation != name
            && (selected_prefix.is_empty() || !name.starts_with(&selected_prefix))
        {
            return;
        }
    }

    let iterations = std::env::var("HHE_BENCH_ITERATIONS")
        .ok()
        .and_then(|value| value.parse().ok())
        .filter(|value: &usize| *value > 0)
        .unwrap_or(iterations);
    let output = std::env::var_os("HHE_BENCH_OUTPUT")
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("benchmark.jsonl"));
    let mut config = BenchmarkConfig::new(
        std::env::var("HHE_BENCH_CIPHER").unwrap_or_else(|_| "unknown".into()),
        std::env::var("HHE_BENCH_OPERATION").unwrap_or_else(|_| name.into()),
        iterations,
        output,
    );
    config.implementation =
        std::env::var("HHE_BENCH_IMPLEMENTATION").unwrap_or_else(|_| implementation.into());
    config.backend = std::env::var("HHE_BENCH_BACKEND").ok();
    config.security_profile = std::env::var("HHE_BENCH_SECURITY_PROFILE").ok();
    config.cpu = std::env::var("HHE_BENCH_CPU")
        .ok()
        .and_then(|v| v.parse().ok());
    config.warmups = std::env::var("HHE_BENCH_WARMUPS")
        .ok()
        .and_then(|v| v.parse().ok());
    config.policy = if std::env::var("HHE_BENCH_POLICY").as_deref() == Ok("strict") {
        Policy::Strict
    } else {
        Policy::Warn
    };
    if let Err(error) = run_benchmark(&config, f) {
        eprintln!("benchmark failed: {error}");
        panic!("benchmark failed");
    }
}

fn prepare_environment(cpu: Option<usize>, policy: Policy) -> io::Result<Environment> {
    let mut warnings = Vec::new();
    let rdtscp = rdtscp_supported();
    if rdtscp != Some(true) {
        warnings.push("RDTSCP is unavailable; cycles are null".into());
    }
    let invariant = cpu_flag("constant_tsc");
    let nonstop = cpu_flag("nonstop_tsc");
    if invariant != Some(true) {
        warnings.push("invariant TSC was not verified".into());
    }
    if nonstop != Some(true) {
        warnings.push("nonstop TSC was not verified".into());
    }

    let governor = fs::read_to_string("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
        .ok()
        .map(|v| v.trim().into());
    if governor.as_deref() != Some("performance") {
        warnings.push("CPU governor is not verified as performance".into());
    }
    let turbo_enabled = turbo_enabled();
    if turbo_enabled != Some(false) {
        warnings.push("Turbo is enabled or unavailable".into());
    }
    let actual_cpu = if let Some(target) = cpu {
        pin_cpu(target)?;
        current_cpu()
    } else {
        current_cpu()
    };
    let affinity_verified = cpu.map(|target| actual_cpu == Some(target));
    if affinity_verified == Some(false) {
        warnings.push("CPU affinity verification failed".into());
    }
    let checks_passed = warnings.is_empty();
    if matches!(policy, Policy::Strict) && !checks_passed {
        return Err(io::Error::new(io::ErrorKind::Other, warnings.join("; ")));
    }
    Ok(Environment {
        policy,
        requested_cpu: cpu,
        actual_cpu,
        affinity_verified,
        rdtscp_supported: rdtscp,
        invariant_tsc: invariant,
        nonstop_tsc: nonstop,
        governor,
        turbo_enabled,
        checks_passed,
        warnings,
    })
}

#[cfg(target_os = "linux")]
fn pin_cpu(cpu: usize) -> io::Result<()> {
    unsafe {
        let mut set: libc::cpu_set_t = std::mem::zeroed();
        libc::CPU_ZERO(&mut set);
        libc::CPU_SET(cpu, &mut set);
        if libc::sched_setaffinity(0, std::mem::size_of::<libc::cpu_set_t>(), &set) != 0 {
            return Err(io::Error::last_os_error());
        }
    }
    Ok(())
}
#[cfg(not(target_os = "linux"))]
fn pin_cpu(_: usize) -> io::Result<()> {
    Err(io::Error::new(
        io::ErrorKind::Unsupported,
        "CPU pinning requires Linux",
    ))
}
#[cfg(target_os = "linux")]
fn current_cpu() -> Option<usize> {
    let cpu = unsafe { libc::sched_getcpu() };
    (cpu >= 0).then_some(cpu as usize)
}
#[cfg(not(target_os = "linux"))]
fn current_cpu() -> Option<usize> {
    None
}

fn calibrate_once() -> Option<f64> {
    static VALUE: std::sync::OnceLock<Option<f64>> = std::sync::OnceLock::new();
    *VALUE.get_or_init(|| {
        let start_time = Instant::now();
        let start = tsc_start()?;
        while start_time.elapsed() < Duration::from_millis(100) {
            std::hint::spin_loop();
        }
        let end = tsc_end()?;
        Some((end - start) as f64 / start_time.elapsed().as_secs_f64())
    })
}
fn proportional_warmups(iterations: usize) -> usize {
    ((iterations + 9) / 10).clamp(1, 10)
}

#[cfg(target_arch = "x86_64")]
fn rdtscp_supported() -> Option<bool> {
    unsafe {
        let max = std::arch::x86_64::__cpuid(0x8000_0000).eax;
        Some(max >= 0x8000_0001 && (std::arch::x86_64::__cpuid(0x8000_0001).edx & (1 << 27)) != 0)
    }
}
#[cfg(not(target_arch = "x86_64"))]
fn rdtscp_supported() -> Option<bool> {
    None
}
#[cfg(target_arch = "x86_64")]
fn tsc_start() -> Option<u64> {
    unsafe {
        std::arch::x86_64::__cpuid(0);
        Some(std::arch::x86_64::_rdtsc())
    }
}
#[cfg(not(target_arch = "x86_64"))]
fn tsc_start() -> Option<u64> {
    None
}
#[cfg(target_arch = "x86_64")]
fn tsc_end() -> Option<u64> {
    if rdtscp_supported() != Some(true) {
        return None;
    }
    unsafe {
        let mut aux = 0;
        let value = std::arch::x86_64::__rdtscp(&mut aux);
        std::arch::x86_64::__cpuid(0);
        Some(value)
    }
}
#[cfg(not(target_arch = "x86_64"))]
fn tsc_end() -> Option<u64> {
    None
}
fn cycle_delta(start: Option<u64>, end: Option<u64>) -> Option<u64> {
    end?.checked_sub(start?)
}

fn memory_sample() -> MemorySample {
    #[cfg(target_os = "linux")]
    unsafe {
        let rss = fs::read_to_string("/proc/self/statm")
            .ok()
            .and_then(|s| s.split_whitespace().nth(1)?.parse::<i64>().ok())
            .map(|pages| pages * libc::sysconf(libc::_SC_PAGESIZE));
        let heap = Some(i64::try_from(libc::mallinfo2().uordblks).ok()).flatten();
        let mut usage: libc::rusage = std::mem::zeroed();
        let peak = (libc::getrusage(libc::RUSAGE_SELF, &mut usage) == 0)
            .then_some(usage.ru_maxrss as i64 * 1024);
        MemorySample { rss, heap, peak }
    }
    #[cfg(not(target_os = "linux"))]
    {
        MemorySample::default()
    }
}
fn memory_delta(a: MemorySample, b: MemorySample) -> MemoryDelta {
    MemoryDelta {
        rss_bytes_delta: zip_delta(a.rss, b.rss),
        heap_bytes_delta: zip_delta(a.heap, b.heap),
        peak_rss_bytes_delta: zip_delta(a.peak, b.peak),
        ru_maxrss_bytes: b.peak,
    }
}
fn zip_delta(a: Option<i64>, b: Option<i64>) -> Option<i64> {
    Some(b?.checked_sub(a?)?)
}

fn make_record<'a>(
    c: &'a BenchmarkConfig,
    run_id: &'a str,
    environment: &'a Environment,
    phase: Phase,
    index: usize,
    time_ns: Option<u64>,
    cycles: Option<u64>,
    memory: MemoryDelta,
) -> Record<'a> {
    Record {
        schema_version: "1.0.0",
        run_id,
        timestamp_utc: format_timestamp(),
        benchmark: Benchmark {
            cipher: &c.cipher,
            operation: &c.operation,
            backend: &c.backend,
            security_profile: &c.security_profile,
            parameters: serde_json::json!({}),
        },
        runtime: Runtime {
            language: "rust",
            implementation: &c.implementation,
            version: option_env!("CARGO_PKG_VERSION"),
            gc_mode: "not_applicable",
            allocator: "system_glibc",
        },
        host: Host {
            hostname: benchmark_hostname(),
            os: std::env::consts::OS,
            arch: std::env::consts::ARCH,
            cpu_model: cpu_model(),
            logical_cpus: std::thread::available_parallelism().ok().map(usize::from),
        },
        environment,
        iteration: Iteration { phase, index },
        measurement: Measurement {
            time_ns,
            cycles,
            memory,
        },
        status: "ok",
        error: None,
    }
}
fn benchmark_hostname() -> Option<String> {
    ["HHE_BENCH_HOSTNAME", "HOSTNAME"].iter().find_map(|name| {
        std::env::var(name)
            .ok()
            .map(|value| value.trim().to_owned())
            .filter(|value| !value.is_empty())
    })
}
fn cpu_flag(flag: &str) -> Option<bool> {
    let text = fs::read_to_string("/proc/cpuinfo").ok()?;
    Some(
        text.lines()
            .filter(|l| l.starts_with("flags"))
            .any(|l| l.split_whitespace().any(|v| v == flag)),
    )
}
fn cpu_model() -> Option<String> {
    fs::read_to_string("/proc/cpuinfo")
        .ok()?
        .lines()
        .find_map(|l| {
            l.strip_prefix("model name")
                .and_then(|v| v.split_once(':'))
                .map(|(_, v)| v.trim().into())
        })
}
fn turbo_enabled() -> Option<bool> {
    fs::read_to_string("/sys/devices/system/cpu/intel_pstate/no_turbo")
        .ok()
        .and_then(|v| match v.trim() {
            "0" => Some(true),
            "1" => Some(false),
            _ => None,
        })
}
fn unix_nanos() -> u128 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_nanos()
}
fn format_timestamp() -> String {
    use time::format_description::well_known::Rfc3339;
    time::OffsetDateTime::now_utc()
        .format(&Rfc3339)
        .unwrap_or_else(|_| "1970-01-01T00:00:00Z".into())
}

fn print_summary(name: &str, values: &mut [u64]) {
    if values.is_empty() {
        return;
    }
    values.sort_unstable();
    let median = values[values.len() / 2];
    println!(
        "-> {name}: n={}, median={} ns, min={} ns, max={} ns",
        values.len(),
        median,
        values[0],
        values[values.len() - 1]
    );
}

pub fn print_header(header: &str) {
    println!("\n ------------------===|\t {header} \t|===------------------")
}
pub fn print_status(message: &str) {
    print!("--> {message}");
    let _ = io::stdout().flush();
}
pub fn print_message(message: &str) {
    println!("--> {message}")
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn warmups_are_proportional() {
        assert_eq!(proportional_warmups(1), 1);
        assert_eq!(proportional_warmups(100), 10);
        assert_eq!(proportional_warmups(1000), 10);
    }
    #[test]
    fn signed_deltas_are_preserved() {
        assert_eq!(zip_delta(Some(10), Some(3)), Some(-7));
    }
    #[test]
    fn memory_delta_retains_absolute_ru_maxrss() {
        let before = MemorySample {
            peak: Some(100),
            ..MemorySample::default()
        };
        let after = MemorySample {
            peak: Some(150),
            ..MemorySample::default()
        };
        let memory = memory_delta(before, after);
        assert_eq!(memory.peak_rss_bytes_delta, Some(50));
        assert_eq!(memory.ru_maxrss_bytes, Some(150));
    }
}
