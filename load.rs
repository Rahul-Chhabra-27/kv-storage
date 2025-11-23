#[cfg(not(target_env = "msvc"))]
use tikv_jemallocator::Jemalloc;
#[cfg(not(target_env = "msvc"))]
#[global_allocator]
static GLOBAL: Jemalloc = Jemalloc;

use std::time::{Duration, Instant};
use tokio::task::JoinSet;
use rand::Rng;

#[derive(Debug, Clone, Copy)]
enum WorkloadType {
    PutAll,
    GetAll,
    GetPopular,
    GetPut,
    Stress,
}

impl WorkloadType {
    fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "putall" | "put-all" | "put_all" => Some(WorkloadType::PutAll),
            "getall" | "get-all" | "get_all" => Some(WorkloadType::GetAll),
            "getpopular" | "get-popular" | "get_popular" => Some(WorkloadType::GetPopular),
            "getput" | "get-put" | "get_put" | "mixed" => Some(WorkloadType::GetPut),
            "stress" => Some(WorkloadType::Stress),
            _ => None,
        }
    }
}

#[derive(Debug, Clone)]
struct Stats {
    successful_requests: u64,
    failed_requests: u64,
    total_latency_ms: u64,
}

impl Stats {
    fn new() -> Self {
        Self {
            successful_requests: 0,
            failed_requests: 0,
            total_latency_ms: 0,
        }
    }

    fn merge(&mut self, other: &Stats) {
        self.successful_requests += other.successful_requests;
        self.failed_requests += other.failed_requests;
        self.total_latency_ms += other.total_latency_ms;
    }

    fn avg_latency_ms(&self) -> f64 {
        if self.successful_requests == 0 {
            0.0
        } else {
            self.total_latency_ms as f64 / self.successful_requests as f64
        }
    }
}

async fn do_set(
    client: &reqwest::Client,
    base_url: &str,
    key: String,
    value: String,
    stats: &mut Stats,
) {
    let start = Instant::now();
    let res = client
        .post(format!("{}/api/kv", base_url))
        .json(&sonic_rs::json!({ "key": key, "value": value }))
        .send()
        .await;

    match res {
        Ok(r) if r.status().is_success() => {
            stats.successful_requests += 1;
            stats.total_latency_ms += start.elapsed().as_millis() as u64;
        }
        _ => stats.failed_requests += 1,
    }
}

async fn do_get(
    client: &reqwest::Client,
    base_url: &str,
    key: &str,
    stats: &mut Stats,
) {
    let start = Instant::now();
    let res = client
        .get(format!("{}/api/kv/{}", base_url, key))
        .send()
        .await;

    match res {
        Ok(_) => {
            stats.successful_requests += 1;
            stats.total_latency_ms += start.elapsed().as_millis() as u64;
        }
        _ => stats.failed_requests += 1,
    }
}

async fn do_delete(
    client: &reqwest::Client,
    base_url: &str,
    key: String,
    stats: &mut Stats,
) {
    let start = Instant::now();
    let res = client
        .delete(format!("{}/api/kv/{}", base_url, key))
        .send()
        .await;

    match res {
        Ok(_) => {
            stats.successful_requests += 1;
            stats.total_latency_ms += start.elapsed().as_millis() as u64;
        }
        _ => stats.failed_requests += 1,
    }
}

// PUT-ALL workload
async fn run_worker_putall(id: usize, base_url: String, duration: Duration) -> Stats {
    let client = reqwest::Client::new();
    let mut stats = Stats::new();
    let start = Instant::now();
    let mut counter = 0u64;

    while start.elapsed() < duration {
        let key = format!("key_put_{}_{}", id, counter);
        let value = format!("v{}", counter);

        do_set(&client, &base_url, key.clone(), value, &mut stats).await;
        do_delete(&client, &base_url, key, &mut stats).await;

        counter += 1;
    }
    stats
}

// GET-ALL workload
async fn run_worker_getall(id: usize, base_url: String, duration: Duration) -> Stats {
    let client = reqwest::Client::new();
    let mut stats = Stats::new();
    let start = Instant::now();
    let mut counter = 0u64;

    while start.elapsed() < duration {
        let key = format!("unique_{}_{}", id, counter);
        do_get(&client, &base_url, &key, &mut stats).await;
        counter += 1;
    }
    stats
}

