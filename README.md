# Performance Benchmark: Legacy vs. Next-Gen Math Engine

A side-by-side architectural benchmark executing identical slot game math loops (10,000,000 simulated spins) on identical hardware. By eliminating dynamic array allocations, reducing runtime state evaluation overhead, and optimizing the execution path, throughput increased **65x**.

```text
Legacy Engine:     4,641 spins/sec  ████░░░░░░░░░░░░░░░░
Next-Gen Engine: 304,853 spins/sec  ████████████████████
                                    🚀 65.6x Performance Gain
```

<p float="left">
  <img src="https://raw.githubusercontent.com/BonBonSlick/php_game_engine_guide/refs/heads/main/legacy_3x3.png" width="45%" />
  <img src="https://raw.githubusercontent.com/BonBonSlick/php_game_engine_guide/refs/heads/main/new_3x3.png" width="45%" />
</p>


## Maximal throughput I could get on my laptop using described principles and basics

To find the absolute upper limit of these optimization techniques, I isolated the core simulation logic into a standalone benchmark script: [`test_your_might.php`](https://github.com/BonBonSlick/php_game_engine_guide/blob/main/test_your_might.php). By completely stripping away framework overhead, network latency, and I/O bottlenecks, this script captures the raw mathematical throughput of our 1D memory layout and unrolled evaluation loops running on a single CPU core.

The results under pure isolation exceeded all expectations, demonstrating what the PHP 8.5 JIT compiler is truly capable of when execution paths become fully predictable:

> ### ⚡ Peak Isolation Benchmark Results
> * **Execution Time:** 0.5377 seconds
> * **Calculated RTP:** 0.1543124
> * **Max Throughput:** **1,859,739 RPS** (Spins per second)

### Performance Evolution Breakdown

| Stage / Engine Level | Throughput (RPS) | Speedup vs Legacy | Latency / 10M Spins |
| :--- | :--- | :--- | :--- |
| **Legacy Engine Baseline** | ~5,000 RPS | 1x (Baseline) | ~555 hours |
| **Optimized Core Architecture** | ~304,853 RPS | ~60x | ~32 seconds |
| **Isolated Hot Path (Standalone)** | **~1,859,739 RPS** | **~370x** | **~5.3 seconds** |

This confirms that the mathematical slot core is no longer a bottleneck. The CPU-bound simulation layer hits its theoretical "speed-of-light" limits when flattened into 1D arrays and freed from memory allocation cycles.

<p >
  <img src="https://raw.githubusercontent.com/BonBonSlick/php_game_engine_guide/refs/heads/main/Screenshot%20from%202026-06-05%2013-23-23.png" width="45%" />
</p>


