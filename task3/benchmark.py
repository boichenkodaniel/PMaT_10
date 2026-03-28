import asyncio
import aiohttp
import time
from dataclasses import dataclass
from typing import List, Optional
import statistics

@dataclass
class BenchmarkResult:
    requests: int
    concurrency: int
    total_time: float
    succeeded: int
    failed: int
    latencies: List[float]

    @property
    def rps(self) -> float:
        return self.requests / self.total_time

    @property
    def avg_latency(self) -> float:
        if not self.latencies:
            return 0.0
        return statistics.mean(self.latencies) * 1000

    @property
    def p50_latency(self) -> float:
        if not self.latencies:
            return 0.0
        return statistics.median(self.latencies) * 1000

    @property
    def p90_latency(self) -> float:
        if not self.latencies:
            return 0.0
        sorted_lat = sorted(self.latencies)
        idx = int(len(sorted_lat) * 0.9)
        return sorted_lat[min(idx, len(sorted_lat)-1)] * 1000

    @property
    def p99_latency(self) -> float:
        if not self.latencies:
            return 0.0
        sorted_lat = sorted(self.latencies)
        idx = int(len(sorted_lat) * 0.99)
        return sorted_lat[min(idx, len(sorted_lat)-1)] * 1000

    def __str__(self) -> str:
        return f"""
  Requests:              {self.requests}
  Succeeded:             {self.succeeded}
  Failed:                {self.failed}
  Total time:            {self.total_time:.2f}s
  Requests/sec:          {self.rps:.2f}

  Latency (ms):
    avg:                 {self.avg_latency:.2f}
    p50:                 {self.p50_latency:.2f}
    p90:                 {self.p90_latency:.2f}
    p99:                 {self.p99_latency:.2f}
"""


async def make_request(session: aiohttp.ClientSession, url: str, method: str = "GET", json_data: dict = None) -> Optional[float]:
    start = time.perf_counter()
    try:
        if method == "GET":
            async with session.get(url) as resp:
                await resp.text()
        elif method == "POST":
            async with session.post(url, json=json_data) as resp:
                await resp.text()
        elif method == "PUT":
            async with session.put(url, json=json_data) as resp:
                await resp.text()
        elif method == "DELETE":
            async with session.delete(url) as resp:
                await resp.text()
        return time.perf_counter() - start
    except Exception:
        return None


async def benchmark(
    url: str,
    requests: int = 1000,
    concurrency: int = 100,
    method: str = "GET",
    json_data: dict = None
) -> BenchmarkResult:
    latencies = []
    succeeded = 0
    failed = 0

    semaphore = asyncio.Semaphore(concurrency)

    async def worker():
        nonlocal succeeded, failed
        async with aiohttp.ClientSession() as session:
            for _ in range(requests // concurrency + 1):
                async with semaphore:
                    lat = await make_request(session, url, method, json_data)
                    if lat is not None:
                        latencies.append(lat)
                        succeeded += 1
                    else:
                        failed += 1

    start_time = time.perf_counter()
    await worker()
    total_time = time.perf_counter() - start_time

    while len(latencies) > requests:
        latencies.pop()
        succeeded -= 1

    return BenchmarkResult(
        requests=requests,
        concurrency=concurrency,
        total_time=total_time,
        succeeded=succeeded,
        failed=failed,
        latencies=latencies
    )


async def main():
    print("=" * 60)
    print("BENCHMARK: FastAPI (port 8000) vs Gin (port 8080)")
    print("=" * 60)

    configs = [
        {"name": "GET /health", "url_fastapi": "http://localhost:8000/health", "url_gin": "http://localhost:8080/health", "method": "GET"},
        {"name": "POST /echo", "url_fastapi": "http://localhost:8000/echo", "url_gin": "http://localhost:8080/echo", "method": "POST", "json": {"message": "Hello World"}},
        {"name": "GET /items", "url_fastapi": "http://localhost:8000/items", "url_gin": "http://localhost:8080/items", "method": "GET"},
    ]

    requests_count = 1000
    concurrency = 50

    results = []

    for config in configs:
        print(f"\n{'='*60}")
        print(f"Test: {config['name']}")
        print(f"Requests: {requests_count}, Concurrency: {concurrency}")
        print("=" * 60)

        print("\n[FastAPI - port 8000]")
        fastapi_result = await benchmark(
            config["url_fastapi"],
            requests=requests_count,
            concurrency=concurrency,
            method=config.get("method", "GET"),
            json_data=config.get("json")
        )
        print(fastapi_result)

        print("[Gin - port 8080]")
        gin_result = await benchmark(
            config["url_gin"],
            requests=requests_count,
            concurrency=concurrency,
            method=config.get("method", "GET"),
            json_data=config.get("json")
        )
        print(gin_result)

        speedup = gin_result.rps / fastapi_result.rps if fastapi_result.rps > 0 else 0
        print(f"\n>>> Gin is {speedup:.2f}x faster than FastAPI for this endpoint")

        results.append({
            "name": config["name"],
            "fastapi": fastapi_result,
            "gin": gin_result,
            "speedup": speedup
        })

    print("\n" + "=" * 60)
    print("SUMMARY TABLE")
    print("=" * 60)
    print(f"{'Endpoint':<20} {'FastAPI RPS':>12} {'Gin RPS':>12} {'Speedup':>10}")
    print("-" * 60)
    for r in results:
        print(f"{r['name']:<20} {r['fastapi'].rps:>12.2f} {r['gin'].rps:>12.2f} {r['speedup']:>10.2f}x")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