// GET-POPULAR workload
async fn run_worker_getpopular(id: usize, base_url: String, duration: Duration) -> Stats {
    let client = reqwest::Client::new();
    let mut stats = Stats::new();
    let start = Instant::now();
    let keys = (1..=10)
        .map(|i| format!("popular_{}", i))
        .collect::<Vec<_>>();

    if id == 0 {
        for k in &keys {
            let _ = client
                .post(format!("{}/api/kv", base_url))
                .json(&sonic_rs::json!({"key":k,"value":"val"}))
                .send()
                .await;
        }
    }
    tokio::time::sleep(Duration::from_millis(100)).await;

    while start.elapsed() < duration {
        let key = &keys[rand::random::<usize>() % keys.len()];
        do_get(&client, &base_url, key, &mut stats).await;
    }
    stats
}

// GET+PUT workload
async fn run_worker_getput(id: usize, base_url: String, duration: Duration) -> Stats {
    let client = reqwest::Client::new();
    let mut stats = Stats::new();
    let start = Instant::now();
    let mut counter = 0u64;

    while start.elapsed() < duration {
        let r = rand::random::<u32>() % 100;

        if r < 70 {
            let key = format!("hot_{}", rand::random::<u32>() % 20);
            do_get(&client, &base_url, &key, &mut stats).await;
        } else if r < 90 {
            let key = format!("key_gp_{}_{}", id, counter);
            let val = format!("v{}", counter);
            do_set(&client, &base_url, key, val, &mut stats).await;
        } else {
            let key = format!("key_gp_{}_{}", id, rand::random::<u32>() % 100);
            do_delete(&client, &base_url, key, &mut stats).await;
        }
        counter += 1;
    }
    stats
}

// STRESS workload
async fn run_worker_stress(id: usize, base_url: String, duration: Duration) -> Stats {
    let client = reqwest::Client::new();
    let mut stats = Stats::new();
    let start = Instant::now();
    let mut counter = 0u64;

    while start.elapsed() < duration {
        let op = rand::random::<u32>() % 100;
        if op < 60 {
            let key = format!("stress_hot_{}", rand::random::<u32>() % 100);
            do_get(&client, &base_url, &key, &mut stats).await;
        } else if op < 85 {
            let key = format!("sk_{}_{}", id, counter);
            let val = format!("v{}", counter);
            do_set(&client, &base_url, key, val, &mut stats).await;
        } else {
            let key = format!("sk_{}_{}", id, rand::random::<u32>() % 1000);
            do_delete(&client, &base_url, key, &mut stats).await;
        }
        counter += 1;
    }
    stats
}

async fn run_load_test(base_url: &str, workers: usize, duration: u64, kind: WorkloadType) {
    let duration = Duration::from_secs(duration);
    let mut tasks = JoinSet::new();
    let start = Instant::now();

    for id in 0..workers {
        let b = base_url.to_string();
        match kind {
            WorkloadType::PutAll => tasks.spawn(async move { run_worker_putall(id, b, duration).await }),
            WorkloadType::GetAll => tasks.spawn(async move { run_worker_getall(id, b, duration).await }),
            WorkloadType::GetPopular => tasks.spawn(async move { run_worker_getpopular(id, b, duration).await }),
            WorkloadType::GetPut => tasks.spawn(async move { run_worker_getput(id, b, duration).await }),
            WorkloadType::Stress => tasks.spawn(async move { run_worker_stress(id, b, duration).await }),
        };
    }

    let mut total = Stats::new();
    while let Some(res) = tasks.join_next().await {
        if let Ok(s) = res {
            total.merge(&s);
        }
    }

    let t = start.elapsed().as_secs_f64();
    let r = total.successful_requests + total.failed_requests;

    println!("Successful: {}", total.successful_requests);
    println!("Failed: {}", total.failed_requests);
    println!("Total: {}", r);
    println!("Throughput: {:.2} req/s", r as f64 / t);
    println!("Avg latency: {:.2} ms", total.avg_latency_ms());
}

#[tokio::main]
async fn main() {
    let args: Vec<String> = std::env::args().collect();

    let base = args.get(1).map(|s| s.as_str()).unwrap_or("http://localhost:4000");
    let workers = args.get(2).and_then(|s| s.parse().ok()).unwrap_or(10);
    let duration = args.get(3).and_then(|s| s.parse().ok()).unwrap_or(30);
    let kind = args.get(4).and_then(|s| WorkloadType::from_str(s)).unwrap_or(WorkloadType::GetPut);

    let client = reqwest::Client::new();
    match client.get(format!("{}/api/kv/test", base)).send().await {
        Ok(_) => {}
        Err(_) => {
            println!("Server not reachable");
            return;
        }
    }

    run_load_test(base, workers, duration, kind).await;
}
